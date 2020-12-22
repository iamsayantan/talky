<template>
  <v-layout
    column
  >
    <v-alert
      v-if="error.isError"
      text
      prominent
      type="error"
      icon="mdi-alert-circle"
    >
      {{ error.errorMessage }}
    </v-alert>
    <v-row
      v-else
      no-gutters
    >
      <v-col cols="6">
        Local Video
        <video class="local-video" autoplay playsinline muted controls ref="localVideo"></video>
      </v-col>
      <v-col v-for="room_member in room_members" :key="room_member.user_details.id" cols="6">
        {{ room_member.user_details.username }}
        <video :id="`remote-video-${room_member.user_details.id}`" :ref="`remoteVideo-${room_member.user_details.id}`" class="local-video" autoplay playsinline></video>
      </v-col>

      <v-btn
        absolute
        dark
        fab
        bottom
        :style="{left: '50%', bottom: '40px', transform:'translateX(-50%)'}"
        color="red"
        @click="hangup"
      >
        <v-icon>mdi-phone-hangup</v-icon>
      </v-btn>
    </v-row>
  </v-layout>
</template>

<script>
  const ROOM_ID_LENGTH = 32;

  export default {
    name: "index",
    asyncData({ app, redirect, route }) {
      if (!app.$auth.loggedIn) {
        return redirect('/login', {return_to: route.path})
      }
    },
    data() {
      return {
        error: {
          isError: false,
          errorMessage: 'Something is not right. Please try again with correct URL.'
        },
        room_types: {
          a: 'AUDIO',
          av: 'AUDIO_VIDEO'
        },

        localStream: null,
        room_id: null,
        room_type: null,
        room_members: {}
      }
    },
    async mounted() {
      const { room_type, room_id } = this.$route.params;
      try {
        this.validateRoomDetails(room_id, room_type);
      } catch (e) {
        console.error('validateRoomDetails', e);
        this.error.isError = true;
        this.error.errorMessage = e.messages;
        return;
      }

      this.room_id = room_id;
      this.room_type = this.room_types[room_type];

      window.onunload = () => {
        this.hangupIfConnected();
      };

      // everything looks okay, its time to initiate the websocket connection to the server.
      await this.$Signalling.open(this.$auth.getToken('local'));
      this.$Signalling.registerOnSignallingMessageHandler(this.signallingHandler.bind(this));
      this.initiate();
    },

    beforeDestroy() {
      this.hangupIfConnected();
    },

    methods: {
      async initiate() {
        const { room_type, room_id } = this.$route.params;

        try {
          await this.initiateLocalVideo();
          this.createOrJoinRoom(room_type, room_id);
        } catch (e) {
          this.error.isError = true
          this.error.errorMessage = e.message
        }
      },

      /**
       * @param roomID
       * @param roomType
       * @returns {Error}
       */
      validateRoomDetails(roomID, roomType) {
        if (!roomID || !roomType) {
          return new Error('Invalid room details');
        }

        if (roomType !== 'a' && roomType !== 'av') {
          return new Error('Invalid room details');
        }

        if (roomID.length !== ROOM_ID_LENGTH) {
          return new Error('Invalid room details');
        }
      },

      /**
       * @param roomType
       * @param roomId
       */
      createOrJoinRoom(roomType, roomId) {
        const payload = {
          room_type: roomType === 'av' ? 'AUDIO_VIDEO' : 'AUDIO',
          room_id: roomId
        };

        this.$Signalling.send('CREATE_OR_JOIN', payload)
      },

      hangupIfConnected() {
        console.log('[hangupIfConnected] Removing user');
        this.hangup()
      },

      /**
       * Handles signalling messages from the server side.
       * @param data
       */
      signallingHandler(data) {
        data = JSON.parse(data);
        switch (data.type) {
          // ROOM_JOIN event is triggered when someone joins a room in the server side. This event occurs as a result
          // of CREATE_OR_JOIN event from the client side. The joined user could be the currently authenticated user
          // or some other peer who joined the room.
          case 'ROOM_JOIN':
            this.handleRoomJoin(data.payload);
            break;
          case 'OFFER':
            this.handleRemoteOffer(data.payload);
            break;
          case 'ANSWER':
            this.handleRemoteAnswer(data.payload);
            break;
          case 'ICE_CANDIDATE':
            this.handleNewICECandidate(data.payload);
            break;
          case 'HANGUP':
            this.handleHangup(data.payload);
            break;
          case 'error':
            this.error.isError = true;
            this.error.errorMessage = data.payload;
            return;
        }
      },

      /**
       * If the joined user is the currently authenticated user, then we start this user's local video stream and
       * wait for other people to join. For any remote peer join, we create a peer connection and send him an offer.
       *
       * @param room_id
       * @param user
       * @returns {Promise<void>}
       */
      async handleRoomJoin({ room_id, user }) {
        console.log('[handleRoomJoin] Handling room join.');
        if (room_id !== this.room_id) {
          console.log(`Mismatching room ID. Current ${this.room_id} Incoming: ${room_id}`);
          return;
        }

        if (this.room_members[user.id]) {
          console.log('Member already inside room.', this.room_members);
          return;
        }

        let messageString = null;
        if (user.id === this.$auth.user.id) {
          console.log('[handleRoomJoin] Event received for currently authenticated user.', user);
          messageString = 'You joined the room';
        } else {
          console.log('[handleRoomJoin] Event received for new user.', user);
          messageString = `${user.username} joined the room`;
          this.processUserJoin(user);
        }

        this.$toast.success(messageString);
      },

      processUserJoin(user) {
        if (this.room_members[user.id]) {
          console.error('[processUserJoin] User already a member of the room', user.username);
          return;
        }

        console.log('[processUserJoin] Creating RTCPeerConnection for user', user.username);
        const peerConnection = this.createPeerConnection(user);

        console.log('[processUserJoin] Adding local stream to RTCPeerConnection for user', user.username);
        this.localStream.getTracks().forEach(track => peerConnection.addTrack(track, this.localStream));

        const mediaStream = new MediaStream();
        this.$set(this.room_members, user.id, {
          user_details: user,
          peer_connection: peerConnection,
          media_stream: mediaStream
        });
      },

      createPeerConnection(user) {
        const peerConnection = new RTCPeerConnection({
          iceServers: [
            {
              urls: [ "stun:bn-turn1.xirsys.com" ]
            },
            {
              urls: [
                "turn:bn-turn1.xirsys.com:80?transport=udp",
                "turn:bn-turn1.xirsys.com:3478?transport=udp",
              ],
              username: "1OefJQNHAvPkVyolvfv1j5pqvgq5OilWKVfp31Yc4AeJQDbKYng-qUFE_WX_vZqxAAAAAF_W7kNpYW1zYXlhbnRhbg==",
              credential: "666c190e-3dc7-11eb-8899-0242ac140004",
            }
          ]
        });

        peerConnection.ontrack = event => {
          console.log('[ontrack]', event);
          document.getElementById(`remote-video-${user.id}`).srcObject = event.streams[0]
        };

        peerConnection.onicecandidate = ({candidate}) => {
          console.log('[onicecandidate]', candidate);
          if (candidate) {
            const payload = {
              room_id: this.room_id,
              user: this.$auth.user,
              target_user_id: user.id,
              candidate: candidate
            };

            this.$Signalling.send('ICE_CANDIDATE', payload)
          }
        };

        peerConnection.onnegotiationneeded = async (evt) => {
          console.log('[onnegotiationneeded]', evt, user);
          if (!peerConnection.remoteDescription) {
            console.log('[onnegotiationneeded] Remote description not set. Generating offer.');
            try {
              const offer = await peerConnection.createOffer({
                offerToReceiveAudio: true,
                offerToReceiveVideo: true,
                iceRestart: true
              });
              await peerConnection.setLocalDescription(offer);
              console.log('[onnegotiationneeded] Local description set.', peerConnection.localDescription);
              const payload = {
                room_id: this.room_id,
                user: this.$auth.user,
                target_user_id: user.id,
                sdp: peerConnection.localDescription
              };

              this.$Signalling.send('OFFER', payload)
            } catch (e) {
              console.error('[onnegotiationneeded] Error', e)
            }
          }
        };

        peerConnection.onremovetrack = (evt) => {
          console.log('onremovetrack', evt);
          const stream = this.$refs.remoteVideo.srcObject;
          if (stream.getTracks().length === 0) {
            this.$refs.remoteVideo.srcObject.removeAttribute('src');
            this.$refs.remoteVideo.srcObject.removeAttribute('srcObject');
          }
        };

        peerConnection.oniceconnectionstatechange = (evt) => {
          console.log('oniceconnectionstatechange', evt)
        };

        peerConnection.onicegatheringstatechange = (evt) => {
          console.log('onicegatheringstatechange', evt)
        };

        peerConnection.onsignalingstatechange = (evt) => {
          console.log('onsignalingstatechange', evt)
        };

        return peerConnection;
      },

      handleHangup({ room_id, user_id }) {
        console.log('[handleHangup] Remote user hung up.', user_id)
        if (room_id !== this.room_id) {
          console.log(`Mismatching room ID. Current ${this.room_id} Incoming: ${room_id}`);
          return;
        }

        if (this.room_members[user_id]) {
          console.log('[handleHangup] Removing remote user.', user_id);
          const user = this.room_members[user_id].user_details;
          this.$delete(this.room_members, user_id);

          console.log('[handleHangup] User removed', user)
          if (this.$refs[`remoteVideo-${user.id}`] && this.$refs[`remoteVideo-${user.id}`].srcObject) {
            // this.$refs[`remoteVideo-${user.id}`].removeAttribute('src');
            // this.$refs[`remoteVideo-${user.id}`].removeAttribute('srcObject');
            // this.$refs[`remoteVideo-${user.id}`].srcObject = null;
          }

          this.$toast.error(`${user.username} left the room`);
        } else {
          console.error('[handleHangup] Remote user not found in the local list.', user_id)
        }
      },

      /**
       * When user receives an offer from a peer, most likely there is no peer connection between them. So we need to create
       * a peer connection and add the user details in the client side of this user.
       *
       * @param room_id
       * @param user
       * @param target_user_id
       * @param sdp
       * @returns {Promise<void>}
       */
      async handleRemoteOffer({ room_id, user, target_user_id, sdp }) {
        console.log('[handleRemoteOffer] Offer Received', { room_id, user, target_user_id, sdp });

        if (!this.room_members[user.id]) {
          console.log('[handleRemoteOffer] User not in member list. Adding to member list.', { room_id, user, target_user_id, sdp });
          this.processUserJoin(user);
        }

        const peerConnection = this.room_members[user.id].peer_connection;

        try {
          const sessionDesc = new RTCSessionDescription(sdp);
          await peerConnection.setRemoteDescription(sessionDesc);
          const answer = await peerConnection.createAnswer();
          await peerConnection.setLocalDescription(answer);

          const payload = {
            room_id: this.room_id,
            user: this.$auth.user,
            target_user_id: user.id,
            sdp: peerConnection.localDescription
          };

          this.$Signalling.send('ANSWER', payload);
        } catch (e) {
          console.error('handleRemoteOffer', e)
        }
      },

      async handleRemoteAnswer({ user, sdp }) {
        console.log('[handleRemoteAnswer] Answer received from user: ', user.username);
        if (!this.room_members[user.id] || !this.room_members[user.id].peer_connection) {
          return;
        }

        const sessionDesc = new RTCSessionDescription(sdp);
        await this.room_members[user.id].peer_connection.setRemoteDescription(sessionDesc)
      },

      async handleNewICECandidate({ user, candidate }) {
        if (!this.room_members[user.id] || !this.room_members[user.id].peer_connection) {
          console.error('[handleNewICECandidate] Error adding ice candidate', {
            room_member: this.room_members[user.id],
            peer_connection: this.room_members[user.id].peer_connection
          });
          return;
        }
        console.log('[handleNewICECandidate]', candidate);
        const rtcCandidate = new RTCIceCandidate(candidate);
        try {
          await this.room_members[user.id].peer_connection.addIceCandidate(rtcCandidate);
          console.log('[handleNewICECandidate] Successfully added ice candidate.')
        } catch (e) {
          console.error('[handleNewICECandidate]', {e, candidate})
        }
      },

      async initiateLocalVideo() {
        try {
          const localStream = await navigator.mediaDevices.getUserMedia({
            audio: true,
            video: {
              width: {
                ideal: 320
              },
              height: {
                ideal: 240
              },
              aspectRatio: {ideal: 1.7777777778}
            },

          });
          console.log('[LocalStream]', localStream)
          this.$refs.localVideo.srcObject = localStream;
          this.localStream = localStream;
        } catch (e) {
          console.error(e)
        }
      },

      async hangup() {
        console.log('[hangup]');
        // this.closeVideoCall();
        this.$Signalling.send('HANGUP', {
          room_id: this.room_id,
          user_id: this.$auth.user.id
        });

        this.$delete(this.room_members, this.$auth.user.id);
        // await this.$router.push('/');
      },

      // async closeVideoCall() {
      //   if (!this.webrtc.pc) {
      //     return;
      //   }
      //
      //   this.webrtc.pc.ontrack = null;
      //   this.webrtc.pc.onremovetrack = null;
      //   this.webrtc.pc.onremovestream = null;
      //   this.webrtc.pc.onicecandidate = null;
      //   this.webrtc.pc.onicecandidate = null;
      //   this.webrtc.pc.oniceconnectionstatechange = null;
      //   this.webrtc.pc.onsignalingstatechange = null;
      //   this.webrtc.pc.onicegatheringstatechange = null;
      //   this.webrtc.pc.onnegotiationneeded = null;
      //
      //   if (this.$refs.remoteVideo && this.$refs.remoteVideo.srcObject) {
      //     this.$refs.remoteVideo.srcObject.getTracks().forEach(track => track.stop());
      //     this.$refs.remoteVideo.removeAttribute('src');
      //     this.$refs.remoteVideo.removeAttribute('srcObject');
      //   }
      //
      //   if (this.$refs.localVideo && this.$refs.localVideo.srcObject) {
      //     this.$refs.localVideo.srcObject.getTracks().forEach(track => track.stop());
      //     this.$refs.localVideo.removeAttribute('src');
      //     this.$refs.localVideo.removeAttribute('srcObject');
      //   }
      //
      //   this.webrtc.pc.close();
      //   this.webrtc.pc = null;
      // }
    }
  }
</script>

<style scoped>
  .local-video {
    height: 100%;
    min-height: 100%;
    max-width: 100%;
    /*object-fit: cover;*/
  }
</style>
