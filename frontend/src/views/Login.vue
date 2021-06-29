<template>
  <div id="login" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col justify-center items-center">
    <div v-if="!verificationSent" class="shadow-sm px-10 py-8 bg-white rounded-md flex flex-col justify-center items-start">
      <img src="../assets/logo-transparent-dark.png" class="w-20 h-20 mb-4"/>
      <h1 class="font-display text-3xl text-text font-medium">Login to <span class="text-primary">Axiom</span></h1>
      <h2 class="font-display text-sm text-gray-600 font-regular py-2">Enter your email to get to changing the way you think</h2>
      <input v-model="email" type="text" placeholder="Email" class="bg-gray-100 p-2 w-full rounded font-display border border-transparent focus:outline-none focus:ring-2 focus:ring-purple-600 focus:border-transparent mt-4" :disabled="loadingEmail">
      <span v-if="errorText != ''" class="text-red-500 my-3 font-body text-xs">{{ errorText }}</span>

      <button @click="loginWithEmail" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-4 rounded w-full flex flex-row justify-center items-center">
        <BeatLoader :size="8.5" color="#ffffff" v-if="loadingEmail" />
        <div v-else>
          Sign In
        </div>
      </button> 

      <div class="flex flex-row w-full items-center my-2">
        <div class="flex-grow bg-gray-100" style="height: 1px;"></div>
        <span class="px-2 text-gray-500 upppercase text-sm"> OR </span>
        <div class="flex-grow bg-gray-100" style="height: 1px;"></div>
      </div>
      
      <button @click="loginWithGoogle" class="bg-gray-100 tracking-widest font-body text-xs text-medium text-gray-500 uppercase p-2 rounded w-full flex flex-row justify-center items-center">
        <BeatLoader :size="8.5" color="#D1D5DB" v-if="loadingGoogle" />
        <div v-else class= "flex flex-row justify-left w-full">
          <img src="../assets/g-icon.png" class="w-4" />
          <span class="mx-auto">Sign In With Google</span>
        </div>
      </button> 
    </div>
    <div v-else class="shadow-sm px-10 py-8 bg-white rounded-md flex flex-col justify-center items-start">
      <img src="../assets/mail-sent.png" class="h-20 mb-4 self-center"/>
      <h1 class="font-display text-xl text-text font-medium">Magic link sent!</h1>
      <h2 class="text-sm text-gray-600 font-regular py-2">We've sent a magic link to <span class="font-semibold">{{ email }}</span>. Click it from this device to login and start learning 😎</h2>

      <button @click="verificationSent = !verificationSent" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-4 rounded w-full flex flex-row justify-center items-center">
        <div>
          Back to login
        </div>
      </button>
    </div>
  </div> 
</template>

<script>
import { BeatLoader } from '@saeris/vue-spinners'

import { getSelf } from '../services/LearnerService'

import firebase from "firebase/app";
import "firebase/auth";
import "firebase/functions";

let googleProvider = new firebase.auth.GoogleAuthProvider();

// Imported firebase function
const addUserClaim = firebase.functions().httpsCallable('addUserClaim');

// Settinng up email configuration
let actionCodeSettings = {
  // URL you want to redirect back to. The domain (www.example.com) for this
  // URL must be in the authorized domains list in the Firebase Console.
  url: 'https://betav2.axiom.academy/#/verify',
  // This must be true.
  handleCodeInApp: true,
};

export default {
  name: 'Login',
  components: {
    BeatLoader,
  },
  data: function() {
    return {
      loadingEmail: false,
      loadingGoogle: false,
      errorText: "",
      email: "",
      verificationSent: false,
      unsubAuth: null,
    }
  },
  created: async function () {
    // # Check if logged in firebase
    this.unsubAuth = firebase.auth().onAuthStateChanged(async (user) => {
      console.log("Login")
      if (user) {
        addUserClaim({email: user.email})
        let token = await user.getIdToken()

        // Store the token in localStorage
        localStorage.setItem("FB_TOKEN", token)
        localStorage.setItem("EMAIL", user.email)
        
        // CHeck if they're a new user
        let res = await getSelf(token, user.email)

        if (res.data.getLearner == null) {
          // Go to the account creation page
          this.$router.push({ name: 'register' })
        } else {
          this.$router.push({ name: 'home'})
        }
      }
    })
  },
  beforeDestroy() {
    this.unsubAuth()
  },
  methods: {
    loginWithEmail: async function () {
      // Attempt to login
      this.loadingEmail = true
      try {
        await firebase.auth().sendSignInLinkToEmail(this.email, actionCodeSettings)
        window.localStorage.setItem('emailForSignIn', this.email)

        this.verificationSent = true
        this.loadingEmail = false
      } catch (err) {
        console.log(err)
        this.errorText = "We can't log you in right now. Try again later :("
        this.loadingEmail = false
      }
      return
    },
    loginWithGoogle: async function () {
      this.loadingGoogle = true
      try {
        await firebase.auth().signInWithPopup(googleProvider) 
      } catch (err) {
        console.log(err)
        this.errorText = "We can't log you in right now. Try again later :("
        this.loadingGoogle = false
      }
      return
    }
  }
}
</script>

<style>

</style>
