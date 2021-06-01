<template>
  <div id="lectures" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col justify-start">
    <nav class="p-4 border-b border-purple-200 mb-3">
      <ul class="flex flex-row items-center justify-between">
        <h1 class="font-display text-3xl text-secondary font-medium">Lecture History</h1>
        <button class="w-12 h-12 rounded-lg text-secondary bg-purple-100" @click="goBack"><i class="fas fa-arrow-left"></i></button>
      </ul>
    </nav>
    
    <div v-if="!loading">
      <div v-if="lectures.length == 0" class="shadow-sm bg-white rounded-md flex flex-col justify-center items-center p-6 mt-4">
        <img src="../assets/empty.png" class="w-24" />
        <h3 class="font-display text-lg text-text font-medium pt-4">No past lectures</h3>
        <h6 class="font-display text-sm text-text text-center mt-2">As you finish lectures, they'll appear here in case you want to take a look at them again :)</h6>
      </div>
      <div v-else v-for="lecture in lectures" :key="lecture.id" class="shadow-sm bg-white rounded-md flex flex-col py-6 my-3">
        <div class="flex flex-row justify-start items-center px-6">
          <h3 class="font-display text-lg font-medium text-text">{{ lecture.title }}</h3>
          <span v-if="!lecture.completed" class="font-body text-xs text-white ml-auto px-2 py-1 bg-red-400 uppercase rounded tracking-wide">overdue</span>
        </div>
        <span class="font-body text-sm text-gray-400 px-6 pb-4">{{ getFormattedDate(lecture.scheduled_date) }}</span>
        <!-- Image preview -->
        <iframe class="w-full h-52" allowFullScreen="allowFullScreen" frameBorder="0"
          :src="lecture.video_link">
        </iframe>
        <span class="font-body text-sm font-light text-text px-6 pt-4">
          {{ lecture.description }}
        </span>
        
        <button v-if="!lecture.completed" @click="completeAndReviewCards(lecture.id)" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mx-6 mt-4 rounded">Complete</button>
        <button v-else @click="reviewCards(lecture.id)" class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mx-6 mt-4 rounded">Review</button>
      </div>
    </div>
    <div v-else class="flex-grow flex flex-col justify-center items-center">
      <MoonLoader class="self-center" color="#7938D8"/>
    </div>
  </div> 
</template>

<script>
import { MoonLoader } from '@saeris/vue-spinners'
import { getLecturesPast, getLectureFlashcards, completeLecture } from '../services/LectureService.js'

import firebase from "firebase/app";
import "firebase/auth";

const options = {
     year: "numeric",
     month:"short",
     day:"2-digit"
}

export default {
  name: 'Lectures',
  components: {
    MoonLoader,
  },
  data: function () {
    return {
      loading: true,
      token: null,
      lectures: Array,
    }
  },
  created: async function () {
    this.loading = true
    // # Check and retrieve firebase credentials
    firebase.auth().onAuthStateChanged(async (user) => {
      this.loading = true
      if (user) {
        this.token = await user.getIdToken(true)

        // Get the past lectures
        this.lectures = await getLecturesPast(this.token)

        if(this.lectures.length != 0) {
          this.lectures.sort((a,b) => {
            let d1 = new Date(a.scheduled_date)
            let d2 = new Date(b.scheduled_date)

            return d2 - d1
          })
          
        this.loading = false
        }
      } else {
        this.$router.push({ name: 'login' })
        this.loading = false
      }
    })
  },
  methods: {
    goBack: function() {
      this.$router.go(-1)
    },
    getFormattedDate: function(dateString) {
      return new Date(dateString).toLocaleDateString("en-US", options)
    },
    reviewCards: async function(id) {
      let lectureFlashcards = await getLectureFlashcards(this.token, id)
      this.$router.push({ name: 'review', params: { title: 'Lecture Review', reviewCards: lectureFlashcards, done: async function() {
        return
      }}}) 
    },
    completeAndReviewCards: async function(id) {
      // Get the lecture flashcards
      await completeLecture(this.token, id)

      let lectureFlashcards = await getLectureFlashcards(this.token, id)
      this.$router.push({ name: 'review', params: { title: 'Lecture Review', reviewCards: lectureFlashcards, done: async function() {
        return
      }}}) 
    },
  }
}
</script>

<style>

</style>
