<template>
  <v-layout
    column
  >
    <v-row no-gutters>
      <div>
        <span class="display-3" style="color: #069587">Create Your Room</span>
        <div style="margin-top: 20px">
          <span class="subtitle-1">Invite you friends to your room to have a realtime audio video conferencing.</span>
          <br>
          <span class="subtitle-1">Audio only rooms can have upto 20 members while audio video rooms can have maximum 4 members.</span>
        </div>
      </div>
      <v-layout
        column
        justify-center
        align-center
      >
        <div>
          <v-radio-group v-model="room_type" row>
            <v-radio label="Audio Only" value="AUDIO" color="teal"></v-radio>
            <v-radio label="Audio Video" value="AUDIO_VIDEO" color="teal"></v-radio>
          </v-radio-group>
        </div>
        <div>
          <v-btn color="teal" dark large block outlined @click="handleCreateRoom">Create Room</v-btn>
        </div>
      </v-layout>
    </v-row>
  </v-layout>
</template>

<script>
  import { v4 } from 'uuid'

  export default {
    name: "index",
    data() {
      return {
        room_type: 'AUDIO'
      }
    },
    methods: {
      handleCreateRoom() {
        let roomId = v4();
        roomId = roomId.split('-').join('');

        const path = `${this.room_type === 'AUDIO' ?  'a' : 'av' }/${roomId}`;
        this.$router.push(path)
      }
    }
  }
</script>

<style scoped>

</style>
