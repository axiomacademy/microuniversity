import Vue from 'vue'
import App from './App.vue'
import router from './router'

// Setting up firebase
// Firebase App (the core Firebase SDK) is always required and must be listed first
import firebase from "firebase/app";

// If you enabled Analytics in your project, add the Firebase SDK for Analytics
import "firebase/analytics";

// Add the Firebase products that you want to use
import "firebase/auth";

// import Review from './Review.vue'

import('@/assets/styles/index.css');

Vue.config.productionTip = false

// Setting up firebase
const firebaseConfig = {
    apiKey: "AIzaSyAvCX9jHDse4kxAVlztKPU0hJ8Bzipl__s",
    authDomain: "axiom-20a6e.firebaseapp.com",
    databaseURL: "https://axiom-20a6e.firebaseio.com",
    projectId: "axiom-20a6e",
    storageBucket: "axiom-20a6e.appspot.com",
    messagingSenderId: "816357746057",
    appId: "1:816357746057:web:9520beb33425f4ef9f2d58",
    measurementId: "G-CNNC5C5FS3"
  };

// Initialize Firebase
firebase.initializeApp(firebaseConfig);
firebase.analytics();

new Vue({
  router,
  render: h => h(App)
}).$mount('#app')
