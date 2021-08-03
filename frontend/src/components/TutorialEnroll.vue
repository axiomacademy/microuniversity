<template>
  <div class="bg-white shadow-sm px-6 py-4 rounded-md">
    <h2 class="font-body font-bold text-md text-text">{{ tutorial.title }}</h2>
    <div class="flex items-center pt-2">
      <h5 class="font-body text-xs text-test">{{ tutorial.description }}</h5>
    </div>
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
