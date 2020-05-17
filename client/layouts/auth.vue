<template>
  <v-app>
    <v-content>
      <div class="d-flex flex-column" style="min-height: 100vh">
        <div class="teal d-flex" style="flex: 2;">
          <v-layout
            column
            justify-center
            align-center
          >
            <v-icon large dark class="display-4">mdi-chat-plus</v-icon>
            <span class="display-3 font-weight-medium" style="color: white">konference</span>
          </v-layout>
        </div>
        <div style="flex: 5">
          <nuxt />
        </div>
      </div>
    </v-content>
  </v-app>
</template>

<script>
  export default {
    name: "auth"
  }
</script>

<style scoped>

</style>
