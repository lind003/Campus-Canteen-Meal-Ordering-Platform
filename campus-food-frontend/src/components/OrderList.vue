<script setup>
import { ref, onMounted, watch, computed } from 'vue'  
import { useAuthStore } from '../stores/auth'
import { orderAPI } from '../utils/api'

const props = defineProps(['filter'])
const emit = defineEmits(['create-order', 'complain'])
const auth = useAuthStore()


const orders = ref([])
const loading = ref(false)
const shownDetail = ref(null)          // 当前展开的订单ID
const toggleDetail = (id) => {
  shownDetail.value = shownDetail.value === id ? null : id
}
const handleCreateOrder = () => {
  console.log('创建订单按钮被点击')
  console.log('当前认证状态:', auth.isAuthenticated)
  console.log('当前用户:', auth.user)
  emit('create-order')
}

const counterPart = (order) => {
  if (auth.user?.role === 'runner') {
    // 跑腿员 看下单人
    return {
      name: order.demander_name,
      phone: order.demander_phone,
       card: order.demander_card,
      addr: order.delivery_address,
      tip:  order.tip
    }
  } else {
    // 下单人看跑腿员
    return {
      name: order.runner_name,
      card: order.runner_card,
      phone: order.runner_phone,
      addr: order.delivery_address,   
      tip:  order.tip
    }
  }
}

// 状态映射
const statusMap = {
  'pending': { text: '待接单', type: 'warning' },
  'accepted': { text: '已接单', type: 'primary' },
  'processing': { text: '处理中', type: 'info' },
  'completed': { text: '已完成', type: 'success' },
  'cancelled': { text: '已取消', type: 'danger' }
}

// 监听认证状态变化
watch(() => auth.isAuthenticated, (newVal) => {
  console.log('OrderList - 认证状态变化:', newVal)
  if (newVal) {
    loadOrders()
  } else {
    orders.value = []
  }
})



const loadOrders = async () => {
  console.log('开始加载订单，认证状态:', auth.isAuthenticated)
  
  if (!auth.isAuthenticated) {
    console.log('用户未登录，无法加载订单')
    return
  }
  
  loading.value = true
  try {
    console.log('调用订单API，filter:', props.filter)
    const response = await orderAPI.list(props.filter)
    console.log('订单加载成功:', response.data)
    orders.value = response.data.data
  } catch (error) {
    console.error('加载订单失败:', error)
    if (error.response?.status === 401) {
      console.log('Token 无效，执行登出')
      auth.logout()
    }
  } finally {
    loading.value = false
  }
}

// 监听 filter 变化
watch(() => props.filter, (newFilter) => {
  if (auth.isAuthenticated) {
    loadOrders()
  }
})

// 获取状态文本
const getStatusText = (status) => {
  return statusMap[status]?.text || status
}

// 获取状态类型（用于颜色）
const getStatusType = (status) => {
  return statusMap[status]?.type || 'info'
}

// 判断用户是否能接单
const canTakeOrder = (order) => {
  return auth.user?.role === 'runner' && order.status === 'pending'
}

// 判断用户是否能完成订单
const canCompleteOrder = (order) => {
  return auth.user?.role === 'runner' && order.status === 'accepted'
}

// 判断用户是否能取消订单
const canCancelOrder = (order) => {
  const userRole = auth.user?.role
  if (!userRole) return false
  
  if (userRole === 'demander') {
    return order.status === 'pending' // 需求者只能取消待接单的订单
  }
  
  if (userRole === 'admin') {
    return order.status === 'pending' || order.status === 'accepted' // 管理员可以取消进行中的订单
  }
  
  if (userRole === 'runner') {
    return false // 跑腿员不能取消订单
  }
  
  return false
}

// 接单
const takeOrder = async (orderId) => {
  if (!confirm('确认接单吗？')) return
  
  try {
    const response = await orderAPI.take(orderId)
    console.log('接单成功:', response)
    alert('接单成功')
    loadOrders() // 重新加载订单列表
  } catch (error) {
    console.error('接单失败:', error)
    alert(error.response?.data?.message || '接单失败')
  }
}

// 完成订单
const completeOrder = async (orderId) => {
  if (!confirm('确认完成订单吗？')) return
  
  try {
    const response = await orderAPI.complete(orderId)
    console.log('完成订单成功:', response)
    alert('订单完成成功')
    loadOrders() // 重新加载订单列表
  } catch (error) {
    console.error('完成订单失败:', error)
    alert(error.response?.data?.message || '完成订单失败')
  }
}

// 取消订单
const cancelOrder = async (orderId) => {
  const reason = prompt('请输入取消原因:', '用户取消')
  if (reason === null) return // 用户取消了
  
  if (!confirm('确认取消订单吗？')) return
  
  try {
    const response = await orderAPI.cancel(orderId,  reason)
    console.log('取消订单成功:', response)
    alert('订单取消成功')
    loadOrders() // 重新加载订单列表
  } catch (error) {
    console.error('取消订单失败:', error)
    alert(error.response?.data?.message || '取消订单失败')
  }
}
onMounted(() => {
  console.log('OrderList 加载 - 认证状态:', auth.isAuthenticated)
  console.log('当前 user:', auth.user)
  
  if (auth.isAuthenticated) {
    loadOrders()
  } else {
    console.log('用户未登录，无法加载订单')
  }
})
</script>

<template>
  <div class="order-list">
    <div class="d-flex justify-content-between align-items-center mb-3">
      <h4>订单列表</h4>
      <button 
        v-if="auth.isAuthenticated"
        class="btn btn-primary"
         @click="handleCreateOrder"
      >
        创建订单
      </button>
    </div>

    <div v-if="!auth.isAuthenticated" class="alert alert-warning">
      请先登录查看订单
    </div>

    <div v-else-if="loading" class="text-center">
      <div class="spinner-border" role="status">
        <span class="visually-hidden">加载中...</span>
      </div>
    </div>

    <div v-else-if="orders.length === 0" class="alert alert-info">
      暂无订单
    </div>

    <div v-else>
      <div v-for="order in orders" :key="order.order_id" class="card mb-3">
        <div class="card-body">
          <h5 class="card-title">订单 #{{ order.order_id }}</h5>
          <p class="card-text">
            <strong>状态:</strong> 
            <span class="badge" :class="`bg-${getStatusType(order.status)}`">
              {{ getStatusText(order.status) }}
            </span>
          </p>
          <p class="card-text"><strong>金额:</strong> ¥{{ order.total_amount }}</p>
          <p class="card-text"><strong>配送地址:</strong> {{ order.delivery_address }}</p>
          <p class="card-text"><strong>创建时间:</strong> {{ new Date(order.order_time).toLocaleString() }}</p>
          
          <!-- 操作按钮 -->
          <div class="d-flex gap-2">
            <!-- 详情按钮 -->
          <button class="btn btn-outline-secondary btn-sm"
            @click="toggleDetail(order.order_id)">
          详细信息
          </button>

          <!-- 展开区域 -->
          <div v-if="shownDetail === order.order_id"
            class="w-100 mt-2 small text-muted">
            <div><strong>姓名：</strong>{{ counterPart(order).name }}</div>
            <div><strong>校园卡号：</strong>{{ counterPart(order).card || '暂无' }}</div>
            <div><strong>联系方式：</strong>{{ counterPart(order).phone }}</div>
            <div><strong>送达地址：</strong>{{ counterPart(order).addr }}</div>
            <div><strong>小费：</strong>¥{{ counterPart(order).tip.toFixed(2) }}</div>
          </div>
            <!-- 接单按钮 -->
            <button 
              v-if="canTakeOrder(order)"
              class="btn btn-success btn-sm"
              @click="takeOrder(order.order_id)"
            >
              接单
            </button>
            
            <!-- 完成订单按钮 -->
            <button 
              v-if="canCompleteOrder(order)"
              class="btn btn-primary btn-sm"
              @click="completeOrder(order.order_id)"
            >
              完成
            </button>
            
            <!-- 取消订单按钮 -->
            <button 
              v-if="canCancelOrder(order)"
              class="btn btn-danger btn-sm"
              @click="cancelOrder(order.order_id)"
            >
              取消
            </button>
            
            <!-- 投诉按钮 -->
            <button 
              class="btn btn-warning btn-sm"
              @click="emit('complain', order.order_id)"
            >
              投诉
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.order-list {
  padding: 20px;
}

.card {
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  border: none;
}

.card-title {
  color: #333;
  margin-bottom: 15px;
}

.card-text {
  margin-bottom: 8px;
  color: #666;
}

.badge {
  font-size: 0.875em;
  padding: 0.35em 0.65em;
}

.gap-2 {
  gap: 8px;
}

.btn-sm {
  padding: 0.25rem 0.5rem;
  font-size: 0.875rem;
}
</style>