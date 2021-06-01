<template>
  <div class="flex flex-row pt-4 items-start">
    <img class="w-16 h-16 rounded-lg shadow-sm" :src="image" />
    <div class="flex flex-col pl-4">
      <h1 class="font-display text-md font-medium text-text">{{ title }}</h1>
      <h2 class="font-regular text-sm text-purple-400">{{ id }}</h2>
      <h2 class="font-regular text-xs text-text mt-1">{{ description }}</h2>
      <div class="flex flex-row mt-2 items-start">
        <div class="text-secondary bg-purple-100 px-4 h-8 flex justify-center items-center rounded-full text-xs mr-auto">
          {{ readableDuration }}
        </div>
      </div>
    </div>

  </div>
</template>

<script>
export default {
  name: 'ModuleListElement',
  props: {
    title: String,
    id: String,
    image: String,
    description: String,
    duration: Number
  },
  computed: {
    readableDuration () {
      if(this.duration < 7) {
        return this.duration + " days"
      } else if (this.duration < 30) {
        return this.duration + " days ≈ " + (this.duration/7).toFixed(2) + " weeks"
      } else if (this.duration < 365) {
        return this.duration + " days ≈ " + (this.duration/30).toFixed(2) + " months"
      } else {
        return this.duration + " days"
      }
    },
  }
}
</script>

<style>

</style>
