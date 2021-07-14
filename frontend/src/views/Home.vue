<template>
  <div id="home" class="wrapper bg-purple-50 min-h-screen flex flex-col pb-32 items-center" v-bind:class="{ 'game': exploreTabOpen }"> 

    <div v-if="!loading" class="lg:w-6/12 w-full">

      <div v-if="openTab == 'Explore'" id="explore-tab">
        EXPLORE
      </div>
      
      <div v-if="openTab == 'Challenges'" id="learn-tab" class="py-10">
        <h1 class="text-3xl font-display font-bold max-w-0 px-10 text-secondary">Mining Activities</h1>
        <div class="px-4 py-4 mt-6">
          <DailyReviewCard :dailyReviewCards="dailyReviewCards" />
        </div>
        <h2 class="text-xl font-body font-bold mt-8 mb-2 px-10 text-text">Challenges</h2>
        <h2 class="text-xl font-body font-bold mt-8 mb-2 px-10 text-text">Tutorials</h2>
      </div>
      
      <div v-if="openTab == 'Learn'" id="learn-tab" class="py-10">
        <h1 class="text-3xl font-display font-bold max-w-0 px-10 text-secondary">Gather Knowledge</h1>
        <h2 class="text-xl font-body font-bold mt-6 px-10 text-text">Recommendations</h2>
        <div class="overflow-x-auto flex flex-nowrap mt-4 pl-4 horizontal">
          <LectureCard v-for="lecture in lectures" :key="lecture.title" :lecture="lecture" :token="token" class="lectureCard" />
        </div>
      </div>
      
      <div v-if="openTab == 'Me'" id="me-tab">
        ME
      </div>
    </div>

    <!-- LOADING INDICATOR -->
    <div v-else class="flex-grow flex flex-col justify-center items-center">
      <MoonLoader class="self-center" color="#7938D8"/>
    </div>

    <!-- Floating tab buttons -->
    <div class="fixed inset-x-0 bottom-0">
      <div class="mb-6 bg-white shadow-lg mx-4 flex lg:w-4/12 h-16 rounded-lg justify-around overflow-hidden">
        <button class="w-4/12 text-2xl focus:outline-none" @click="setActiveTab('Explore')"
          v-bind:class="{ 'bg-purple-100': exploreTabOpen, 'text-secondary': exploreTabOpen }">
          <i class="fas fa-globe-europe"></i> 
        </button>
        <button class="w-4/12 text-2xl focus:outline-none" @click="setActiveTab('Challenges')"
          v-bind:class="{ 'bg-purple-100': challengesTabOpen, 'text-secondary': challengesTabOpen }">
          <i class="fas fa-space-shuttle"></i>
        </button>
        <button class="w-4/12 text-2xl focus:outline-none" @click="setActiveTab('Learn')" 
          v-bind:class="{ 'bg-purple-100': learnTabOpen, 'text-secondary': learnTabOpen }">
          <i class="fas fa-graduation-cap"></i>
        </button>
        <button class="w-4/12 text-2xl focus:outline-none" @click="setActiveTab('Me')" 
          v-bind:class="{ 'bg-purple-100': meTabOpen, 'text-secondary': meTabOpen }">
          <i class="fas fa-user-astronaut"></i>
        </button>
      </div>   
    </div>
  </div>
</template>

<script>
import { MoonLoader } from '@saeris/vue-spinners'

import LectureCard from '../components/LectureCard.vue'
import DailyReviewCard from '../components/DailyReviewCard.vue'

import firebase from "firebase/app";
import "firebase/auth";

export default {
  name: 'App',
  components: {
    MoonLoader,
    LectureCard,
    DailyReviewCard,
  },
  data: function () {
    return {
      loading: true,
      token: "",
      email: "",
      openTab: "Learn",
      unsubAuth: null,
      dailyReviewCards: [],
      lectures: [
        {
          title: "Electronic Computing",
          subject: "Computer Science",
          description: "Learn all about electronic computing and be aware of the 1819 census",
          video_link: "https://www.youtube.com/embed/LN0ucKNX0hc"
        },
        {
          title: "Electronic Computing",
          subject: "Computer Science",
          description: "Learn all about electronic computing and be aware of the 1819 census",
          video_link: "https://www.youtube.com/embed/LN0ucKNX0hc"
        },
        {
          title: "Electronic Computing",
          subject: "Computer Science",
          description: "Learn all about electronic computing and be aware of the 1819 census",
          video_link: "https://www.youtube.com/embed/LN0ucKNX0hc"
        }
      ],
    }
  },
  computed: {
    learnTabOpen: function () {
      return this.openTab == "Learn"
    },
    exploreTabOpen: function () {
      return this.openTab == "Explore"
    },
    meTabOpen: function () {
      return this.openTab == "Me"
    },
    challengesTabOpen: function () {
      return this.openTab == "Challenges"
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
      this.openTab = tab
    },
    logout: async function() {
      this.loading = true
      await firebase.auth().signOut();
    },
  },
}
</script>

<style>

.game {
  background-image: url("../assets/bg.jpeg");
}

.horizontal::-webkit-scrollbar {
  display: none;
}
 
.horizontal {
  -ms-overflow-style: none;  /* IE and Edge */
  scrollbar-width: none;
  scroll-snap-type: x mandatory;
}

.lectureCard {
  scroll-snap-align: center;
  min-width: 80vw;
}

</style>
