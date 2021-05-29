<template>
  <div id="review" class="wrapper bg-purple-50 h-screen p-4 flex flex-col justify-start">
    <nav class="p-4 border-b border-purple-200">
      <ul class="flex flex-row items-center justify-between">
        <h1 class="font-display text-3xl text-secondary font-medium">{{ $route.params.title }}</h1>
        <button class="w-12 h-12 rounded-lg text-secondary bg-purple-100" @click="goBack"><i class="fas fa-arrow-left"></i></button>
      </ul>
    </nav>
    
    <h1 class="font-display text-2xl text-secondary mt-8 font-medium">{{ index + 1 }} of {{ totalCards }}</h1> 
    <FlashCard class="w-full h-2/4 self-center mt-4" :flipped="flippedCard"
      :front="topText"
      :back="bottomText"
      @click.native="flipCard"
      />
    <div class="button-group flex flex-row justify-around pt-5">
      <button class="w-24 h-24 bg-white shadow-sm rounded-lg" @click="failCard"><i class="fas fa-times text-3xl text-red-500"></i></button>
      <button class="w-24 h-24 bg-white shadow-sm rounded-lg" @click="passCard"><i class="fas fa-check text-3xl text-green-500"></i></button>
    </div>
  </div> 
</template>

<script>
import FlashCard from '../components/FlashCard.vue'

import firebase from "firebase/app";
import "firebase/auth";

// services
import { passFlashcard, failFlashcard } from '../services/ReviewService.js'

export default {
  name: 'App',
  components: {
    FlashCard,
  },
  data: function() {
    return {
      token: null,
      index: 0,
      flippedCard: false,
      cards: Array,
    }
  },
  computed: {
    totalCards: function() {
      return this.cards.length
    },
    topText: function() {
      return this.cards[this.index].top_side
    },
    bottomText: function() {
      return this.cards[this.index].bottom_side
    },
  },
  created: async function() {
    // # Check and retrieve firebase credentials
    let user = await firebase.auth().currentUser;
 
    if(user == null) { 
      this.$router.push({ name: 'login' })
    }

    this.token = await user.getIdToken(true)
    // Copy in the param
    this.cards = this.$route.params.reviewCards
  },
  methods: {
    goBack: function() {
      this.$router.go(-1)
    },
    passCard: async function() {
      await passFlashcard(this.token, this.cards[this.index].id) 

      if(this.index < (this.totalCards - 1)) {
        this.flippedCard = false
        this.index += 1
      } else {
        // Done so call the done() callback and exit
        await this.$route.params.done(this.token)
        this.$router.go(-1) 
      }
    },
    failCard: async function() {
      await failFlashcard(this.token, this.cards[this.index].id) 
      
      if(this.index < (this.totalCards - 1)) {
        this.flippedCard = false
        this.index += 1
      } else {
        // Done so call the done() callback and exit
        await this.$route.params.done(this.token)
        this.$router.go(-1) 
      }
    },
    flipCard: function() {
      this.flippedCard = !(this.flippedCard)
    },
  }
}
</script>

<style>

</style>
