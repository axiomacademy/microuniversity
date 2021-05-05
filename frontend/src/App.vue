<template>
  <div class="wrapper bg-purple-50 min-h-screen p-4">
    <nav class="p-4 border-b border-gray-200">
      <ul class="flex flex-row items-center">
        <img src="./assets/logo-transparent-dark.png" class="w-14 h-14"/>
        <div class="text-secondary bg-purple-100 px-4 h-12 flex justify-center items-center rounded-full ml-auto mr-4">
          🔥 30
        </div>
        <button class="w-12 h-12 rounded-lg text-secondary bg-purple-100"><i class="fas fa-sign-out-alt"></i></button>
      </ul>
    </nav>
    
    <DailyReviewCard class="mt-4" />
    <!-- Daily lesson section -->
    <h1 class="font-display text-2xl text-secondary pl-4 mt-8 font-normal">Today's Session</h1> 
    <DailyLessonCard class="mt-4" />
    <button class="bg-purple-200 w-full font-display font-light text-secondary py-2 px-6 rounded flex mt-4">View previous lessons...</button>
    <h1 class="font-display text-2xl text-secondary pl-4 mt-8 font-normal">Upcoming Tutorials</h1>
    <TutorialListElement class="mt-4" />
    <TutorialListElement class="mt-2" />
    <TutorialListElement class="mt-2" />
  </div>
</template>

<script>
import DailyReviewCard from './components/DailyReviewCard.vue'
import DailyLessonCard from './components/DailyLessonCard.vue'
import TutorialListElement from './components/TutorialListElement.vue'

export default {
  name: 'App',
  components: {
    DailyReviewCard,
    DailyLessonCard,
    TutorialListElement,
  }
}
</script>

<style>

</style>
