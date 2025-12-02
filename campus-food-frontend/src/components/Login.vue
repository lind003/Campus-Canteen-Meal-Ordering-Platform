<template>
  <div v-if="modelValue" class="modal-overlay" @click="close">
    <div class="modal-content" @click.stop>
      <h2>登录</h2>
      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label>手机号:</label>
          <input v-model="form.phone" type="text" placeholder="请输入手机号" required>
        </div>
        <div class="form-group">
          <label>密码:</label>
          <input v-model="form.password" type="password" placeholder="请输入密码" required>
        </div>
        <div class="form-actions">
          <button type="submit" :disabled="loading">
            {{ loading ? '登录中...' : '登录' }}
          </button>
          <button type="button" @click="close">取消</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { authAPI } from '../utils/api'
import { useAuthStore } from '../stores/auth'

const emit = defineEmits(['update:modelValue', 'loginSuccess'])
const props = defineProps(['modelValue'])
const auth = useAuthStore()

const form = ref({
  phone: '',
  password: ''
})
const loading = ref(false)

const close = () => {
  emit('update:modelValue', false)
}

const handleLogin = async () => {
  loading.value = true
  try {
    const response = await authAPI.login(form.value)
    console.log('登录响应:', response.data)
    
    if (response.data.success) {
      const responseData = response.data.data
      const userData = responseData.user
      const token = responseData.token
      
      console.log('用户数据:', userData)
      console.log('Token:', token)
      
      if (userData && token) {
        // 处理可能的 NULL 值
        const processedUserData = {
          ...userData,
          grade: userData.grade || '', // 如果 grade 为 null，设置为空字符串
          college: userData.college || '', // 同样处理其他可能为 null 的字段
          name: userData.name || ''
        }
        
        auth.login(processedUserData, token)
        emit('loginSuccess')
        close()
        form.value = { phone: '', password: '' }
        alert('登录成功！')
      } else {
        alert('登录响应数据不完整')
      }
    } else {
      alert('登录失败: ' + response.data.message)
    }
  } catch (error) {
    console.error('登录错误:', error)
    if (error.response?.data?.message) {
      alert('登录失败: ' + error.response.data.message)
    } else {
      alert('登录失败，请检查网络连接')
    }
  } finally {
    loading.value = false
  }
}

watch(() => props.modelValue, (newVal) => {
  if (!newVal) {
    form.value = { phone: '', password: '' }
  }
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  min-width: 400px;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
}

.form-group input {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1.5rem;
}

.form-actions button {
  flex: 1;
  padding: 0.75rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.form-actions button[type="submit"] {
  background: #007bff;
  color: white;
}

.form-actions button[type="submit"]:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.form-actions button[type="button"] {
  background: #6c757d;
  color: white;
}
</style>