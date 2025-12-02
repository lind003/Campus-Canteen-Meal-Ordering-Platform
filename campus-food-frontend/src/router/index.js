import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'


const routes = [
  { path: '/', component: Home },
  { path: '/user', component: UserCenter, meta: { requiresAuth: true } },
  { path: '/admin', component: Admin, meta: { requiresAuth: true, requiresAdmin: true } }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const auth = JSON.parse(localStorage.getItem('user') || '{}')
  
  if (to.meta.requiresAuth && !auth.token) {
    next('/')
  } else if (to.meta.requiresAdmin && auth.role !== 'admin') {
    next('/')
  } else {
    next()
  }
})

export default router