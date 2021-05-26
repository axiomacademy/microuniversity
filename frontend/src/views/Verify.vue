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

import { verifyOtp } from '../services/LoginService.js'

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
    let code = this.$route.query.code
    let email = this.$route.query.email

    if((code == "") || (email == "")) {
      this.errorText = "You clicked an invalid magic link. Try again!"
      return
    }
    
    if((code == null) || (email == null)) {
      this.errorText = "You clicked an invalid magic link. Try again!"
      return
    }

    // Verify the authenticity of the code and get jwt
    try {
      let response = await verifyOtp(email, code)

      if (response.jwt != null) {
          // Set the jwt to the localstorage and route to the home page
          localStorage.setItem("token", response.jwt)
          this.$router.push({ name: 'home'})
        } else {
          this.errorText = "We can't log you in right now. Try again later :("
        }
    } catch (err) {
      if(err == 401) {
        //Unauthorised
        this.errorText = "Invalid OTP. Please try again"
      } else {
        console.log(err)
        this.errorText = "We can't log you in right now. Try again later :("
      }
    }
  },
}
</script>

<style>

</style>
