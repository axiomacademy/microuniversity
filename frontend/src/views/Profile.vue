<template>
  <div id="profile" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col justify-start items-center">
    <nav class="p-4 border-b border-purple-200 mb-3 w-full lg:w-6/12">
      <ul class="flex flex-row items-center justify-between">
        <h1 class="font-display text-3xl text-secondary font-medium">Edit Profile</h1>
        <button @click="$router.go(-1)" class="w-12 h-12 rounded-lg text-secondary bg-purple-100"><i class="fas fa-arrow-left"></i></button>
      </ul>
    </nav>

    <div v-if="loading != true" class="m-4 w-full lg:w-6/12">
      <div class="w-20 h-20 rounded-lg flex justify-center items-center text-white text-3xl tracking-widest uppercase" :style="profileHash">{{ self.first_name[0] + self.last_name[0] }}</div>
      <div class="text-text text-lg font-body pt-4 font-semibold leading-snug">{{ self.first_name }} </div>
      <div class="text-text text-lg font-body leading-snug">{{ self.last_name }} </div>
      <div class="text-gray-600 text-md font-body leading-snug">{{ self.timezone }} </div>
      <div class="text-secondary bg-purple-100 px-4 h-10 w-20 mt-4 flex justify-center items-center rounded-full">
        🔥 {{ this.self.streak }}
      </div>
      
      <div class="bg-purple-200 w-full my-4" style="height: 1px;"></div>
      
      <input v-model="newFirstName" type="text" placeholder="First Name" class="bg-purple-100 p-2 w-full rounded font-display border border-transparent focus:outline-none focus:ring-2 focus:ring-purple-600 focus:border-transparent mt-4 placeholder-purple-300" :disabled="loading">
      <input v-model="newLastName" type="text" placeholder="Last Name" class="bg-purple-100 p-2 w-full rounded font-display border border-transparent focus:outline-none focus:ring-2 focus:ring-purple-600 focus:border-transparent mt-2 placeholder-purple-300" :disabled="loading">
      <span v-if="errorTextName != ''" class="text-red-500 my-3 font-body text-xs">{{ errorTextName }}</span>
      <button @click="updateName" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-4 rounded w-full flex flex-row justify-center items-center">
        <BeatLoader :size="8.5" color="#ffffff" v-if="loadingNameUpdate" />
        <div v-else>
          Update Name
        </div>
      </button> 
      
      <div class="bg-purple-200 w-full my-6" style="height: 1px;"></div>
     
      <div class="text-sm text-text">Your detected timezone is <span class="font-bold">{{ newTimezone }}</span>. The local time here is <span class="text-secondary">{{ currentTime }}.</span></div>
      <span v-if="errorTextTZ != ''" class="text-red-500 my-3 font-body text-xs">{{ errorTextTZ }}</span>
      <button @click="updateTZ" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-4 rounded w-full flex flex-row justify-center items-center">
        <BeatLoader :size="8.5" color="#ffffff" v-if="loadingTZUpdate" />
        <div v-else>
          Update Timezone
        </div>
      </button> 
    </div>
    
    <div v-else class="flex-grow flex flex-col justify-center items-center">
      <MoonLoader class="self-center" color="#7938D8"/>
    </div>

  </div>
  
</template>

<script>
import ColorHash from 'color-hash'
import { BeatLoader, MoonLoader } from '@saeris/vue-spinners'

import { getSelf, updateSelf } from '../services/LearnerService.js'

import firebase from "firebase/app";
import "firebase/auth";

const colorHash = new ColorHash({saturation: 0.5, lightness: 0.8})

export default {
  name: 'Profile',
  components: {
    BeatLoader,
    MoonLoader,
  },
  data: function() {
    return {
      self: {},
      newFirstName: "",
      newLastName: "",
      newTimezone: "",
      currentTime: "",
      loading: true,
      loadingNameUpdate: false,
      loadingTZUpdate: false,
      errorTextName: "",
      errorTextTZ: "",
    }
  },
  created () {
    this.loading = true
    firebase.auth().onAuthStateChanged(async (user) => {
      if (user) {
        this.token = await user.getIdToken(true)
        this.newTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone

        let opts = {
          hour: 'numeric',
          minute:'2-digit',
          timeZone: this.newTimezone,
          timeZoneName: 'long'
        }
        this.currentTime = new Date().toLocaleString('en', opts) 
        setInterval(function(){ this.currentTime = new Date().toLocaleString('en', opts) }, 1000); 

        // Retrieving old stuff
        this.self = await getSelf(this.token)
        this.loading = false
      } else {
        this.$router.push({ name: 'login' })
      }
    })
  },
  computed: {
    profileHash() {
      return {
        backgroundColor: colorHash.hex(this.self.first_name + this.self.last_name)
      }
    }
  },
  methods: {
    async updateName() {
      this.loadingNameUpdate = true
      this.errorTextName = ""

      if (this.newFirstName == "" || this.newLastName == "") {
        this.errorTextName = "Please fill in all the fields!"
        this.loadingNameUpdate = false
        return
      }
      try {
        await updateSelf(this.token, this.newFirstName, this.newLastName, this.self.time_zone)

        this.self.first_name = this.newFirstName
        this.self.last_name = this.newLastName
        this.newFirstName = ""
        this.newLastName = ""
        this.loadingNameUpdate = false
      } catch (err) {
        console.log(err)
        this.errorTextName = "We can't update your name right now, try again later!"
        this.newFirstName = ""
        this.newLastName = ""
        this.loadingNameUpdate = false
      } 
    },
    async updateTZ() {
      this.loadingTZUpdate = true
      this.errorTextTZ = ""

      try {
        await updateSelf(this.token, this.self.first_name, this.self.last_name, this.newTimezone)
        this.self.time_zone = this.newTimezone
        this.loadingTZUpdate = false
      } catch (err) {
        this.errorTextTZ = "We can't update your timezone right now, try again later!"
        this.loadingTZUpdate = false
      }
    }
  }
}
</script>

<style>

</style>
