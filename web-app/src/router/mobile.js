import Vue from 'vue'
import VueRouter from 'vue-router'
import TasksList from '@/views/mobile/TasksList.vue'
import TaskPage from '@/views/Task.vue'

Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    name: 'dashboard',
    component: TasksList
  },
  {
    path: '/task/:name',
    name: 'task',
    component: TaskPage,
    props: true
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes,
  scrollBehavior (to, from, savedPosition) {
    return { x: 0, y: 0 }
  }
})

export default router
