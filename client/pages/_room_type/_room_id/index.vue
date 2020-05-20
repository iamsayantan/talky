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
        <video class="local-video" autoplay playsinline muted ref="localVideo"></video>
      </v-col>
      <v-col cols="6">
        Remote Video
        <video id="remote-video" class="local-video" autoplay playsinline muted ref="remoteVideo"></video>
      </v-col>
    </v-row>
    <v-btn
      absolute
      dark
      fab
      bottom
      :style="{left: '50%', bottom: '40px', transform:'translateX(-50%)'}"
      color="red"
    >
      <v-icon>mdi-phone-hangup</v-icon>
    </v-btn>
  </v-layout>
</template>

<script>
  export default {
    name: "index",
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
        webrtc: {
          room_id: null,
          room_type: null,
          localStream: null,
          pc: null,
          mediaStreamConstraint: {
            audio: true,
            video: true
          }
        }
      }
    },
    async mounted() {
      const { room_type, room_id } = this.$route.params;
      if (!room_type || !room_id) {
        this.error.isError = true
        return;
      }

      if (room_type !== 'a' && room_type !== 'av') {
        this.error.isError = true;
        return;
      }

      if (room_id.length !== 32) {
        this.error.isError = true;
        return;
      }

      this.webrtc.room_type = this.room_types[room_type];
      this.webrtc.room_id = room_id;

      // everything looks okay, its time to initiate the websocket connection to the server.
      await this.$Signalling.open(this.$auth.getToken('local'));
      this.createOrJoinRoom(room_type, room_id);
      this.$Signalling.registerOnSignallingMessageHandler(this.signallingHandler.bind(this));

      console.log('Inside mounted block')
      this.initiateLocalVideo()
    },

    beforeDestroy() {

    },

    methods: {
      createOrJoinRoom(roomType, roomId) {
        const payload = {
          room_type: roomType === 'av' ? 'AUDIO_VIDEO' : 'AUDIO',
          room_id: roomId
        };

        this.$Signalling.send('CREATE_OR_JOIN', payload)
      },

      signallingHandler(data) {
        data = JSON.parse(data);
        switch (data.type) {
          case 'error':
            this.error.isError = true;
            this.error.errorMessage = data.payload;
            return;
          case 'OFFER':
            this.handleRemoteOffer(data.payload);
            break;
          case 'ANSWER':
            this.handleRemoteAnswer(data.payload);
            break;
          case 'ICE_CANDIDATE':
            this.handleNewICECandidate(data.payload);
            break;
        }
      },

      createPeerConnection() {
        if (this.webrtc.pc) {
          return;
        }

        this.webrtc.pc = new RTCPeerConnection(null);

        this.webrtc.pc.ontrack = event => {
          console.log('ontrack', event.streams);
          this.$refs.remoteVideo.srcObject = event.streams[0]
        };

        this.webrtc.pc.onicecandidate = ({candidate}) => {
          console.log('onicecandidate:', candidate);
          if (candidate) {
            const payload = {
              room_id: this.webrtc.room_id,
              user_id: this.$auth.user.id,
              target_user_id: null,
              candidate: candidate
            };

            this.$Signalling.send('ICE_CANDIDATE', payload)
          }
        };

        this.webrtc.pc.onnegotiationneeded = async (evt) => {
          console.log('onnegotiationneeded', evt);
          try {
            const offer = await this.webrtc.pc.createOffer();
            await this.webrtc.pc.setLocalDescription(offer);
            const payload = {
              room_id: this.webrtc.room_id,
              user_id: this.$auth.user.id,
              target_user_id: null,
              sdp: this.webrtc.pc.localDescription
            };

            this.$Signalling.send('OFFER', payload)
          } catch (e) {
            console.log(e)
          }
        };

        this.webrtc.pc.onremovetrack = (evt) => {
          console.log('onremovetrack', evt);
        };

        this.webrtc.pc.oniceconnectionstatechange = (evt) => {
          console.log('oniceconnectionstatechange', evt)
        };

        this.webrtc.pc.onicegatheringstatechange = (evt) => {
          console.log('onicegatheringstatechange', evt)
        }

        this.webrtc.pc.onsignalingstatechange = (evt) => {
          console.log('onsignalingstatechange', evt)
        }
      },

      async handleRemoteOffer(offerPayload) {
        try {
          const sessionDesc = new RTCSessionDescription(offerPayload.sdp);
          await this.webrtc.pc.setRemoteDescription(sessionDesc);
          const answer = await this.webrtc.pc.createAnswer();
          await this.webrtc.pc.setLocalDescription(answer);

          const payload = {
            room_id: this.webrtc.room_id,
            user_id: this.$auth.user.id,
            target_user_id: offerPayload.user_id,
            sdp: this.webrtc.pc.localDescription
          };

          this.$Signalling.send('ANSWER', payload);
        } catch (e) {
          console.error('handleRemoteOffer')
        }
      },

      async handleRemoteAnswer(answerPayload) {
        const sessionDesc = new RTCSessionDescription(answerPayload.sdp);
        await this.webrtc.pc.setRemoteDescription(sessionDesc)
      },

      async handleNewICECandidate(candidatePayload) {
        console.log('ReceivedCandidate', candidatePayload)
        const candidate = new RTCIceCandidate(candidatePayload.candidate)
        try {
          await this.webrtc.pc.addIceCandidate(candidate);
        } catch (e) {
          console.error('icecandidate error', {e, candidate})
        }
      },

      async initiateLocalVideo() {
        // peer connection already established.
        if (this.webrtc.pc) {
          return;
        }

        this.createPeerConnection();

        try {
          const localStream = await navigator.mediaDevices.getUserMedia(this.webrtc.mediaStreamConstraint);
          this.$refs.localVideo.srcObject = localStream;
          localStream.getTracks().forEach(track => this.webrtc.pc.addTrack(track, localStream))
        } catch (e) {
          console.log('Error in something')
          this.error.isError = true;
          this.error.errorMessage = e.message
        }
      }
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
