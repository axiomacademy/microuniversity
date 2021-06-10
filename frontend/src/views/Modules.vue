<template> 
  <div id="modules" class="wrapper bg-purple-50 min-h-screen p-4 flex flex-col pb-32">
    <nav class="p-4 border-b border-purple-200 mb-3">
      <ul class="flex flex-row items-center justify-between">
        <h1 class="font-display text-3xl text-secondary font-medium">My Modules</h1>
        <button class="w-12 h-12 rounded-lg text-secondary bg-purple-100" @click="goBack"><i class="fas fa-arrow-left"></i></button>
      </ul>
    </nav>
    
    <div v-if="!loading">
      <h1 class="font-display text-2xl text-secondary pl-4 mt-4">Active Module 📅</h1>
      <ModuleListElement v-for="module in getCurrentModules" @click.native="moduleOpen(module)" :key="module.id" :title="module.title" :id="module.id" :description="module.description" :image="module.image" :duration="module.duration"/>
      <div v-if="getCurrentModules.length == 0" class="bg-purple-200 font-display font-light text-sm text-secondary py-2 px-6 rounded flex mt-4 mx-4">No ongoing modules. Find something new to start learning today!</div>
  
      <div class="w-full self-center bg-purple-200 my-8" style="height: 1px;"></div>

      <h1 class="font-display text-2xl text-secondary pl-4">Completed Modules ✅</h1>
      <ModuleListElement v-for="module in getCompletedModules" @click.native="moduleOpen(module)" :key="module.id" :title="module.title" :id="module.id" :description="module.description" :image="module.image" :duration="module.duration"/>
      <div v-if="getCompletedModules.length == 0" class="bg-purple-200 font-display font-light text-sm text-secondary py-2 px-6 rounded flex mt-4 mx-4">No completed modules!</div>
    </div>

    <div v-else class="flex-grow flex flex-col justify-center items-center">
      <MoonLoader class="self-center" color="#7938D8"/>
    </div>
  </div>
</template>

<script>
import { MoonLoader } from '@saeris/vue-spinners'
import ModuleListElement from '../components/ModuleListElement.vue'

import { getSelfCohorts } from '../services/CohortService'

import firebase from "firebase/app";
import "firebase/auth";

export default {
  name: "Modules",
  components: {
    ModuleListElement,
    MoonLoader
  },
  data: function() {
    return {
      loading: true,
      token: "",
      modules: []
    }
  },
  computed: {
    getCurrentModules() {
      return this.modules.filter(module => module.status < 3)
    },
    getCompletedModules() {
      return this.modules.filter(module => module.status == 3)
    },
  },
  created() {
    this.loading = true

    firebase.auth().onAuthStateChanged(async (user) => {
      this.loading = true
      if (user) {
        this.token = await user.getIdToken(true) 
        this.modules = await getSelfCohorts(this.token)

        if(this.modules == null) {
          this.modules = []
        }
        
        this.loading = false
      } else {
        this.$router.push({ name: 'login' })
        this.loading = false
      }
    })
  },
  methods: {
    goBack: function() {
      this.$router.go(-1)
    },
    moduleOpen: function (module) {
      this.$router.push({name: 'module', params: { module: module }})
    },
  }
}
</script>

<style>

</style>
