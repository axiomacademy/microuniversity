<template>
  <div class="bg-white shadow-sm px-6 py-4 rounded-md">
    <h2 class="font-body font-bold text-md text-text">{{ challenge.title }}</h2>
    <span class="font-body text-xs text-secondary uppercase">{{ challenge.subject }}</span>
    <div class="flex items-center pt-2">
      <h5 class="font-body text-xs text-test">{{ challenge.description }}</h5>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ChallengeStatus',
  props: {
    challenge: Object,
  },
}
</script>

<style>

</style>
