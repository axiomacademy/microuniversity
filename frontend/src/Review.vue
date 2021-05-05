<template>
  <div class="wrapper bg-purple-50 h-screen p-4 flex flex-col justify-start">
    <nav class="p-4 border-b border-gray-200">
      <ul class="flex flex-row items-center justify-between">
        <h1 class="font-display text-3xl text-secondary font-medium">Session Review</h1>
        <button class="w-12 h-12 rounded-lg text-secondary bg-purple-100"><i class="fas fa-sign-out-alt"></i></button>
      </ul>
    </nav>
    
    <h1 class="font-display text-2xl text-secondary mt-8 font-medium">1 of 20</h1> 
    <FlashCard class="w-full h-2/4 self-center mt-4"/>
    <div class="button-group flex flex-row justify-around pt-5">
      <button class="w-24 h-24 bg-white shadow-sm rounded-lg"><i class="fas fa-times text-3xl text-red-500"></i></button>
      <button class="w-24 h-24 bg-white shadow-sm rounded-lg"><i class="fas fa-check text-3xl text-green-500"></i></button>
    </div>
  </div> 
</template>

<script>
import FlashCard from './components/FlashCard.vue'

export default {
  name: 'App',
  components: {
    FlashCard,
  }
}
</script>

<style>

</style>
