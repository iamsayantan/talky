<template>
  <v-layout
    column
    justify-center
    align-center
  >
    <div style="margin-top: 20px; width: 90%">
      <v-text-field
        v-model="registration.username"
        outlined
        name="username"
        label="Username"
        required
      ></v-text-field>
      <v-text-field
        v-model="registration.password"
        outlined
        type="password"
        name="password"
        label="Password"
        required
      ></v-text-field>
      <v-text-field
        v-model="registration.first_name"
        outlined
        name="first_name"
        label="First Name"
        required
      ></v-text-field>
      <v-text-field
        v-model="registration.last_name"
        outlined
        name="last_name"
        label="Last Name"
        required
      ></v-text-field>
      <v-btn :loading="loading" color="teal" dark large block @click="handleRegister">Register</v-btn>
    </div>
    <div style="margin-top: 15px; color: teal">
      <v-btn to="login" text small color="teal" class="subtitle-1">Login</v-btn>
    </div>
  </v-layout>
</template>

<script>
  export default {
    name: "register",
    auth: false,
    layout: 'auth',
    data() {
      return {
        loading: false,
        registration: {
          username: null,
          password: null,
          first_name: null,
          last_name: null
        }
      }
    },
    methods: {
      async handleRegister() {
        try {
          const {data} = await this.$axios.post('/user/v1/register', this.registration);
          await this.$auth.loginWith('local', {
            data: this.registration
          });
          await this.$router.push('/')
        } catch (e) {
          console.error(e)
        }
      }
    }
  }
</script>

<style scoped>

</style>
