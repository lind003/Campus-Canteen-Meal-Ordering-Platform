<script setup>
import { ref, onMounted, watch } from 'vue'
import { useAuthStore } from './stores/auth'
import LoginModal from './components/Login.vue'
import RegisterModal from './components/Register.vue'
import OrderList from './components/OrderList.vue'
import CreateOrderModal from './components/CreateOrder.vue'
import ComplaintModal from './components/Complaint.vue'
import { useRouter } from 'vue-router'


const auth = useAuthStore()
const showLogin = ref(false)
const showRegister = ref(false)
const showCreateOrder = ref(false)
const showComplaint = ref(false)
const selectedOrderId = ref(null)

const router = useRouter()


const handleClick=()=>{
  console.log('🔥 按钮被点击了')
  console.log('当前路由对象:', router.currentRoute.value)
  console.log('准备跳转 /user')
  router.push('/user').then(() => {
    console.log('✅ 路由跳转成功')
  }).catch(err => {
    console.error('❌ 路由跳转失败:', err)
  })
}

const logout = () => {
  auth.logout()
  // 刷新页面确保状态清除
  window.location.reload()
}

const handleLoginSuccess = () => {
  showLogin.value = false
  console.log('登录成功回调 - 当前状态:', {
    isAuthenticated: auth.isAuthenticated,
    user: auth.user,
    token: auth.token
  })
  // 不需要刷新页面，因为 store 状态已经更新
}

// 修复注册成功处理
const handleRegisterSuccess = (value) => {
  if (!value) { // 当模态框关闭时
    showRegister.value = false
    console.log('注册成功 - 当前状态:', {
      isAuthenticated: auth.isAuthenticated,
      user: auth.user
    })
  }
}

const handleOrderCreated = () => {
  showCreateOrder.value = false
  console.log('订单创建成功')
}

const handleComplaint = (orderId) => {
  selectedOrderId.value = orderId
  showComplaint.value = true
}
// 在组件挂载时检查认证状态
onMounted(() => {
  console.log('App 加载 - 初始化认证状态')
  auth.checkAuth()
})
</script>

<template>
  <div id="app">
    <nav class="navbar navbar-expand-lg navbar-dark bg-primary">
      <div class="container">
        <span class="navbar-brand">校园食堂代买餐平台</span>
        <div class="navbar-nav ms-auto">
          <template v-if="auth.isAuthenticated && auth.user">
            <!-- 用户信息下拉菜单 -->
            <div class="nav-item dropdown">
              <a class="nav-link dropdown-toggle" href="#" role="button" data-bs-toggle="dropdown">
                欢迎, {{ auth.user.name }}
              </a>
              <ul class="dropdown-menu">
                <li><button class="dropdown-item" @click="handleClick">个人信息</button></li>
                <li><hr class="dropdown-divider"></li>
                <li><a class="dropdown-item" href="#" @click="logout">退出登录</a></li>
              </ul>
            </div>
            
            <button class="btn btn-outline-light btn-sm me-2" @click="showCreateOrder = true">
              发布需求
            </button>
          </template>
          <template v-else>
            <button class="btn btn-outline-light btn-sm me-2" @click="showLogin = true">登录</button>
            <button class="btn btn-light btn-sm" @click="showRegister = true">注册</button>
          </template>
        </div>
      </div>
    </nav>

    <main class="container mt-4">
      <!-- 主页内容 -->
      <router-view 
        :filter="'all'"
        @create-order="showCreateOrder = true"
        @complain="handleComplaint"
      />
    </main>

    <!-- 登录模态框 -->
    <LoginModal 
      v-model="showLogin" 
      @login-success="handleLoginSuccess" 
    />
    
    <!-- 注册模态框 -->
    <RegisterModal 
      v-model="showRegister" 
      @update:modelValue="handleRegisterSuccess"
    />
    
    <!-- 创建订单模态框 -->
    <CreateOrderModal 
      v-model="showCreateOrder" 
      @order-created="handleOrderCreated" 
    />
    
    <!-- 投诉模态框 -->
    <ComplaintModal 
      v-model="showComplaint" 
      :order-id="selectedOrderId" 
    />
   
  </div>
</template>