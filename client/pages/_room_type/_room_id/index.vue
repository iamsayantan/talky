<template>
  <v-layout
    column
  >
    <v-alert
      v-if="error"
      text
      prominent
      type="error"
      icon="mdi-alert-circle"
    >
      Something is not right. Please try again with correct URL.
    </v-alert>
  </v-layout>
</template>

<script>
  export default {
    name: "index",
    data() {
      return {
        error: false
      }
    },
    mounted() {
      const { room_type, room_id } = this.$route.params;
      if (!room_type || !room_id) {
        this.error = true
        return;
      }

      if (room_type !== 'a' && room_type !== 'av') {
        this.error = true;
        return;
      }

      if (room_id.length !== 32) {
        this.error = true;
        return;
      }

      // everything looks okay, its time to initiate the websocket connection to the server.
      this.$Signalling.open(this.$auth.getToken('local'));
      this.createOrJoinRoom(room_type, room_id);
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
      }
    }
  }
</script>

<style scoped>

</style>
