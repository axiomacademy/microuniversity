<template>
  <div id="module" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col justify-start">
  <nav class="p-4 border-b border-purple-200">
    <ul class="flex flex-row items-center justify-between">
      <img class="w-12 h-12 rounded-lg shadow-sm" :src="module.image" />
      <h1 class="font-display text-xl text-secondary font-medium">{{ module.id }}</h1>
      <button class="w-12 h-12 rounded-lg text-secondary bg-purple-100" @click="$router.go(-1)"><i class="fas fa-arrow-left"></i></button>
    </ul>
  </nav>
  <h1 class="font-display text-2xl text-secondary font-medium px-4 mt-6">{{ module.title }}</h1>
  <h2 class="font-regular text-sm text-text mt-2 px-4">{{ module.description }}</h2>

  <div class="w-full self-center bg-purple-200 my-6" style="height: 1px;"></div>
    <!-- Cohort booking and schedulinng -->
    <h1 class="font-display text-2xl text-secondary font-medium px-4">Available Cohorts</h1>

    <div v-if="showAvailability && !loading" class="w-full px-4">
      <h2 class="font-regular text-xs text-text font-light mt-1 mb-4">Every Axiom module comes with a weekly hour-long tutorial session. Pick the timeslot you are comfortable with and we'll enroll you in a cohort that works for you!</h2>
      <div class="flex flex-row justify-start items-center mt-2">
        <Chip class="mr-1" :focused="showDay == 0" @click.native="showDay = 0"> Friday </Chip>
        <Chip class="mx-1" :focused="showDay == 1" @click.native="showDay = 1"> Saturday </Chip>
        <Chip class="mx-1" :focused="showDay == 2" @click.native="showDay = 2"> Sunday </Chip>
      </div>

      <div v-if="showDay == 0" class="grid grid-flow-row grid-cols-3 gap-2 mt-6">
        <button v-for="slot in friTimeslots" :key="slot.datetime.getTime()" @click="selectedCohort = slot.cohort" class="shadow-sm rounded-md h-full w-full py-2 text-xs flex justify-center items-center" :class="timeClassObject(slot.cohort)">
          {{ getFormattedTime(slot.datetime) }}
        </button>
      </div>
      <div v-if="showDay == 1" class="grid grid-flow-row grid-cols-3 gap-2 mt-6">
        <button v-for="slot in satTimeslots" :key="slot.datetime.getTime()" @click="selectedCohort = slot.cohort" class="shadow-sm rounded-md h-full w-full py-2 text-xs flex justify-center items-center" :class="timeClassObject(slot.cohort)">
          {{ getFormattedTime(slot.datetime) }}
        </button>
      </div>
      <div v-if="showDay == 2" class="grid grid-flow-row grid-cols-3 gap-2 mt-6">
        <button v-for="slot in sunTimeslots" :key="slot.datetime.getTime()" @click="selectedCohort = slot.cohort" class="shadow-sm rounded-md h-full w-full py-2 text-xs flex justify-center items-center" :class="timeClassObbject(slot.cohort)">
          {{ getFormattedTime(slot.datetime) }}
        </button>
      </div>
      
      <span v-if="errorText != ''" class="text-red-500 my-3  font-body text-xs px-4">{{ errorText }}</span>

      <button @click="enrollCohort" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-8 w-full rounded flex flex-row justify-center items-center">
        <BeatLoader :size="8.5" color="#ffffff" v-if="loadingEnroll" />
        <div v-else>
          Enroll
        </div>
      </button> 
    </div> 

    <div v-if="!showAvailability && !loading" class="bg-purple-200 font-display font-light text-secondary py-2 px-6 rounded flex mt-4 mx-4">{{ fillerText }}</div>
    <div v-if="loading" class="flex-grow flex flex-col justify-start items-center mt-8">
      <MoonLoader class="self-center" color="#7938D8"/>
    </div>

  </div> 
</template>

<script>
import { BeatLoader } from '@saeris/vue-spinners'
import { MoonLoader } from '@saeris/vue-spinners'
import Chip from '../components/Chip.vue'

import { getAvailableCohorts, joinCohort } from '../services/CohortService'
import firebase from "firebase/app";
import "firebase/auth";

export default {
  name: "Module",
  components: {
    Chip,
    BeatLoader,
    MoonLoader
  },
  data: function () {
    return {
      loadingEnroll: false,
      loading: false,
      optionsTime: {
        hour12 : true,
        hour:  "numeric",
        minute: "numeric",
        seconds:"numeric"
      },
      friTimeslots: [],
      satTimeslots: [],
      sunTimeslots: [],
      selectedCohort: "",
      showDay: 0,
      module: {},
      errorText: "",
      showAvailability: true,
      fillerText: "",
    }
  },
  created() {
    this.module = this.$route.params.module
    this.loading =true

    firebase.auth().onAuthStateChanged(async (user) => {
      if (user) {
        // Get the available cohort
        this.token = await user.getIdToken(true) 

        let availableCohorts = await getAvailableCohorts(this.token, this.module.id)
        console.log(availableCohorts) 

        if(availableCohorts == null) {
          this.fillerText = "You are already enrolled in a cohort for this module 👍"
          this.showAvailability = false
        } else if(availableCohorts.length == 0) {
          this.fillerText = "We do not have any cohorts open at the moment 😔"
          this.showAvailability = false
        } else {
          // Converting cohorts into clicks
          for (let cohort of availableCohorts) {
            // Convert to local time
            let time = new Date(Date.UTC(0, 0, cohort.tutorial_day + 1, 0, cohort.tutorial_time, 0, 0))
            console.log(time)
            if (time.getDay() == 5) {
              // It's a friday
              this.friTimeslots.push({datetime: time, cohort: cohort.id})
            } else if (time.getDay() == 6) {
              this.satTimeslots.push({datetime: time, cohort: cohort.id})
            } else if (time.getDay() == 7) {
              this.sunTimeslots.push({datetime: time, cohort:cohort.id})
            }
          }
        }

        this.loading = false
      } else {
        this.$router.push({ name: 'login' })
        this.loading = false
      }
    })
  },
  methods: {
    getFormattedTime(time) {
      return time.toLocaleTimeString("en-US", this.optionsTime)
    },
    timeClassObject: function (cohort) {
      return {
        'bg-white': cohort != this.selectedCohort,
        'bg-primary': cohort == this.selectedCohort,
        'text-text': cohort != this.selectedCohort,
        'text-white': cohort == this.selectedCohort,
      }
    },
    async enrollCohort() {
      this.loadingEnroll = true
      if (this.selectedCohort == "") {
        this.errorText = "Please select a tutorial slot above"
        this.loadingEnroll = false
        return
      }
      try {
        await joinCohort(this.token, this.selectedCohort)

        this.$router.go(-1)
        this.loadingEnroll = false
      } catch (err) {
        console.log(err)
        this.errorText = "We can't enroll you right now. Try again later :("
        this.loadingEnroll = false
      }
    }
  }
}
</script>

<style>

</style>
