<template>
  <div id="home" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col">
    <!-- Nav bar -->
    <nav class="p-4 border-b border-purple-200">
      <ul class="flex flex-row items-center">
        <img src="../assets/logo-transparent-dark.png" class="w-14 h-14"/>
        <div class="text-secondary bg-purple-100 px-4 h-12 flex justify-center items-center rounded-full ml-auto mr-4">
          🔥 {{ streak }}
        </div>
        <button @click="logout" class="w-12 h-12 rounded-lg text-secondary bg-purple-100"><i class="fas fa-bars"></i></button>
      </ul>
    </nav>
      
    <div v-if="!loading">
      <!-- Explore tab with modules, enrolled cohorts etc. -->
      <div v-id="openTab == 'Explore'" id="explore-tab">
      </div>

      <!-- Learning tab with daily review, lessons and tutorials -->
      <div v-if="openTab == 'Learn'" id="learn-tab">
        <DailyReviewCard :dailyReviewCards="this.dailyReviewCards" class="mt-4" />
        <!-- Daily lecture section -->
        <h1 class="font-display text-2xl text-secondary pl-4 mt-8">Today's Lecture</h1> 
        <DailyLectureCard class="mt-4" :todayLecture="this.todayLecture" /> 
        <button class="bg-purple-200 w-full font-display font-light text-secondary py-2 px-6 rounded flex mt-4" @click="$router.push({ name: 'lectures' })">View previous lectures...</button>
        <h1 class="font-display text-2xl text-secondary pl-4 mt-8 font-normal mb-2">Upcoming Tutorials</h1>

        <!-- Generating Tutorial List -->
        <div v-if="upcomingTutorials == null" class="shadow-sm bg-white rounded-md flex flex-col justify-center items-center py-6 mt-4">
          <img src="../assets/relax.png" class="w-24" />
          <h3 class="font-display text-lg text-text font-medium pt-4">No upcoming tutorials</h3>
          <h6 class="font-display text-md text-text">Looks like you can relax for a while!</h6>
        </div>
        <TutorialListElement v-else v-for="tutorial in upcomingTutorials" :key="tutorial.id" :title="tutorial.title" :datetime="tutorial.scheduled_time" class="mt-2" />
      </div>
    </div>

    <!-- LOADING INDICATOR -->
    <div v-else class="flex-grow flex flex-col justify-center items-center">
      <MoonLoader class="self-center" color="#7938D8"/>
    </div>

    <!-- Floating tab buttons -->
    <div class="fixed inset-x-0 bottom-0 mb-8 bg-white shadow-lg mx-auto flex w-9/12 h-16 rounded-full justify-around overflow-hidden">
      <button class="w-6/12 font-display focus:outline-none" @click="openTab = 'Explore'"
        v-bind:class="{ 'bg-purple-100': exploreTabOpen, 'text-secondary': exploreTabOpen, 'text-text': exploreTabOpen}">
        🌎 <span class="pl-1" v-bind:class="{ 'font-medium': exploreTabOpen }">Explore</span>
      </button>
      <div class="h-full bg-purple-200" style="width: 1px;"></div>
      <button class="w-6/12 font-display focus:outline-none" @click="openTab = 'Learn'" 
        v-bind:class="{ 'bg-purple-100': learnTabOpen, 'text-secondary': learnTabOpen, 'text-text': learnTabOpen }">
        📖 <span class="pl-1" v-bind:class="{ 'font-medium': learnTabOpen }">Learn</span>
      </button>
    </div>
  </div>
</template>

<script>
import DailyReviewCard from '../components/DailyReviewCard.vue'
import DailyLectureCard from '../components/DailyLectureCard.vue'
import TutorialListElement from '../components/TutorialListElement.vue'
import { MoonLoader } from '@saeris/vue-spinners'


// Services
import { getSelf } from '../services/LearnerService.js'
import { getLectureToday } from '../services/LectureService.js'
import { getUpcomingTutorials } from '../services/TutorialService.js'
import { getDailyReview } from '../services/ReviewService.js'

export default {
  name: 'App',
  components: {
    DailyReviewCard,
    DailyLectureCard,
    TutorialListElement,
    MoonLoader,
  },
  data: function () {
    return {
      loading: true,
      token: "",
      learnerId: "",
      streak: 0,
      todayLecture: null,
      upcomingTutorials: null,
      dailyReviewCards: null,
      openTab: "Learn",
    }
  },
  computed: {
    learnTabOpen: function () {
      return this.openTab == "Learn"
    },
    exploreTabOpen: function () {
      return this.openTab == "Explore"
    },
  },
  created: async function () {
    this.loading = true

    // # Check if the JWT exists
    let token = localStorage.getItem("token")
    if(token == null) { 
      this.$router.push({ name: "login" })
    }

    this.token = token
    
    // Get all the important data
    let self = await getSelf(this.token)
    this.learnerId = self.id
    this.streak = self.streak

    // Get today's lecture if any
    this.todayLecture = await getLectureToday(this.token)

    // Get all the upcoming tutorials
    this.upcomingTutorials = await getUpcomingTutorials(this.token)
    if(this.upcomingTutorials != null) {
      this.upcomingTutorials.sort((a,b) => {
        let d1 = new Date(a.scheduled_time)
        let d2 = new Date(b.scheduled_time)

        return d1 - d2
      })
    }

    // Get the daily review if there is one
    this.dailyReviewCards = await getDailyReview(this.token)

    this.loading = false
  },
  methods: {
    logout: function() {
      localStorage.removeItem("token")
      this.$router.push({ name: "login" })
    },
  },
}
</script>

<style>

</style>
