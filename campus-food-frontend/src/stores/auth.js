import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  // 从 localStorage 初始化状态
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))
  const token = ref(localStorage.getItem('token') || null)
  
  const isAuthenticated = computed(() => {
    const hasToken = !!token.value
    console.log('认证状态计算 - token:', token.value, 'hasToken:', hasToken)
    return hasToken
  })

  // 添加 isLoggedIn 作为别名，兼容现有代码
  const isLoggedIn = computed(() => !!token.value)
  
  function login(userData, userToken) {
    console.log('执行登录 - 输入数据:', { userData, userToken })
    
    // 处理可能的 NULL 值
    const processedUserData = {
      id: userData.user_id || userData.id,
      name: userData.name || '',
      phone: userData.phone || '',
      college: userData.college || '',
      grade: userData.grade || '',
      role: userData.role || 'student'
    }
    
    user.value = processedUserData
    token.value = userToken
    
    // 保存到 localStorage
    localStorage.setItem('token', userToken)
    localStorage.setItem('user', JSON.stringify(processedUserData))
    
    console.log('登录后状态:', { 
      user: user.value, 
      token: token.value,
      isAuthenticated: isAuthenticated.value,
      storedToken: localStorage.getItem('token')
    })
  }
  
  function logout() {
    console.log('执行登出')
    user.value = null
    token.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }
  
  // 检查并同步认证状态
  function checkAuth() {
    const storedToken = localStorage.getItem('token')
    const storedUser = localStorage.getItem('user')
    
    console.log('检查认证状态:', { storedToken, storedUser })
    
    if (storedToken && storedUser) {
      try {
        user.value = JSON.parse(storedUser)
        token.value = storedToken
        console.log('认证状态恢复成功:', {
          user: user.value,
          token: token.value,
          isAuthenticated: isAuthenticated.value
        })
      } catch (error) {
        console.error('恢复认证状态失败:', error)
        logout()
      }
    } else {
      console.log('没有找到存储的认证信息')
    }
  }
  
  return {
    user,
    token,
    isAuthenticated,
    isLoggedIn, // 兼容性
    login,
    logout,
    checkAuth
  }
})