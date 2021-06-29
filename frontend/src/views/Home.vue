<template>
  <div id="home" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col pb-32 items-center">
    <!-- Nav bar -->
    <nav class="relative p-4 border-b border-purple-200 lg:w-6/12 w-full">
      <ul class="flex flex-row items-center">
        <img src="../assets/logo-transparent-dark.png" class="w-14 h-14"/>
        <div class="text-secondary bg-purple-100 px-4 h-12 flex justify-center items-center rounded-full ml-auto mr-4">
          🔥 {{ streak }}
        </div>
        <button @click="showMenu = !showMenu" class="w-12 h-12 rounded-lg text-secondary bg-purple-100"><i class="fas fa-bars"></i></button>
      </ul>

      <!-- Hamburger menu -->
    <div v-if="showMenu" class="absolute bottom-0 right-0 -mb-32 mr-4 bg-purple-100 shadow-sm flex w-8/12 md:w-4/12 rounded-md overflow-hidden">
      <ul class="flex flex-col w-full">
        <li class="text-sm px-4 py-3 text-text">Signed in as <span class="font-medium">{{ email }}</span></li>
        <div class="bg-purple-200 w-full" style="height: 1px;"></div>
        <button @click="$router.push({ name: 'profile' })" class="text-sm text-text py-2 text-left px-4 hover:bg-purple-200 focus:bg-purple-200">
          Profile
        </button>
        <div class="bg-purple-200 w-full" style="height: 1px;"></div>
        <button @click="logout" class="text-sm text-text py-2 text-left px-4 hover:bg-purple-200 focus:bg-purple-200">
          Sign Out
        </button>
      </ul>
    </div>
    </nav>
      
    <div v-if="!loading" class="lg:w-6/12 w-full">
      <!-- Explore tab with modules, enrolled cohorts etc. -->
      <div v-if="openTab == 'Explore'" id="explore-tab">
        <h1 class="font-display text-2xl text-secondary pl-4 mt-6">Explore Modules 🧠</h1> 
      </div>

      <!-- Learning tab with daily review, lessons and tutorials -->
      <div v-if="openTab == 'Learn'" id="learn-tab">
        <DailyReviewCard :dailyReviewCards="this.dailyReviewCards" class="mt-4" />
        <!-- Daily lecture section -->
        <h1 class="font-display text-2xl text-secondary pl-4 mt-8">Today's Lecture</h1> 
        <DailyLectureCard class="mt-4" :todayLecture="this.todayLecture" :token="this.token" /> 
        <button class="bg-purple-200 w-full font-display font-light text-secondary py-2 px-6 rounded flex mt-4">View previous lectures...</button>
        <h1 class="font-display text-2xl text-secondary pl-4 mt-8 font-normal mb-2">Upcoming Tutorials</h1>

        <!-- Generating Tutorial List -->
        <div v-if="upcomingTutorials.length == 0" class="shadow-sm bg-white rounded-md flex flex-col justify-center items-center py-6 mt-4">
          <img src="../assets/relax.png" class="w-24" />
          <h3 class="font-display text-lg text-text font-medium pt-4">No upcoming tutorials</h3>
          <h6 class="font-display text-md text-text">Looks like you can relax for a while!</h6>
        </div>
        <TutorialListElement v-else v-for="tutorial in upcomingTutorials" :key="tutorial.id" :title="tutorial.title" :datetime="tutorial.scheduled_time" class="mt-2" />
      </div>
    </div>

    <!-- LOADING INDICATOR -->
    <div v-else class="flex-grow flex flex-col justify-center items-center">
      <MoonLoader class="self-center" color="#7938D8"/>
    </div>

    <!-- Floating tab buttons -->
    <div class="fixed inset-x-0 bottom-0 mb-8 bg-white shadow-lg mx-auto flex w-9/12 lg:w-4/12 h-16 rounded-full justify-around overflow-hidden">
      <button class="w-6/12 font-display focus:outline-none" @click="setActiveTab('Explore')"
        v-bind:class="{ 'bg-purple-100': exploreTabOpen, 'text-secondary': exploreTabOpen, 'text-text': exploreTabOpen}">
        🌎 <span class="pl-1" v-bind:class="{ 'font-medium': exploreTabOpen }">Explore</span>
      </button>
      <div class="h-full bg-purple-200" style="width: 1px;"></div>
      <button class="w-6/12 font-display focus:outline-none" @click="setActiveTab('Learn')" 
        v-bind:class="{ 'bg-purple-100': learnTabOpen, 'text-secondary': learnTabOpen, 'text-text': learnTabOpen }">
        📖 <span class="pl-1" v-bind:class="{ 'font-medium': learnTabOpen }">Learn</span>
      </button>
    </div>
  </div>
</template>

<script>
import DailyReviewCard from '../components/DailyReviewCard.vue'
import DailyLectureCard from '../components/DailyLectureCard.vue'
import TutorialListElement from '../components/TutorialListElement.vue'
import { MoonLoader } from '@saeris/vue-spinners'


// Services

import firebase from "firebase/app";
import "firebase/auth";

export default {
  name: 'App',
  components: {
    DailyReviewCard,
    DailyLectureCard,
    TutorialListElement,
    MoonLoader,
  },
  data: function () {
    return {
      loading: true,
      token: "",
      email: "",
      streak: 0,
      todayLecture: null,
      upcomingTutorials: [],
      dailyReviewCards: [],
      openTab: "Learn",
      showMenu: false,
      existingCohort: {},
      unsubAuth: null,
    }
  },
  computed: {
    learnTabOpen: function () {
      return this.openTab == "Learn"
    },
    exploreTabOpen: function () {
      return this.openTab == "Explore"
    },
  },
  created: async function () {
    this.loading = true

    // Based on observer
    this.unsubAuth = firebase.auth().onAuthStateChanged(async (user) => {
      console.log("Home")
      this.loading = true
      if (user) {
        this.token = await user.getIdToken(true)
        this.email = user.email

        console.log(this.token)

        // set the localstorage
        localStorage.setItem("FB_TOKEN", this.token)
        localStorage.setItem("EMAIL", user.email)

        this.loading = false
      } else {
        this.$router.push({ name: 'login' })
        this.loading = false
      }
    })
  },
  beforeDestroy() {
    this.unsubAuth()
  },
  methods: {
    setActiveTab: async function(tab) {
      if(tab == "Learn") {
        this.openTab = "Learn"
        this.loading = true
        // Functions go here
        this.loading = false
      } else if (tab == "Explore") {
        this.openTab = "Explore"
        this.loading = true
        // Functions go here
        this.loading = false
      } else {
        return 
      }
    },
    logout: async function() {
      this.loading = true
      await firebase.auth().signOut();
    },
  },
}
</script>

<style>

</style>
