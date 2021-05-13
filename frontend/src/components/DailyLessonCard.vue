<template>
  <div>
    <div v-if="this.showFiller" class="shadow-sm bg-white rounded-md flex flex-col justify-center items-center py-6">
      <img src="../assets/done.png" class="w-16 h-16" />
      <h3 class="font-display text-lg text-text font-medium pt-4">No daily lesson</h3>
      <h6 class="font-display text-md text-text">You're done for the day!</h6>
    </div>
    <div v-else class="shadow-sm bg-white rounded-md flex flex-col py-6">
      <h3 class="font-display text-lg font-medium text-text px-6">{{ todayLesson.title }}</h3>
      <span class="font-body text-sm text-gray-400 px-6 pb-4">{{ todayLesson.module }}</span>
      <!-- Image preview -->
      <iframe class="w-full h-52" allowFullScreen="allowFullScreen" frameBorder="0"
        :src="todayLesson.video_link">
      </iframe>
      <span class="font-body text-sm font-light text-text px-6 pt-4">
        {{ todayLesson.description }}
      </span>
      
      <button @click="markLessonAsComplete" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mx-6 mt-4 rounded">Complete</button>
    </div>

  </div>  
</template>

<script>
import { completeLesson, getLessonFlashcards } from '../services/LessonService.js'

export default {
  name: 'DailyLessonCard',
  data: function () {
    return {
      token: null,
      empty: false,
    }
  },
  props: {
    todayLesson: Object,
  },
  computed: {
    showFiller: function () {
      return (this.todayLesson == null || this.empty )
    },
  },
  methods: {
    markLessonAsComplete: async function () {
      // Get the lesson flashcards
      await completeLesson(this.token, this.todayLesson.id)
      
      let lessonFlashcards = await getLessonFlashcards(this.token, this.todayLesson.id)
      this.$router.push({ name: 'review', params: { title: 'Lesson Review', reviewCards: lessonFlashcards, done: async function() {
        return
      }}}) 
    },
  },
  created: async function () {
    // # Check if the JWT exists
    let token = localStorage.getItem("token")
    if(token != null) {
      this.token = token    
    } else {
      // Route to login page
      this.$router.push({ path: '/review' })
    }
  } 
}

</script>

<style>

</style>
