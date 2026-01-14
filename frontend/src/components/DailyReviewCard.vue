<template>
  <button class="shadow-sm p-6 bg-white rounded-md w-full flex flex-col justify-start items-start" @click="startReview" :disabled="dailyReviewCards.length == 0">
    <h5 class="font-body text-md">{{ today }}</h5>
    <h1 class="font-body font-semibold text-3xl text-secondary font-medium">Daily Review</h1>
    <h6 v-if="this.remainingCards == 0" class="font-body text-sm mt-5">✅ You’ve completed today’s review</h6>
    <h6 v-else class="font-body text-sm mt-5">You have <span class="font-body font-bold text-sm">{{ remainingCards }}</span> cards left to review</h6>
  </button> 
</template>

<script>
import { completeReview } from '../services/ReviewService'

const options = {
  weekday: "long",
  month:"short",
  day:"2-digit"
}

export default {
  name: 'DailyReviewCard',
  data: function() {
    return {
      today: new Date().toLocaleDateString("en-US",options),
    }
  },
  props: {
    dailyReviewCards: Array,
  },
  computed: {
    remainingCards: function() {
      return this.dailyReviewCards.length
    }, 
  },
  methods: {
    startReview: function() {
      this.$router.push({ name: 'review', params: { title: 'Daily Review', reviewCards: this.dailyReviewCards, done: async function(token) {
        await completeReview(token) 
      }}}) 
    },
  }
}
</script>

<style>

</style>
