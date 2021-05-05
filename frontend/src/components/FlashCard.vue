<template>
  <div class="">
    <div v-if="!flipped" class="shadow-sm p-6 bg-white rounded-md flex justify-center items-center w-full h-full">
      <span class="font-display text-2xl text-secondary text-center">
        What's the benefit of a gated latch over an AND-OR latch?
      </span>
    </div>
    <div v-else class="shadow-sm p-6 bg-primary rounded-md flex justify-center items-center w-full h-full">
      <span class="font-display text-2xl text-white text-center">
        A gated latch only needs one data input line, but an AND-OR latch needs a SET and RESET line
      </span>
    </div>
  </div>
  
</template>

<script>
export default {
  name: 'FlashCard',
  data: function () {
  return {
    flipped: true,
  }
}
}
</script>

<style>

</style>
