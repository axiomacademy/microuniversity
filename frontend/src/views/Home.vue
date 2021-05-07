<template>
  <div id="home" class="wrapper bg-purple-50 min-h-screen p-4">
    <nav class="p-4 border-b border-gray-200">
      <ul class="flex flex-row items-center">
        <img src="../assets/logo-transparent-dark.png" class="w-14 h-14"/>
        <div class="text-secondary bg-purple-100 px-4 h-12 flex justify-center items-center rounded-full ml-auto mr-4">
          🔥 {{ streak }}
        </div>
        <button class="w-12 h-12 rounded-lg text-secondary bg-purple-100"><i class="fas fa-sign-out-alt"></i></button>
      </ul>
    </nav>
    
    <DailyReviewCard :dailyReviewCards="this.dailyReviewCards" class="mt-4" />
    <!-- Daily lesson section -->
    <h1 class="font-display text-2xl text-secondary pl-4 mt-8 font-normal">Today's Lesson</h1> 
    <DailyLessonCard class="mt-4" :todayLesson="this.todayLesson" /> 
    <button class="bg-purple-200 w-full font-display font-light text-secondary py-2 px-6 rounded flex mt-4" @click="$router.push({ name: 'lessons' })">View previous lessons...</button>
    <h1 class="font-display text-2xl text-secondary pl-4 mt-8 font-normal mb-2">Upcoming Tutorials</h1>

    <!-- Generating Tutorial List -->
    <TutorialListElement v-for="tutorial in this.upcomingTutorials" :key="tutorial.id" :title="tutorial.title" :datetime="tutorial.scheduled_time" class="mt-2" />
  </div>
</template>

<script>
import DailyReviewCard from '../components/DailyReviewCard.vue'
import DailyLessonCard from '../components/DailyLessonCard.vue'
import TutorialListElement from '../components/TutorialListElement.vue'

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
  },
  data: function () {
    return {
      token: "",
      learnerId: "",
      streak: 0,
      todayLesson: null,
      upcomingTutorials: null,
      dailyReviewCards: null,
    }
  },
  created: async function () {
    // # Check if the JWT exists
    let token = localStorage.getItem("token")
    if(token == null) { 
      this.$router.push({ path: '/login' })
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
  },

}
</script>

<style>

</style>
