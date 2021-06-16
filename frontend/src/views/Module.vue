<template>
  <div id="module" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col justify-start items-center">
  <nav class="p-4 border-b border-purple-200 lg:w-6/12 w-full">
    <ul class="flex flex-row items-center justify-between">
      <img class="w-12 h-12 rounded-lg shadow-sm" :src="module.image" />
      <h1 class="font-display text-xl text-secondary font-medium">{{ module.id }}</h1>
      <button class="w-12 h-12 rounded-lg text-secondary bg-purple-100" @click="$router.go(-1)"><i class="fas fa-arrow-left"></i></button>
    </ul>
  </nav>

  <div class="w-full lg:w-6/12">
  <h1 class="font-display text-2xl text-secondary font-medium px-4 mt-6">{{ module.title }}</h1>
  <h2 class="font-regular text-sm text-text mt-2 px-4">{{ module.description }}</h2>
  <div class="w-full self-center bg-purple-200 my-6" style="height: 1px;"></div>

    <div class="w-full" v-if="enrolled">
      <h1 v-if="!loading" class="font-display text-2xl text-secondary font-medium px-4">Enrollment Status</h1>
      <div v-if="!loading" class="bg-white w-full shadow-sm rounded-md flex flex-col justify-start items-start px-6 py-4 mt-4">
        <h2 class="font-display text-md text-text font-medium">{{ readableStatus(existingCohort.status) }}</h2>
        <h2 class="font-display text-sm text-gray-400">Tutorial session every {{ readableDay(existingCohort.tutorial_day, existingCohort.tutorial_time) }}, {{ readableTime(existingCohort.tutorial_day, existingCohort.tutorial_time) }}</h2>

        <h3 v-if="existingCohort.status == 0" class="font-body text-sm mt-2">{{ existingCohort.learner_count }}/15 learners enrolled <span class="text-primary">( {{ Math.round(existingCohort.learner_count/15 * 100) }}% )</span></h3>
        
        <h3 v-if="existingCohort.status == 1" class="font-body text-sm mt-2">We'll be setting your module start date shortly</h3>
        
        <h3 v-if="existingCohort.status == 2" class="font-body text-sm mt-2">Starting on <span class="text-primary">{{ readableDateTime(existingCohort.start_date) }}</span></h3>
         
        <span v-if="errorTextDeenroll != ''" class="text-red-500 my-3 font-body text-xs">{{ errorTextDeenroll }}</span>
        
        <button v-if="existingCohort.status == 0" @click="leaveCohort" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-8 w-full rounded flex flex-row justify-center items-center">
          <BeatLoader :size="8.5" color="#ffffff" v-if="loadingLeave" />
          <div v-else>
            De-enroll
          </div>
        </button> 
      </div>
    </div>

    <!-- Cohort booking and schedulinng -->
    <div class="w-full" v-if="!enrolled">
      <h1 v-if="!loading" class="font-display text-2xl text-secondary font-medium px-4">Available Cohorts</h1>
      <div v-if="showAvailability && !loading" class="w-full px-4">
        <h2 class="font-regular text-xs text-text font-light mt-1 mb-4">Every Axiom module comes with a weekly hour-long tutorial session. Pick the timeslot you are comfortable with and we'll enroll you in a cohort that works for you!</h2>
        <div class="flex flex-row justify-start items-center mt-2">
          <Chip class="mr-1 cursor-pointer" :focused="showDay == 0" @click.native="showDay = 0"> Friday </Chip>
          <Chip class="mx-1 cursor-pointer" :focused="showDay == 1" @click.native="showDay = 1"> Saturday </Chip>
          <Chip class="mx-1 cursor-pointer" :focused="showDay == 2" @click.native="showDay = 2"> Sunday </Chip>
        </div>

        <div v-if="showDay == 0" class="grid grid-flow-row grid-cols-3 gap-2 mt-6">
          <button v-for="slot in friTimeslots" :key="slot.datetime.getTime()" @click="selectedCohort = slot.cohort" class="relative shadow-sm rounded-md h-full w-full py-2 text-xs flex flex-col justify-center items-center" :class="timeClassObject(slot.cohort)">
            <span>{{ getFormattedTime(slot.datetime) }}</span>
            <span class="text-xs font-light">{{ Math.round(slot.learner_count/15 * 100) }}% full</span>
          </button>
        </div>
        <div v-if="showDay == 1" class="grid grid-flow-row grid-cols-3 gap-2 mt-6">
          <button v-for="slot in satTimeslots" :key="slot.datetime.getTime()" @click="selectedCohort = slot.cohort" class="shadow-sm rounded-md h-full w-full py-2 text-xs flex flex-col justify-center items-center" :class="timeClassObject(slot.cohort)">
            {{ getFormattedTime(slot.datetime) }}
            <span class="text-xs font-light">{{ Math.round(slot.learner_count/15 * 100) }}% full</span>
          </button>
        </div>
        <div v-if="showDay == 2" class="grid grid-flow-row grid-cols-3 gap-2 mt-6">
          <button v-for="slot in sunTimeslots" :key="slot.datetime.getTime()" @click="selectedCohort = slot.cohort" class="shadow-sm rounded-md h-full w-full py-2 text-xs flex flex-col justify-center items-center" :class="timeClassObject(slot.cohort)">
            {{ getFormattedTime(slot.datetime) }}
            <span class="text-xs font-light">{{ Math.round(slot.learner_count/15 * 100) }}% full</span>
          </button>
        </div>
        
        <span v-if="errorText != ''" class="text-red-500 font-body text-xs">{{ errorText }}</span>

        <button @click="enrollCohort" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-8 w-full rounded flex flex-col justify-center items-center">
          <BeatLoader :size="8.5" color="#ffffff" v-if="loadingEnroll" />
          <div v-else>
            Enroll
          </div>
        </button> 
      </div> 
      <div v-if="!showAvailability && !loading" class="bg-purple-200 font-display font-light text-sm text-secondary py-2 px-6 rounded flex mt-4 mx-4">{{ fillerText }}</div> 
    </div>

    <div v-if="loading" class="flex-grow flex flex-col justify-start items-center mt-8">
      <MoonLoader class="self-center" color="#7938D8"/>
    </div>
  </div>
  </div> 
</template>

<script>
import { BeatLoader } from '@saeris/vue-spinners'
import { MoonLoader } from '@saeris/vue-spinners'
import Chip from '../components/Chip.vue'

import { getAvailableCohorts, joinCohort, getSelfActiveCohort, leaveModuleCohort } from '../services/CohortService'
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
      existingCohort: {},
      loadingEnroll: false,
      loadingLeave: false,
      loading: false,
      optionsTime: {
        hour12 : true,
        hour:  "numeric",
        minute: "numeric",
        seconds:"numeric"
      },
      optionsDay: {
        weekday: "long",
      },
      friTimeslots: [],
      satTimeslots: [],
      sunTimeslots: [],
      selectedCohort: "",
      showDay: 0,
      module: {},
      errorText: "",
      errorTextDeenroll: "",
      showAvailability: true,
      fillerText: "",
      enrolled: false,
    }
  },
  created() {
    this.module = this.$route.params.module
    this.loading = true

    firebase.auth().onAuthStateChanged(async (user) => {
      if (user) {
        // Get the available cohort
        this.token = await user.getIdToken(true) 

        // Check if they are already enrolled in a cohort for this module
        this.existingCohort = await getSelfActiveCohort(this.token)

        if (this.existingCohort != null) {
          // Either you're enrolled in this module or another module
          if (this.existingCohort.module == this.module.id) {
            this.enrolled = true
            this.loading = false
            this.showAvailability = false
            return
          } else {
            this.fillerText = `You can only be enrolled in one module at a time. Currently you're enrolled in ${this.existingCohort.module}`
            this.enrolled = false
            this.showAvailability = false
            this.loading = false
            return
          }
        }

        await this.getModuleAvailableCohorts()

        this.enrolled = false
        this.loading = false
      } else {
        this.$router.push({ name: 'login' })
        this.loading = false
      }
    })
  },
  computed: {
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
        this.existingCohort = await getSelfActiveCohort(this.token)
        if (this.existingCohort != null) {
          this.enrolled = true
        } else {
          this.errorText = "We can't enroll you right now. Try again later :("
        }

        this.loadingEnroll = false
      } catch (err) {
        console.log(err)
        this.errorText = "We can't enroll you right now. Try again later :("
        this.loadingEnroll = false
      }
    },
    async leaveCohort() {
      this.loadingLeave = true
      try {
        await leaveModuleCohort(this.token, this.module.id) 
        this.loading = true
        await this.getModuleAvailableCohorts()
        this.enrolled = false
        this.loadingLeave = false
        this.loading = false
      } catch (err) {
        console.log(err)
        this.errorTextDeenroll = "We can't de-enroll you right now. Try again later :("
        this.loadingLeave = false
      }
    },
    readableStatus(status) {
      switch(status) {
        case 0:
          return "Cohort unconfirmed"
        case 1:
          return "Cohort confirmed ✨"
        case 2:
          return "Cohort ongoing! Keep at it 🔥"
        case 3:
          return "Cohort completed ✅"
        default:
          return ""
      }
    },
    readableDay(day, time) {
      let date = new Date(Date.UTC(0, 0, day + 1, 0, time, 0, 0))
      return Intl.DateTimeFormat('en-US', this.optionsDay).format(date)
    },
    readableTime(day, time) {
      let date = new Date(Date.UTC(0, 0, day + 1, 0, time, 0, 0))
      return date.toLocaleTimeString("en-US", this.optionsTime)
    },
    readableDateTime(date) {
      let properDate = new Date(date)
      return Intl.DateTimeFormat('en-US', {
        weekday: 'long', year: 'numeric', month: 'long', day: 'numeric'}).format(properDate)
    },
    async getModuleAvailableCohorts() {
      this.friTimeslots = []
      this.satTimeslots = []
      this.sunTimeslots = []

      let availableCohorts = await getAvailableCohorts(this.token, this.module.id)

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
            this.friTimeslots.push({datetime: time, cohort: cohort.id, learner_count: cohort.learner_count})
          } else if (time.getDay() == 6) {
            this.satTimeslots.push({datetime: time, cohort: cohort.id, learner_count: cohort.learner_count})
          } else if (time.getDay() == 0) {
            this.sunTimeslots.push({datetime: time, cohort:cohort.id, learner_count: cohort.learner_count})
          }
        }

        this.showAvailability = true
      }
    },
  }
}
</script>

<style>

</style>
