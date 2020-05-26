<template>
  <div>
    <v-layout column>
      <v-alert
        v-if="error.isError"
        text
        prominent
        type="error"
        icon="mdi-alert-circle"
        >{{ error.errorMessage }}</v-alert
      >
      <v-row v-else no-gutters>
        <v-col cols="12">
          <div class="grid-wrapper">
            <div class="grid-containter">
              <div
                class="grid-item"
                :style="users.length == 1 ? { height: '92vh' } : {}"
                :key="user"
                v-for="user in users"
              >
                <video width="100%" height="100%" controls>
                  <video
                    id="remote-video"
                    class="local-video"
                    autoplay
                    playsinline
                    ref="remoteVideo"
                  ></video>
                </video>
              </div>
            </div>
            <div class="grid-item-me" :class="{ mini: minimizeMyWindow }">
              <v-btn icon color="red" @click="toggleMyWindow">
                <v-icon>mdi-close</v-icon>
              </v-btn>
              <video
                class="local-video"
                autoplay
                playsinline
                muted
                ref="localVideo"
              ></video>
            </div>
            <v-btn
              absolute
              dark
              fab
              bottom
              :style="{
                left: '50%',
                bottom: '40px',
                transform: 'translateX(-50%)'
              }"
              color="red"
              @click="hangup"
            >
              <v-icon>mdi-phone-hangup</v-icon>
            </v-btn>
          </div>
        </v-col>
      </v-row>
    </v-layout>
  </div>
</template>

<script>
export default {
  name: "index",
  data() {
    return {
      minimizeMyWindow: false,
      users: [1],
      error: {
        isError: false,
        errorMessage:
          "Something is not right. Please try again with correct URL."
      },
      room_types: {
        a: "AUDIO",
        av: "AUDIO_VIDEO"
      },
      webrtc: {
        room_id: null,
        room_type: null,
        room_members: {},
        localStream: null,
        pc: null,
        mediaStreamConstraint: {
          audio: true,
          video: true
        }
      }
    };
  },
  async mounted() {
    const { room_type, room_id } = this.$route.params;
    if (!room_type || !room_id) {
      this.error.isError = true;
      return;
    }

    if (room_type !== "a" && room_type !== "av") {
      this.error.isError = true;
      return;
    }

    if (room_id.length !== 32) {
      this.error.isError = true;
      return;
    }

    this.webrtc.room_type = this.room_types[room_type];
    this.webrtc.room_id = room_id;

    window.onunload = () => {
      this.hangupIfConnected();
    };

    // everything looks okay, its time to initiate the websocket connection to the server.
    await this.$Signalling.open(this.$auth.getToken("local"));
    this.createOrJoinRoom(room_type, room_id);
    this.$Signalling.registerOnSignallingMessageHandler(
      this.signallingHandler.bind(this)
    );
  },

  beforeDestroy() {
    this.hangupIfConnected();
  },

  methods: {
    toggleMyWindow() {
      this.minimizeMyWindow = !this.minimizeMyWindow;
    },
    hangupIfConnected() {
      console.log("hangupIfConnected called");
      if (this.webrtc.room_members[this.$auth.user.id]) {
        this.hangup();
      }
    },

    createOrJoinRoom(roomType, roomId) {
      const payload = {
        room_type: roomType === "av" ? "AUDIO_VIDEO" : "AUDIO",
        room_id: roomId
      };

      this.$Signalling.send("CREATE_OR_JOIN", payload);
    },

    signallingHandler(data) {
      data = JSON.parse(data);
      switch (data.type) {
        case "error":
          this.error.isError = true;
          this.error.errorMessage = data.payload;
          return;
        case "OFFER":
          this.handleRemoteOffer(data.payload);
          break;
        case "ANSWER":
          this.handleRemoteAnswer(data.payload);
          break;
        case "ICE_CANDIDATE":
          this.handleNewICECandidate(data.payload);
          break;
        case "ROOM_JOIN":
          this.handleRoomJoin(data.payload);
          break;
        case "HANGUP":
          this.handleHangup(data.payload);
          break;
      }
    },

    createPeerConnection() {
      if (this.webrtc.pc) {
        return;
      }

      this.webrtc.pc = new RTCPeerConnection({
        iceServers: [
          {
            urls: "stun:stun.l.google.com:19302"
          }
        ]
      });

      this.webrtc.pc.ontrack = event => {
        console.log("ontrack", event.streams);
        this.$refs.remoteVideo.srcObject = event.streams[0];
      };

      this.webrtc.pc.onicecandidate = ({ candidate }) => {
        console.log("onicecandidate:", candidate);
        if (candidate) {
          const payload = {
            room_id: this.webrtc.room_id,
            user_id: this.$auth.user.id,
            target_user_id: null,
            candidate: candidate
          };

          this.$Signalling.send("ICE_CANDIDATE", payload);
        }
      };

      this.webrtc.pc.onnegotiationneeded = async evt => {
        console.log("onnegotiationneeded", evt);
        try {
          const offer = await this.webrtc.pc.createOffer();
          await this.webrtc.pc.setLocalDescription(offer);
          const payload = {
            room_id: this.webrtc.room_id,
            user_id: this.$auth.user.id,
            target_user_id: null,
            sdp: this.webrtc.pc.localDescription
          };

          this.$Signalling.send("OFFER", payload);
        } catch (e) {
          console.log(e);
        }
      };

      this.webrtc.pc.onremovetrack = evt => {
        console.log("onremovetrack", evt);
        const stream = this.$refs.remoteVideo.srcObject;
        if (stream.getTracks().length === 0) {
          this.$refs.remoteVideo.srcObject.removeAttribute("src");
          this.$refs.remoteVideo.srcObject.removeAttribute("srcObject");
        }
      };

      this.webrtc.pc.oniceconnectionstatechange = evt => {
        console.log("oniceconnectionstatechange", evt);
      };

      this.webrtc.pc.onicegatheringstatechange = evt => {
        console.log("onicegatheringstatechange", evt);
      };

      this.webrtc.pc.onsignalingstatechange = evt => {
        console.log("onsignalingstatechange", evt);
      };
    },

    async handleRoomJoin({ room_id, user }) {
      if (room_id !== this.webrtc.room_id) {
        console.log(
          `Mismatching room ID. Current ${this.room_id} Incoming: ${room_id}`
        );
        return;
      }

      if (this.webrtc.room_members[user.id]) {
        console.log("Member already inside room.", this.webrtc.room_members);
        return;
      }

      this.webrtc.room_members[user.id] = user;
      const messageString = `${
        user.id === this.$auth.user.id ? "You" : user.username
      } joined the room`;
      this.$toast.success(messageString);

      if (user.id === this.$auth.user.id) {
        this.initiateLocalVideo();
      }
    },

    handleHangup({ room_id, user_id }) {
      if (room_id !== this.webrtc.room_id) {
        console.log(
          `Mismatching room ID. Current ${this.room_id} Incoming: ${room_id}`
        );
        return;
      }

      if (this.webrtc.room_members[user_id]) {
        const user = this.webrtc.room_members[user_id];
        delete this.webrtc.room_members[user_id];

        if (this.$refs.remoteVideo.srcObject) {
          this.$refs.remoteVideo.srcObject = null;
          this.$refs.remoteVideo.removeAttribute("src");
          this.$refs.remoteVideo.removeAttribute("srcObject");
        }

        this.$toast.error(`${user.username} left the room`);
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

        this.$Signalling.send("ANSWER", payload);
      } catch (e) {
        console.error("handleRemoteOffer");
      }
    },

    async handleRemoteAnswer(answerPayload) {
      const sessionDesc = new RTCSessionDescription(answerPayload.sdp);
      await this.webrtc.pc.setRemoteDescription(sessionDesc);
    },

    async handleNewICECandidate(candidatePayload) {
      console.log("ReceivedCandidate", candidatePayload);
      const candidate = new RTCIceCandidate(candidatePayload.candidate);
      try {
        await this.webrtc.pc.addIceCandidate(candidate);
      } catch (e) {
        console.error("icecandidate error", { e, candidate });
      }
    },

    async initiateLocalVideo() {
      // peer connection already established.
      if (this.webrtc.pc) {
        return;
      }

      this.createPeerConnection();

      try {
        const localStream = await navigator.mediaDevices.getUserMedia(
          this.webrtc.mediaStreamConstraint
        );
        this.$refs.localVideo.srcObject = localStream;
        localStream
          .getTracks()
          .forEach(track => this.webrtc.pc.addTrack(track, localStream));
      } catch (e) {
        console.log("Error in something");
        this.error.isError = true;
        this.error.errorMessage = e.toString();
      }
    },

    async hangup() {
      this.closeVideoCall();
      this.$Signalling.send("HANGUP", {
        room_id: this.webrtc.room_id,
        user_id: this.$auth.user.id
      });

      delete this.webrtc.room_members[this.$auth.user.id];
      await this.$router.push("/");
    },

    async closeVideoCall() {
      if (!this.webrtc.pc) {
        return;
      }

      this.webrtc.pc.ontrack = null;
      this.webrtc.pc.onremovetrack = null;
      this.webrtc.pc.onremovestream = null;
      this.webrtc.pc.onicecandidate = null;
      this.webrtc.pc.onicecandidate = null;
      this.webrtc.pc.oniceconnectionstatechange = null;
      this.webrtc.pc.onsignalingstatechange = null;
      this.webrtc.pc.onicegatheringstatechange = null;
      this.webrtc.pc.onnegotiationneeded = null;

      if (this.$refs.remoteVideo && this.$refs.remoteVideo.srcObject) {
        this.$refs.remoteVideo.srcObject
          .getTracks()
          .forEach(track => track.stop());
        this.$refs.remoteVideo.removeAttribute("src");
        this.$refs.remoteVideo.removeAttribute("srcObject");
      }

      if (this.$refs.localVideo && this.$refs.localVideo.srcObject) {
        this.$refs.localVideo.srcObject
          .getTracks()
          .forEach(track => track.stop());
        this.$refs.localVideo.removeAttribute("src");
        this.$refs.localVideo.removeAttribute("srcObject");
      }

      this.webrtc.pc.close();
      this.webrtc.pc = null;
    }
  }
};
</script>
<style>
html {
  overflow-y: hidden !important;
}
</style>
<style scoped>
.grid-wrapper {
  width: 100%;
  height: 93vh;
  overflow-y: scroll;
  background: rgb(17, 17, 17);
  position: relative;
}
.grid-containter {
  background: rgb(17, 17, 17);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  position: absolute;
  left: 0;
  top: 0;
  width: 100%;
}
.grid-containter .grid-item {
  flex-grow: 1;
  width: 33%;
  background: rgb(17, 17, 17);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 10px;
}

.grid-item-me {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  position: sticky;
  left: 87%;
  top: 85%;
  background: transparent;
  z-index: 999999;
  width: 200px;
  height: 0;
  transition: all 1s ease;
}

.grid-item-me video {
  height: 200px;
  width: 250px;
}

.grid-item-me button {
  position: absolute;
  right: -28px;
  top: -96px;
  z-index: 999999;
}

.grid-item-me.mini video {
  transition: all 1s ease;
  width: 0;
  height: 0;
}

.grid-item-me.mini button {
  transition: all 1s ease;
  right: 0px;
  top: 52px;
}

@media (max-width: 640px) {
  .grid-containter .grid-item {
    width: 50%;
  }

  .grid-item-me {
    top: 89%;
    left: 100%;
    width: 138px;
    height: 100px;
  }

  .grid-item-me button {
    right: -3px;
    top: -47px;
  }
}
</style>
