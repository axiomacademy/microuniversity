<template>
  <div id="login" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col justify-center items-center">
    <div class="shadow-sm px-10 py-8 bg-white rounded-md flex flex-col justify-center items-start">
      <img src="../assets/logo-transparent-dark.png" class="w-20 h-20 mb-4"/>
      <h1 class="font-display text-3xl text-text font-medium">Login to <span class="text-primary">Axiom</span></h1>
      <h2 class="font-display text-sm text-gray-600 font-regular py-2">Please login with the credentials that have been provided to you over email</h2>
      <input v-model="username" type="text" placeholder="Username" class="bg-gray-100 p-2 w-full rounded font-display border border-transparent focus:outline-none focus:ring-2 focus:ring-purple-600 focus:border-transparent mt-4" :disabled="loading">
      <input v-model="password" type="password" placeholder="Password" class="bg-gray-100 p-2 w-full rounded font-display border border-transparent focus:outline-none focus:ring-2 focus:ring-purple-600 focus:border-transparent mt-2" :disabled="loading">
      <span v-if="errorText != ''" class="text-red-500 my-3 font-body text-xs">{{ errorText }}</span>
      <button @click="loginLearner" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-4 rounded w-full flex flex-row justify-center items-center">
        <BeatLoader :size="8.5" color="#ffffff" v-if="loading" />
        <div v-else>
          Login
        </div>
      </button>
    </div>
  </div>
    
</template>

<script>
import { BeatLoader } from '@saeris/vue-spinners'
import { loginLearner } from '../services/LearnerService.js';

export default {
  name: 'Login',
  components: {
    BeatLoader,
  },
  data: function() {
    return {
      loading: false,
      errorText: "",
      username: "",
      password:"",
    }
  },
  created: function () {
    // # Check if the JWT exists
    let token = localStorage.getItem("token")
    if(token != null) { 
      this.$router.push({ name: 'home' })
    }
  },
  methods: {
    loginLearner: async function () {
      // Attempt to login
      this.loading = true
      try {
        let response = await loginLearner(this.username, this.password)

        if (response.jwt != null) {
          // Set the jwt to the localstorage and route to the home page
          localStorage.setItem("token", response.jwt)
          this.$router.push({ name: 'home'})
        } else {
          this.errorText = "We can't log you in right now. Try again later :("
        }
  
        this.loading = false
      } catch (err) {
        if(err == 401) {
          //Unauthorised
          this.errorText = "Invalid username or password! Please try again."
        } else {
          this.errorText = "We can't log you in right now. Try again later :("
        }

        this.loading = false
      }
      return
    },
  }
}
</script>

<style>

</style>
