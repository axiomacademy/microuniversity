<template>
  <div class="bg-white shadow-sm px-6 py-3 rounded-md">
    <div class="flex items-center justify-between">
      <h2 class="font-display text-md text-text">{{ title }}</h2>
      <!-- span class="font-display text-xs text-secondary uppercase">30 mins</span-->
    </div>
    <div class="flex items-center pt-1">
      <h5 class="font-body text-xs text-primary uppercase">{{ formattedTime }}</h5>
      <span class="px-1 text-xs text-gray-400 font-body">|</span>
      <h5 class="font-body text-xs text-gray-400">{{ formattedDate }}</h5>
    </div>
  </div>
</template>

<script>
const optionsTime = {
   hour12 : true,
   hour:  "numeric",
   minute: "numeric",seconds:"numeric"
}

const optionsDate = {
     year: "numeric",
     month:"short",
     day:"2-digit"
}

export default {
  name: 'TutorialListElement',
  props: {
    title: String,
    datetime: String,
  },
  computed: {
    formattedTime: function() {
      return new Date(this.datetime).toLocaleTimeString("en-US", optionsTime) 
    },
    formattedDate: function() {
      return new Date(this.datetime).toLocaleDateString("en-US", optionsDate) 
    },
  }
}
</script>

<style>

</style>
