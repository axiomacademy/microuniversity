<template>
  <div id="lessons" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col justify-start">
    <nav class="p-4 border-b border-gray-200 mb-3">
      <ul class="flex flex-row items-center justify-between">
        <h1 class="font-display text-3xl text-secondary font-medium">Lesson History</h1>
        <button class="w-12 h-12 rounded-lg text-secondary bg-purple-100" @click="goBack"><i class="fas fa-arrow-left"></i></button>
      </ul>
    </nav>
    
    <div v-if="!loading">
      <div v-if="lessons == null" class="shadow-sm bg-white rounded-md flex flex-col justify-center items-center p-6 mt-4">
        <img src="../assets/empty.png" class="w-24" />
        <h3 class="font-display text-lg text-text font-medium pt-4">No past lessons</h3>
        <h6 class="font-display text-sm text-text text-center mt-2">As you finish lessons, they'll appear here in case you want to take a look at them again :)</h6>
      </div>
      <div v-else v-for="lesson in lessons" :key="lesson.id" class="shadow-sm bg-white rounded-md flex flex-col py-6 my-3">
        <div class="flex flex-row justify-start items-center px-6">
          <h3 class="font-display text-lg font-medium text-text">{{ lesson.title }}</h3>
          <span v-if="!lesson.completed" class="font-body text-xs text-white ml-auto px-2 py-1 bg-red-400 uppercase rounded tracking-wide">overdue</span>
        </div>
        <span class="font-body text-sm text-gray-400 px-6 pb-4">{{ getFormattedDate(lesson.scheduled_date) }}</span>
        <!-- Image preview -->
        <iframe class="w-full h-52" allowFullScreen="allowFullScreen" frameBorder="0"
          :src="lesson.video_link">
        </iframe>
        <span class="font-body text-sm font-light text-text px-6 pt-4">
          {{ lesson.description }}
        </span>
        
        <button v-if="!lesson.completed" @click="completeAndReviewCards(lesson.id)" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mx-6 mt-4 rounded">Complete</button>
        <button v-else @click="reviewCards(lesson.id)" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mx-6 mt-4 rounded">Review</button>
      </div>
    </div>
    <div v-else class="flex-grow flex flex-col justify-center items-center">
      <MoonLoader class="self-center" color="#7938D8"/>
    </div>
  </div> 
</template>

<script>
import { MoonLoader } from '@saeris/vue-spinners'
import { getLessonsPast, getLessonFlashcards, completeLesson } from '../services/LessonService.js'

const options = {
     year: "numeric",
     month:"short",
     day:"2-digit"
}

export default {
  name: 'Lessons',
  components: {
    MoonLoader,
  },
  data: function () {
    return {
      loading: true,
      token: null,
      lessons: Array,
    }
  },
  created: async function () {
    this.loading = true
    // # Check if the JWT exists
    let token = localStorage.getItem("token")
    if(token == null) { 
      this.$router.push({ path: '/login' })
    }
    this.token = token

    // Get the past lessons
    this.lessons = await getLessonsPast(this.token)

    // Sorting the lessons
    this.lessons.sort((a,b) => {
      let d1 = new Date(a.scheduled_date)
      let d2 = new Date(b.scheduled_date)

      return d2 - d1
    })

    this.loading = false
  },
  methods: {
    goBack: function() {
      this.$router.go(-1)
    },
    getFormattedDate: function(dateString) {
      return new Date(dateString).toLocaleDateString("en-US", options)
    },
    reviewCards: async function(id) {
      let lessonFlashcards = await getLessonFlashcards(this.token, id)
      this.$router.push({ name: 'review', params: { title: 'Lesson Review', reviewCards: lessonFlashcards, done: async function() {
        return
      }}}) 
    },
    completeAndReviewCards: async function(id) {
      // Get the lesson flashcards
      await completeLesson(this.token, id)

      let lessonFlashcards = await getLessonFlashcards(this.token, id)
      this.$router.push({ name: 'review', params: { title: 'Lesson Review', reviewCards: lessonFlashcards, done: async function() {
        return
      }}}) 
    },
  }
}
</script>

<style>

</style>
