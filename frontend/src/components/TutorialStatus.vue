<template>
  <div class="bg-white shadow-sm px-6 py-4 rounded-md">
    <div class="flex items-center justify-between items-center">
      <h2 class="font-body font-bold text-md text-text">{{ tutorial.title }}</h2>
      <span class="font-body text-xs text-accent uppercase">{{ tutorial.status }}</span>
    </div>
    <div class="flex items-center pt-2">
      <h5 class="font-body text-xs text-test">{{ tutorial.description }}</h5>
    </div>
    <button class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mt-4 rounded w-full">De-enroll</button>
  </div>
</template>

<script>
export default {
  name: 'TutorialStatus',
  props: {
    tutorial: Object,
  },
}
</script>

<style>

</style>
