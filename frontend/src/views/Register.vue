<template>
  <div id="more-info" class="bg-purple-50 min-h-screen p-4 flex flex-col justify-center items-center">
    <div class="shadow-sm px-8 py-8 bg-white rounded-md flex flex-col justify-center items-start lg:w-4/12">
      <img src="../assets/welcome.png" class="w-25"/>
      <h1 class="font-display text-3xl text-text font-medium">Hello there!</h1>
      <h2 class="font-display text-sm text-gray-600 font-regular py-2">We're glad to have you join the Axiom community. To finish creating your account, we need a bit more information about you.
      </h2>
      <input v-model="firstName" type="text" placeholder="First Name" class="bg-gray-100 p-2 w-full rounded font-display border border-transparent focus:outline-none focus:ring-2 focus:ring-purple-600 focus:border-transparent mt-4" :disabled="loading">
      <input v-model="lastName" type="text" placeholder="Last Name" class="bg-gray-100 p-2 w-full rounded font-display border border-transparent focus:outline-none focus:ring-2 focus:ring-purple-600 focus:border-transparent mt-2" :disabled="loading">

      <div class="mt-4 text-sm text-text">Your detected timezone is <span class="font-bold">{{ timezone }}</span>. The local time here is <span class="text-secondary">{{ currentTime }}</span></div>
      
      <span v-if="errorText != ''" class="text-red-500 my-3 font-body text-xs">{{ errorText }}</span>

      <button @click="registerAccount" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-4 rounded w-full flex flex-row justify-center items-center">
        <BeatLoader :size="8.5" color="#ffffff" v-if="loading" />
        <div v-else>
          Continue
        </div>
      </button> 
    </div>
  </div>
</template>

<script>
import { BeatLoader } from '@saeris/vue-spinners'

import { updateSelf } from "../services/LearnerService.js"

import firebase from "firebase/app";
import "firebase/auth";

export default {
  name: 'Register',
  components: {
    BeatLoader,
  },
  data: function() {
    return {
      token: "",
      loading: false,
      timezone: "",
      currentTime: "",
      firstName: "",
      lastName: "",
      errorText: "",
    }
  },
  created () {
    firebase.auth().onAuthStateChanged(async (user) => {
      if (user) {
        this.token = await user.getIdToken(true)
        this.timezone = Intl.DateTimeFormat().resolvedOptions().timeZone

        let opts = {
          hour: 'numeric',
          minute:'2-digit',
          timeZone: this.timezone,
          timeZoneName: 'long'
        }
        this.currentTime = new Date().toLocaleString('en', opts) 
        setInterval(function(){ this.currentTime = new Date().toLocaleString('en', opts) }, 1000); 
      } else {
        this.$router.push({ name: 'login' })
      }
    })
  },
  methods: {
    async registerAccount() {
      this.loading = true
      this.errorText = ""

      if (this.firstName == "" || this.lastName == "") {
        this.errorText = "Please fill in all the fields!"
        this.loading = false
        return
      }
      try {
        await updateSelf(this.token, this.firstName, this.lastName, this.timezone)
        this.$router.push({ name: 'home' })
        this.loading = false
      } catch (err) {
        console.log(err)
        this.errorText = "We can't register you right now, try again later!"
        await firebase.auth().signOut();
        this.loading = false
      }
    },
  },
}
</script>

<style>

</style>
