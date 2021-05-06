<template>
  <div class="bg-white shadow-sm px-6 py-2 rounded-md">
    <h2 class="font-display text-md text-text">{{ title }}</h2>
    <h5 class="font-body text-sm text-gray-400">1st May, 1400 - 1700</h5>
  </div>
</template>

<script>
export default {
  name: 'TutorialListElement',
  props: {
    title: String,
    datetime: String,
  },
}
</script>

<style>

</style>
