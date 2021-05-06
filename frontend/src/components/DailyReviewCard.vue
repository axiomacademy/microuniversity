<template>
  <button class="shadow-sm p-6 bg-white rounded-md w-full flex flex-col justify-start items-start" @click="startReview">
    <h5 class="font-display text-md">{{ today }}</h5>
    <h1 class="font-display text-3xl text-secondary font-medium">Daily Review</h1>
    <h6 v-if="this.remainingCards == 0" class="font-body text-sm mt-5">✅ You’ve completed today’s review</h6>
    <h6 v-else class="font-body text-sm mt-5">You have <span class="font-body font-bold text-sm">{{ remainingCards }}</span> cards left to review</h6>
  </button> 
</template>

<script>
export default {
  name: 'DailyReviewCard',
  data: function() {
    return {
      today: new Date().toDateString(),
    }
  },
  props: {
    dailyReviewCards: Array,
  },
  computed: {
    remainingCards: function() {
      return (this.dailyReviewCards == null) ? 0 : this.dailyReviewCards.length
    }, 
  },
  methods: {
    startReview: function() {
      this.$router.push({ name: 'review', params: { title: 'Daily Review', reviewCards: this.dailyReviewCards }}) 
    },
  }
}
</script>

<style>

</style>
