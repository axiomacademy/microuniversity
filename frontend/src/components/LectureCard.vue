<template>
  <div class="shadow-sm bg-white rounded-md flex flex-col py-6 mx-4">
    <h3 class="font-body text-lg font-bold text-text px-6">{{ lecture.title }}</h3>
    <span class="font-body text-sm text-gray-400 px-6 pb-4">{{ lecture.subject }}</span>
    <!-- Image preview -->
    <iframe class="w-full h-52" allowFullScreen="allowFullScreen" frameBorder="0"
      :src="lecture.video_link">
    </iframe>
    <span class="font-body text-sm font-light text-text px-6 pt-4">
      {{ lecture.description }}
    </span>
    
    <button @click="markLectureAsComplete" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mx-6 mt-4 rounded">Complete</button>
  </div>
</template>

<script>
import { completeLecture, getLectureFlashcards } from '../services/LectureService.js'

export default {
  name: 'LectureCard',
  props: {
    lecture: Object,
    token: String,
  },
  methods: {
    markLectureAsComplete: async function () {
      // Get the lecture flashcards
      await completeLecture(this.token, this.lecture.id)
      
      let lectureFlashcards = await getLectureFlashcards(this.token, this.lecture.id)
      this.$router.push({ name: 'review', params: { title: 'Lecture Review', reviewCards: lectureFlashcards, done: async function() {
        return
      }}}) 
    },
  },
}

</script>

<style>

</style>
