import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import UserCenter from '@/views/UserCenter.vue'


const routes = [
  { path: '/', component: Home },
  { path: '/user', component: UserCenter, meta: { requiresAuth: true } },
  //{ path: '/admin', component: Admin, meta: { requiresAuth: true, requiresAdmin: true } }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  console.log(' 路由守卫检查:', to.path, '需要认证:', to.meta.requiresAuth)

  const token = localStorage.getItem('token') 
  if (to.meta.requiresAuth && !token) {
    console.warn(' 没有 token，跳回 /')
    next('/')
  } else {
    console.log(' 有 token，通过守卫')
    next()
  }
})
router.beforeEach((to, from, next) => {
  console.log(' 路由跳转:', from.path, '→', to.path)
  next()
})
export default router