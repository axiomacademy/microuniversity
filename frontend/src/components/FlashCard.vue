<template>
  <div class="">
    <div v-if="!flipped" class="shadow-sm p-6 bg-white rounded-md flex justify-center items-center w-full h-full">
      <span class="font-display text-2xl text-secondary text-center">
        {{ front }}
      </span>
    </div>
    <div v-else class="shadow-sm p-6 bg-primary rounded-md flex justify-center items-center w-full h-full">
      <span class="font-display text-2xl text-white text-center">
        {{ back }}
      </span>
    </div>
  </div>
  
</template>

<script>
export default {
  name: 'FlashCard',
  props: {
    flipped: Boolean,
    front: String,
    back: String,
  },
}
</script>

<style>

</style>
