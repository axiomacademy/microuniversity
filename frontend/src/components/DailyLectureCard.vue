<template>
  <div>
    <div v-if="this.showFiller" class="shadow-sm bg-white rounded-md flex flex-col justify-center items-center py-6">
      <img src="../assets/done.png" class="w-16 h-16" />
      <h3 class="font-display text-lg text-text font-medium pt-4">No daily lecture</h3>
      <h6 class="font-display text-md text-text">You're done for the day!</h6>
    </div>
    <div v-else class="shadow-sm bg-white rounded-md flex flex-col py-6">
      <h3 class="font-display text-lg font-medium text-text px-6">{{ todayLecture.title }}</h3>
      <span class="font-body text-sm text-gray-400 px-6 pb-4">{{ todayLecture.module }}</span>
      <!-- Image preview -->
      <iframe class="w-full h-52" allowFullScreen="allowFullScreen" frameBorder="0"
        :src="todayLecture.video_link">
      </iframe>
      <span class="font-body text-sm font-light text-text px-6 pt-4">
        {{ todayLecture.description }}
      </span>
      
      <button @click="markLectureAsComplete" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mx-6 mt-4 rounded">Complete</button>
    </div>

  </div>  
</template>

<script>
import { completeLecture, getLectureFlashcards } from '../services/LectureService.js'

import firebase from "firebase/app";
import "firebase/auth";

export default {
  name: 'DailyLectureCard',
  data: function () {
    return {
      token: null,
    }
  },
  props: {
    todayLecture: Object,
  },
  computed: {
    showFiller: function () {
      return (this.todayLecture == null )
    },
  },
  methods: {
    markLectureAsComplete: async function () {
      // Get the lecture flashcards
      await completeLecture(this.token, this.todayLecture.id)
      
      let lectureFlashcards = await getLectureFlashcards(this.token, this.todayLecture.id)
      this.$router.push({ name: 'review', params: { title: 'Lecture Review', reviewCards: lectureFlashcards, done: async function() {
        return
      }}}) 
    },
  },
  created: async function () {
    // Check and retrieve firebase credentials
    firebase.auth().onAuthStateChanged(async (user) => {
      if (user) {
        console.log("Logged in")
        this.token = await user.getIdToken(true) 
      } else {
        console.log("Not logged in")
        this.$router.push({ name: 'login' })
      }
    })
  } 
}

</script>

<style>

</style>
