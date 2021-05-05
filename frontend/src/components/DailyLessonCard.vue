<template>
  <div>
    <div v-if="empty" class="shadow-sm bg-white rounded-md flex flex-col justify-center items-center py-6">
      <img src="../assets/empty.png" class="w-16 h-16" />
      <h3 class="font-display text-lg text-text font-medium pt-4">No new sessions</h3>
      <h6 class="font-display text-md text-text">You're done for the day!</h6>
    </div>
    <div v-else class="shadow-sm bg-white rounded-md flex flex-col py-6">
      <h3 class="font-display text-lg font-medium text-text px-6">Computational History</h3>
      <span class="font-body text-sm text-gray-400 px-6 pb-4">How Electrons Think </span>
      <!-- Image preview -->
      <div class="w-full h-52 bg-gray-200"></div>
      <span class="font-body text-sm font-light text-text px-6 pt-4">
        Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor
      </span>
      
      <button class="bg-primary hover:bg-secondary tracking-widest font-body text-xs text-medium text-white uppercase p-2 mx-6 mt-4 rounded">Complete</button>
    </div>

  </div>  
</template>

<script>
export default {
  name: 'DailyLessonCard',
  data: function () {
    return {
      empty: false,
    }
  }
}
</script>

<style>

</style>
