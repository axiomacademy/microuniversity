<template>
  <div id="verify" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col justify-center items-center">
      <MoonLoader v-if="errorText == ''" class="self-center" color="#7938D8"/>
      <div v-else class="shadow-sm px-10 py-8 bg-white rounded-md flex flex-col justify-center items-start">
        <img src="../assets/error.png" class="h-40 mb-4 self-center"/>
        <h1 class="font-display text-xl text-text font-medium">Oh no!</h1>
        <h2 class="font-display text-sm text-gray-600 font-regular py-2">{{ errorText }}</h2>

        <button @click="$router.push({ name: 'login' })" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-4 rounded w-full flex flex-row justify-center items-center">
          <BeatLoader :size="8.5" color="#ffffff" v-if="loading" />
          <div v-else>
            Back to login
          </div>
        </button>
      </div>
  </div>  
</template>

<script>
import { MoonLoader } from '@saeris/vue-spinners'
import { getSelf } from '../services/LearnerService'
import firebase from "firebase/app";
import "firebase/auth";

export default {
  name: 'Verify',
  components: {
    MoonLoader,
  },
  data: function () {
    return {
      errorText: "",
    }
  },
  mounted: async function () {
    console.log("Verify page reached")

    if (firebase.auth().isSignInWithEmailLink(window.location.href)) {
      let email = window.localStorage.getItem('emailForSignIn');
      if (!email) {
        // User opened link on a different device, need access to email
        email = window.prompt('Please provide your email for confirmation');
      }

      try {
        await firebase.auth().signInWithEmailLink(email, window.location.href)
        // Clear email from storage.
        window.localStorage.removeItem('emailForSignIn');
        
        // CHeck if they're a new user
        let token = await firebase.auth().currentUser.getIdToken(true)
        console.log(token)
        let self = await getSelf(token)
        
        if (self.first_name == "" || self.last_name == "") {
          this.$router.push({ name: 'register' })
        } else {
          this.$router.push({ name: 'home'})
        }
      } catch (err) {
        console.log(err)
        this.errorText = "We can't log you in right now. Try again later :("
        return
      }
    } else {
      this.errorText = "You clicked an invalid magic link. Try again!"
      return
    }
  },
}
</script>

<style>

</style>
