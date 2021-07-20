<template>
  <div id="home" class="wrapper bg-purple-50 min-h-screen flex flex-col pb-24"> 

    <div v-if="!loading" class="lg:w-6/12 w-full flex-grow flex flex-col">

      <div v-if="openTab == 'Explore'" id="explore-tab" class="pt-6 flex-grow flex flex-col items-center px-6">
        <img src="../assets/planet.svg" class="w-8/12"/>
        <h1 class="font-display text-3xl mt-4 font-bold">{{ currentPlanet.name }}</h1>
        <h2 class="font-display text-lg text-secondary">{{ currentPlanet.starSystem.name }}</h2>
        
        <div class="mt-2 flex">
          <Chip class="w-20 mt-1 mr-2">1000 🪙</Chip>
          <Chip class="w-20 mt-1">100 ⚡</Chip>
        </div>

        <span class="font-body text-lg text-text mt-6 self-start font-semibold">Mining Status</span>
        <div class="h-6 mt-2 relative w-full rounded-full overflow-hidden">
          <div class="w-full h-full bg-purple-100 absolute"></div>
          <div class="h-full bg-primary absolute" style="width:10%"></div>
        </div>
        <div class="shadow-sm px-6 py-3 bg-white rounded-md flex items-center w-full mt-8">
          <img src="../assets/planets.svg" class="w-12">
          <div class="ml-6">
            <h1 class="font-normal text-text text-lg">Visit another planet</h1>
            <Chip class="w-20 mt-1">10 ⚡</Chip>
          </div>
        </div>
        <div class="shadow-sm px-6 py-3 bg-white rounded-md flex items-center w-full mt-3">
          <img src="../assets/galaxy.svg" class="w-12">
          <div class="ml-6">
            <h1 class="font-normal text-text text-lg">Visit nearby starsystem</h1>
            <Chip class="w-20 mt-1">30 ⚡</Chip>
          </div>
        </div>
        <div class="shadow-sm px-6 py-3 bg-white rounded-md flex items-center w-full mt-3">
          <img src="../assets/rocket.svg" class="w-12">
          <div class="ml-6">
            <h1 class="font-normal text-text text-lg">Recharge rocket</h1>
            <Chip class="w-20 mt-1">100 🪙</Chip>
          </div>
        </div>
      </div>
      
      <div v-if="openTab == 'Challenges'" id="learn-tab" class="pt-10 flex-grow">
        <h1 class="text-3xl font-display font-bold max-w-0 px-10 text-secondary">Mining Activities</h1>
        <div class="px-4 pt-4">
          <DailyReviewCard :dailyReviewCards="dailyReviewCards" />
        </div>

        <h2 class="text-2xl font-semibold mt-6 mb-3 px-10 text-text">Challenges</h2>
        <ChallengeStatus v-if="activeChallenge != null" :challenge="activeChallenge" class="mx-4" />
        <div v-else  class="overflow-x-auto flex flex-nowrap my-auto pl-4 h-full horizontal">
          <ChallengeAccept v-for="challenge in challenges" :key="challenge.title" :challenge="challenge" class="min-w-50 mr-4" />
        </div>


        <h2 class="text-2xl font-semibold mt-6 mb-3 px-10 text-text">Tutorials</h2>
        <TutorialCohortStatus v-if="enrolledCohort != null" :tutorial="enrolledCohort" class="mx-4" />
        <div v-else class="overflow-x-auto flex flex-nowrap my-auto pl-4 h-full horizontal">
          <TutorialEnroll v-for="tutorial in tutorials" :key="tutorial.title" :tutorial="tutorial" class="min-w-50 mr-4" />
        </div>
      </div>
      
      <div v-if="openTab == 'Learn'" id="learn-tab" class="pt-10 flex-grow flex flex-col justify-center">
        <h1 class="text-3xl font-display font-bold max-w-0 px-10 text-secondary">Gather Knowledge</h1>
        <input v-model="search" type="text" placeholder="Search for any knowledge..." class="bg-purple-100 p-2 rounded font-display border border-transparent focus:outline-none focus:ring-2 focus:ring-purple-600 focus:border-transparent mt-4 placeholder-purple-300 mx-6" :disabled="loading">
        <div class="overflow-x-auto flex flex-nowrap my-auto pl-4 h-full horizontal">
          <LectureCard v-for="lecture in lectures" :key="lecture.id" :lecture="lecture" :token="token" class="lectureCard" />
        </div>
      </div>
      
      <div v-if="openTab == 'Me'" id="me-tab" class="pt-10 flex-grow flex flex-col px-8">
        <img src="../assets/spaceship.svg" class="w-6/12"/>
        <h1 class="font-display text-xl mt-4 font-bold">Sudharshan</h1>
        <h2 class="font-display text-xl text-text">Sundaramahalingam</h2>
        <span class="font-display text-sm text-accent mt-1">Probably somewhere trying to draw a rock</span>
        <p class="font-body text-md text-text mt-6">I’m an aspiring polymath, on a mission to bring optimism and curiosity back to the world. I want to help people by spreading positivity and creating technologies that solve real problems.</p>
        <div class="flex flex-wrap w-full mt-4">
          <Chip v-for="topic in masteredTopics" :key="topic" class="mt-2 mr-2">{{ topic }}</Chip>
          <Chip class="mt-2 mr-2 font-bold">+ 30</Chip>
        </div>
        
        <button class="bg-purple-100 w-full font-display font-light text-secondary py-2 px-6 rounded flex mt-6">View History</button>
        <button class="bg-purple-100 w-full font-display font-light text-secondary py-2 px-6 rounded flex mt-2">Edit Profile</button>
        <button class="bg-purple-100 w-full font-display font-light text-secondary py-2 px-6 rounded flex mt-2" @click="logout">Logout</button>
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

import Chip from '../components/Chip.vue'
import LectureCard from '../components/LectureCard.vue'
import DailyReviewCard from '../components/DailyReviewCard.vue'
import TutorialEnroll from '../components/TutorialEnroll.vue'
import TutorialCohortStatus from '../components/TutorialCohortStatus.vue'
import ChallengeAccept from '../components/ChallengeAccept.vue'
import ChallengeStatus from '../components/ChallengeStatus.vue'

import firebase from "firebase/app";
import "firebase/auth";

export default {
  name: 'App',
  components: {
    Chip,
    MoonLoader,
    LectureCard,
    DailyReviewCard,
    TutorialEnroll,
    TutorialCohortStatus,
    ChallengeAccept,
    ChallengeStatus,
  },
  data: function () {
    return {
      loading: true,
      token: "",
      email: "",
      openTab: "Learn",
      unsubAuth: null,
      search: "",
      dailyReviewCards: [],
      masteredTopics: [
        "Introduction to Python",
        "Adobe Illustration",
        "Drawing 101",
        "Data Analytics",
        "Classifiers and KNNs",
      ],
      currentPlanet: {
        name: "Venus-2",
        minedKnowledge: "200",
        totalKnowledge: "1000",
        starSystem: {
          name: "The Chitauri System"
        }
      },
      activeChallenge: null,
      challenges: [
        {
          title: "Design a simple logical ciruit",
          status: "UNLOCKED",
          subject: "Computer Science",
          description: "Design a circuit to evaluate any simple logical statement",
        },
        {
          title: "Build a half adder",
          status: "UNLOCKED",
          subject: "Computer Science",
          description: "Using a digital design platform, build a half adder."
        },
        {
          title: "Build a full adder",
          status: "UNLOCKED",
          subject: "Computer Science",
          description: "Using a digital design platform, build a full adder to demonstrate bitwise adding",
        },
      ],
      enrolledCohort: null /*{
          title: "Designing a 8-bit CPU",
          topic: "Computer architecture",
          status: "ENROLLED",
          description: "Run through the process of designing a basic 8-bit CPU",
      }*/,
      tutorials: [
        {
          title: "Designing a 8-bit CPU",
          topic: "Computer architecture",
          description: "Run through the process of designing a basic 8-bit CPU",
        },
        {
          title: "Python Fractal Generator",
          topic: "Introduction to Python",
          description: "Create a Mendelbrot Set using Python to visualise fractals",
        },
      ],
      lectures: [
        {
          id:1,
          title: "Electronic Computing",
          subject: "Computer Science",
          description: "Learn all about electronic computing and be aware of the 1819 census",
          video_link: "https://www.youtube.com/embed/LN0ucKNX0hc"
        },
        {
          id:2,
          title: "Electronic Computing",
          subject: "Computer Science",
          description: "Learn all about electronic computing and be aware of the 1819 census",
          video_link: "https://www.youtube.com/embed/LN0ucKNX0hc"
        },
        {
          id:3,
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

.min-w-50 {
  scroll-snap-align: center;
  min-width: 50vw;
}

</style>
