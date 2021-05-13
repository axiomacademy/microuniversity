<template>
  <div id="home" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col">
    <nav class="p-4 border-b border-gray-200">
      <ul class="flex flex-row items-center">
        <img src="../assets/logo-transparent-dark.png" class="w-14 h-14"/>
        <div class="text-secondary bg-purple-100 px-4 h-12 flex justify-center items-center rounded-full ml-auto mr-4">
          🔥 {{ streak }}
        </div>
        <button @click="logout" class="w-12 h-12 rounded-lg text-secondary bg-purple-100"><i class="fas fa-sign-out-alt"></i></button>
      </ul>
    </nav>
      
    <div v-if="!loading">
      <DailyReviewCard :dailyReviewCards="this.dailyReviewCards" class="mt-4" />
      <!-- Daily lesson section -->
      <h1 class="font-display text-2xl text-secondary pl-4 mt-8 font-normal">Today's Lesson</h1> 
      <DailyLessonCard class="mt-4" :todayLesson="this.todayLesson" /> 
      <button class="bg-purple-200 w-full font-display font-light text-secondary py-2 px-6 rounded flex mt-4" @click="$router.push({ name: 'lessons' })">View previous lessons...</button>
      <h1 class="font-display text-2xl text-secondary pl-4 mt-8 font-normal mb-2">Upcoming Tutorials</h1>

      <!-- Generating Tutorial List -->
      <div v-if="upcomingTutorials == null" class="shadow-sm bg-white rounded-md flex flex-col justify-center items-center py-6 mt-4">
        <img src="../assets/relax.png" class="w-24" />
        <h3 class="font-display text-lg text-text font-medium pt-4">No upcoming tutorials</h3>
        <h6 class="font-display text-md text-text">Looks like you can relax for a while!</h6>
      </div>
      <TutorialListElement v-else v-for="tutorial in upcomingTutorials" :key="tutorial.id" :title="tutorial.title" :datetime="tutorial.scheduled_time" class="mt-2" />
    </div>
    <div v-else class="flex-grow flex flex-col justify-center items-center">
      <MoonLoader class="self-center" color="#7938D8"/>
    </div>
  </div>
</template>

<script>
import DailyReviewCard from '../components/DailyReviewCard.vue'
import DailyLessonCard from '../components/DailyLessonCard.vue'
import TutorialListElement from '../components/TutorialListElement.vue'
import { MoonLoader } from '@saeris/vue-spinners'


// Services
import { getSelf } from '../services/LearnerService.js'
import { getLessonToday } from '../services/LessonService.js'
import { getUpcomingTutorials } from '../services/TutorialService.js'
import { getDailyReview } from '../services/ReviewService.js'

export default {
  name: 'App',
  components: {
    DailyReviewCard,
    DailyLessonCard,
    TutorialListElement,
    MoonLoader,
  },
  data: function () {
    return {
      loading: true,
      token: "",
      learnerId: "",
      streak: 0,
      todayLesson: null,
      upcomingTutorials: null,
      dailyReviewCards: null,
    }
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

    // Get today's lesson if any
    this.todayLesson = await getLessonToday(this.token)

    // Get all the upcoming tutorials
    this.upcomingTutorials = await getUpcomingTutorials(this.token)

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
