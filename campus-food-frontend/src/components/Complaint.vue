<template>
  <div v-if="modelValue" class="modal-overlay" @click="close">
    <div class="modal-content" @click.stop>
      <h2>提交投诉</h2>
      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label>订单ID:</label>
          
          <input :value="orderId" type="number" disabled>
        </div>
        
        <div class="form-group">
          <label>投诉类型:</label>
          <select v-model="form.type" required>
            <option value="">请选择投诉类型</option>
            <option value="delivery">配送问题</option>
            <option value="quality">质量问题</option>
            <option value="service">服务态度</option>
            <option value="other">其他问题</option>
          </select>
        </div>

        <div class="form-group">
          <label>投诉原因:</label>
          <textarea v-model="form.reason" rows="4" placeholder="请详细描述投诉原因..." required></textarea>
        </div>

        <div class="form-actions">
          <button type="submit" :disabled="loading">
            {{ loading ? '提交中...' : '提交投诉' }}
          </button>
          <button type="button" @click="close">取消</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { complaintAPI } from '../utils/api'

const emit = defineEmits(['update:modelValue'])
const props = defineProps({
  modelValue: Boolean,
  orderId: Number
})

const form = ref({
  order_id: props.orderId, // 在 JavaScript 中设置 order_id
  type: '',
  reason: ''
})
const loading = ref(false)

const close = () => {
  emit('update:modelValue', false)
}

const handleSubmit = async () => {
  if (!form.value.type || !form.value.reason) {
    alert('请填写完整的投诉信息')
    return
  }

  // 确保 order_id 正确设置
  const submitData = {
    order_id: props.orderId,
    type: form.value.type,
    reason: form.value.reason
  }

  console.log('提交投诉数据:', submitData)

  loading.value = true
  try {
    const response = await complaintAPI.create(submitData)
    console.log('投诉提交成功:', response.data)
    alert('投诉提交成功！')
    close()
    resetForm()
  } catch (error) {
    console.error('投诉提交失败:', error)
    if (error.response?.data?.message) {
      alert('投诉提交失败: ' + error.response.data.message)
    } else {
      alert('投诉提交失败，请检查网络连接')
    }
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.value = {
    type: '',
    reason: ''
  }
}

// 监听 orderId 变化
watch(() => props.orderId, (newVal) => {
  console.log('订单ID变化:', newVal)
})

// 监听模态框显示/隐藏
watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    console.log('投诉模态框打开，订单ID:', props.orderId)
    resetForm()
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
  min-width: 500px;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: bold;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.form-group input:disabled {
  background-color: #f8f9fa;
  color: #6c757d;
}

.form-group textarea {
  resize: vertical;
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