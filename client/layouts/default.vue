<template>
  <v-app>
    <v-app-bar
      color="white"
      fixed
      app
    >
      <v-toolbar-title><v-icon color="teal" large>mdi-chat-plus</v-icon> <span style="color: #069587; font-weight: bold">konference</span></v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn icon @click="handleLogout">
        <v-icon color="teal">mdi-power</v-icon>
      </v-btn>
    </v-app-bar>
    <v-content>
      <nuxt />
<!--      <v-container :fluid="true">-->
<!--        <nuxt />-->
<!--      </v-container>-->
    </v-content>
  </v-app>
</template>

<script>
export default {
  data () {
    return {
    }
  },
  methods: {
    async handleLogout() {
      try {
        this.$auth.reset();
        await this.$router.push('/login');
      } catch (e) {
        console.error(e)
      }
    }
  }
}
</script>
