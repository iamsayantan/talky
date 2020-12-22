<template>
  <v-layout
    column
    justify-center
    align-center
  >
    <div style="margin-top: 20px; width: 90%">
      <v-text-field
        v-model="login.username"
        outlined
        name="username"
        label="Username"
        required
      ></v-text-field>
      <v-text-field
        v-model="login.password"
        outlined
        type="password"
        name="password"
        label="Password"
        required
      ></v-text-field>
      <v-btn :loading="loading" color="teal" dark large block @click="handleLogin">Login</v-btn>
    </div>
    <div style="margin-top: 15px; color: teal">
      <v-btn text small color="teal" class="subtitle-1" to="register">Register</v-btn>
    </div>
  </v-layout>
</template>

<script>
  export default {
    name: "login",
    auth: false,
    layout: 'auth',

    asyncData({ app, redirect }) {
      if (app.$auth.loggedIn) {
        redirect('/');
      }
    },

    data() {
      return {
        loading: false,
        login: {
          username: null,
          password: null
        }
      }
    },

    methods: {
      async handleLogin() {
        const returnTo = this.$route.query.return_to || '/'
        this.loading = true;
        try {
          await this.$auth.loginWith('local', {
            data: this.login
          });

          await this.$router.push(returnTo)
        } catch (e) {
          console.error(e);
        }
        this.loading = false;
      }
    }
  }
</script>

<style scoped>

</style>
