<template>
  <v-layout
    column
    justify-center
    align-center
  >
    <v-row no-gutters>
      <v-col md="6">
        <v-card
          style="max-width: 100%"
          class="pa-2"
          outlined
          tile
        >
          <!--
          <video ref="selfView" style="max-width: 100%" autoplay playsinline></video>
          -->
          <v-textarea
            v-model="inputText"
            outlined
            name="input-7-4"
            label="Input"
          ></v-textarea>
        </v-card>
      </v-col>
      <v-col color="green" md="6">
        <v-card
          style="max-width: 100%"
        class="pa-2"
        outlined
        tile
        >
          <!--
          <video ref="remoteView" style="max-width: 100%" autoplay playsinline></video>
          -->
          <v-textarea
            v-model="outputText"
            outlined
            name="input-7-4"
            label="Output"
          ></v-textarea>
        </v-card>
      </v-col>

      <v-btn depressed color="primary" @click="sendData">Send</v-btn>
    </v-row>
  </v-layout>
</template>
<script>
export default {
  data() {
    return {
      mediaConstraint: {
        video: true
      },

      localPeerConnection: null,
      remotePeerConnection: null,

      sendChannel: null,
      receiveChannel: null,

      inputText: null,
      outputText: null
    }
  },

  mounted() {
    this.init()
    this.ws()
  },

  methods: {
    async ws() {
      const _this = this
      const socket = new WebSocket(`${process.env.WS_URL}?auth_token=${_this.$auth.getToken('local')}`)

      setInterval(() => {
        console.log('Sending ping data to server')
        socket.send(JSON.stringify({data: 'PING'}))
      }, 6000)

      socket.onopen = function (ev) {
        console.log('Websocket connection established.')
      }

      socket.onmessage = function (ev) {
        _this.$toast.show(ev.data)
      }

      socket.onerror = function (error) {
        _this.$toast.error(error.message)
      }

      socket.onclose = function () {
        _this.$toast.error('Connection closed')
      }
    },

    async init() {
      this.localPeerConnection = new RTCPeerConnection({
        iceServers: [
          {
            urls: 'stun:stun.l.google.com:19302'
          }
        ]
      });

      console.log('Created local peer connection object');

      this.sendChannel = this.localPeerConnection.createDataChannel(`sendDataChannel`, null);
      console.log('created send data channel');

      this.remotePeerConnection = new RTCPeerConnection({
        iceServers: [
          {
            urls: 'stun:stun.l.google.com:19302'
          }
        ]
      });


      this.localPeerConnection.onicecandidate = ({ candidate }) => {
        console.log('Local ICE callback.');
        if (candidate) {
          this.remotePeerConnection.addIceCandidate(candidate)
            .then(() => {console.log('Add ICE candidate success.')}, () => { console.log('Add ICE candidate failed.')})
        }
      }

      this.remotePeerConnection.onicecandidate = ({ candidate }) => {
        console.log('remote ICE callback.');
        if (candidate) {
          this.localPeerConnection.addIceCandidate(candidate)
            .then(() => {console.log('Add ICE candidate success.')}, () => { console.log('Add ICE candidate failed.')})
        }
      }

      try {
        const offer = await this.localPeerConnection.createOffer();
        console.log('Offer', offer);
        await this.localPeerConnection.setLocalDescription(offer);
        await this.remotePeerConnection.setRemoteDescription(offer);

        const answer = await this.remotePeerConnection.createAnswer();
        console.log('Answer', answer)
        await this.remotePeerConnection.setLocalDescription(answer);
        await this.localPeerConnection.setRemoteDescription(answer)
      } catch (e) {
        console.error('Offfer create error: ' + e.toString())
      }


      this.remotePeerConnection.ondatachannel = ({ channel }) => {
        this.receiveChannel = channel;
        this.receiveChannel.onmessage = ({ data }) => {
          this.outputText = data;
        }
      }
      // try {
      //   const stream = await navigator.mediaDevices.getUserMedia(this.mediaConstraint);
      //   stream.getTracks().forEach(track => {
      //     console.log(track.kind)
      //   });
      //   this.$refs.selfView.srcObject = stream
      //   this.$refs.remoteView.srcObject = stream
      // } catch (err) {
      //   console.error(err)
      // }
    },

    sendData() {
      this.sendChannel.send(this.inputText)
    }
  }
}
</script>
