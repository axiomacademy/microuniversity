import Vue from 'vue'
import VueRouter from 'vue-router'
import Home from '../views/Home.vue'

Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    name: 'home',
    component: Home
  },
  {
    path: '/login',
    name: 'login',
    component: () => import(/* webpackChunkName: "about" */ '../views/Login.vue'),
  },
  {
    path: '/review',
    name: 'review',
    component: () => import(/* webpackChunkName: "about" */ '../views/Review.vue'),
  },
  {
    path: '/lessons',
    name: 'lessons',
    component: () => import(/* webpackChunkName: "about" */ '../views/Lessons.vue'),
  }
]

const router = new VueRouter({
  routes
})

export default router
