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
      row
      no-gutters
      style="max-height: 100vh; min-height: 100%"
    >
      <video class="local-video" autoplay playsinline muted ref="localVideo"></video>
<!--      <v-col cols="6"><video autoplay playsinline muted ref="localVideo1"></video></v-col>-->
<!--      <v-col cols="6"><video autoplay playsinline muted ref="localVideo2"></video></v-col>-->
<!--      <v-col cols="6"><video autoplay playsinline muted ref="localVideo3"></video></v-col>-->
<!--      <v-col cols="6"></v-col>-->
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
        webrtc: {
          localStream: null,
          mediaStreamConstraint: {
            audio: true,
            video: true
          }
        }
      }
    },
    mounted() {
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

      // everything looks okay, its time to initiate the websocket connection to the server.
      this.$Signalling.open(this.$auth.getToken('local'));
      this.createOrJoinRoom(room_type, room_id);

      this.initiateLocalVideo()
    },

    methods: {
      createOrJoinRoom(roomType, roomId) {
        const payload = {
          room_type: roomType === 'av' ? 'AUDIO_VIDEO' : 'AUDIO',
          room_id: roomId
        };

        // waiting 2 seconds to give enough time for the websocket connection to establish.
        setTimeout(() => {
          this.$Signalling.send('CREATE_OR_JOIN', payload)
        }, 2000)
      },

      async initiateLocalVideo() {
        try {
          const localStream = await navigator.mediaDevices.getUserMedia(this.webrtc.mediaStreamConstraint);
          this.webrtc.localStream = localStream;
          this.$refs.localVideo.srcObject = localStream;
          // this.$refs.localVideo1.srcObject = localStream;
          // this.$refs.localVideo2.srcObject = localStream;
          // this.$refs.localVideo3.srcObject = localStream;
        } catch (e) {
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
    object-fit: cover;
  }
</style>
