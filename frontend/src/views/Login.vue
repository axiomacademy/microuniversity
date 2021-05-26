<template>
  <div id="login" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col justify-center items-center">
    <div v-if="!verificationSent" class="shadow-sm px-10 py-8 bg-white rounded-md flex flex-col justify-center items-start">
      <img src="../assets/logo-transparent-dark.png" class="w-20 h-20 mb-4"/>
      <h1 class="font-display text-3xl text-text font-medium">Login to <span class="text-primary">Axiom</span></h1>
      <h2 class="font-display text-sm text-gray-600 font-regular py-2">Enter your email to get to changing the way you think</h2>
      <input v-model="email" type="text" placeholder="Email" class="bg-gray-100 p-2 w-full rounded font-display border border-transparent focus:outline-none focus:ring-2 focus:ring-purple-600 focus:border-transparent mt-4" :disabled="loading">
      <span v-if="errorText != ''" class="text-red-500 my-3 font-body text-xs">{{ errorText }}</span>

      <button @click="verifyEmail" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-4 rounded w-full flex flex-row justify-center items-center">
        <BeatLoader :size="8.5" color="#ffffff" v-if="loading" />
        <div v-else>
          Login
        </div>
      </button>
    </div>
    <div v-else class="shadow-sm px-10 py-8 bg-white rounded-md flex flex-col justify-center items-start">
      <img src="../assets/mail-sent.png" class="h-20 mb-4 self-center"/>
      <h1 class="font-display text-xl text-text font-medium">Magic link sent!</h1>
      <h2 class="font-display text-sm text-gray-600 font-regular py-2">We've sent a magic link to <span class="text-primary">{{ email }}</span>. Click it to login and start learning 😎</h2>

      <button @click="verificationSent = !verificationSent" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-4 rounded w-full flex flex-row justify-center items-center">
        <BeatLoader :size="8.5" color="#ffffff" v-if="loading" />
        <div v-else>
          Back to login
        </div>
      </button>
    </div>
  </div> 
</template>

<script>
import { BeatLoader } from '@saeris/vue-spinners'
import { loginEmail } from '../services/LoginService.js';

export default {
  name: 'Login',
  components: {
    BeatLoader,
  },
  data: function() {
    return {
      loading: false,
      errorText: "",
      email: "",
      verificationSent: false,
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
    verifyEmail: async function () {
      // Attempt to login
      this.loading = true
      try {
        await loginEmail(this.email)
        
        this.verificationSent = true
        this.loading = false
      } catch (err) {
        console.log(err)
        this.errorText = "We can't log you in right now. Try again later :("
        this.loading = false
      }
      return
    },
  }
}
</script>

<style>

</style>
