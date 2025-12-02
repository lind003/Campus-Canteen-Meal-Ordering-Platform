<template>
  <div v-if="modelValue" class="modal-overlay" @click="close">
    <div class="modal-content" @click.stop>
      <h2>注册</h2>
      <form @submit.prevent="handleRegister">
        <div class="form-group">
          <label>姓名:</label>
          <input v-model="form.name" type="text" placeholder="请输入姓名" required>
        </div>
        <div class="form-group">
          <label>手机号:</label>
          <input v-model="form.phone" type="text" placeholder="请输入手机号" required>
        </div>
        <div class="form-group">
          <label>学院:</label>
          <input v-model="form.college" type="text" placeholder="请输入学院" required>
        </div>
        <div class="form-group">
          <label>年级:</label>
          <input v-model="form.grade" type="text" placeholder="请输入年级" required>
        </div>
        <div class="form-group">
          <label>密码:</label>
          <input v-model="form.password" type="password" placeholder="请输入密码" required>
        </div>
        <div class="form-group">
          <label>确认密码:</label>
          <input v-model="form.confirmPassword" type="password" placeholder="请再次输入密码" required>
        </div>
        <div class="form-group">
          <label>角色:</label>
          <select v-model="form.role">
            <option value="student">学生</option>
            <option value="teacher">教师</option>
          </select>
        </div>
        <div class="form-actions">
          <button type="submit" :disabled="loading">
            {{ loading ? '注册中...' : '注册' }}
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

const emit = defineEmits(['update:modelValue'])
const props = defineProps(['modelValue'])
const auth = useAuthStore()

const form = ref({
  name: '',
  phone: '',
  college: '',
  grade: '',
  password: '',
  confirmPassword: '',
  role: 'student'
})
const loading = ref(false)

const close = () => {
  emit('update:modelValue', false)
}

const handleRegister = async () => {
  if (form.value.password !== form.value.confirmPassword) {
    alert('两次输入的密码不一致')
    return
  }

  loading.value = true
  try {
    const response = await authAPI.register(form.value)
    console.log('注册响应:', response.data)
    
    if (response.data.success) {
      const responseData = response.data.data
      const userData = responseData.user
      const token = responseData.token
      
      console.log('注册用户数据:', userData)
      console.log('注册Token:', token)
      
      if (userData && token) {
        // 处理可能的 NULL 值
        const processedUserData = {
          ...userData,
          grade: userData.grade || '',
          college: userData.college || '',
          name: userData.name || ''
        }
        
        auth.login(processedUserData, token)
        alert('注册成功！已自动登录')
        close()
        form.value = {
          name: '',
          phone: '',
          college: '',
          grade: '',
          password: '',
          confirmPassword: '',
          role: 'student'
        }
      } else {
        alert('注册响应数据不完整')
      }
    } else {
      alert('注册失败: ' + response.data.message)
    }
  } catch (error) {
    console.error('注册错误:', error)
    if (error.response?.data?.message) {
      alert('注册失败: ' + error.response.data.message)
    } else {
      alert('注册失败，请检查网络连接')
    }
  } finally {
    loading.value = false
  }
}

watch(() => props.modelValue, (newVal) => {
  if (!newVal) {
    form.value = {
      name: '',
      phone: '',
      college: '',
      grade: '',
      password: '',
      confirmPassword: '',
      role: 'student'
    }
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
  max-height: 80vh;
  overflow-y: auto;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
}

.form-group input,
.form-group select {
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
  background: #28a745;
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