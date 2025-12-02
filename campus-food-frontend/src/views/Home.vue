<!-- views/Home.vue -->
<template>
  <div>
    <div class="header-actions">
      <h4>校园代买平台</h4>
      <button 
        v-if="auth.isAuthenticated"
        class="btn btn-primary"
        @click="showCreateOrder = true"
      >
        发布代买需求
      </button>
    </div>

    <!-- 订单列表 -->
    <OrderList 
      :filter="currentFilter"
      @create-order="showCreateOrder = true"
      @complain="handleComplaint"
    />
    
    <!-- 创建订单模态框 -->
    <CreateOrderModal 
      v-model="showCreateOrder" 
      @order-created="handleOrderCreated"
    />
    <ComplaintModal v-model="showComplaint" :order-id="selectedOrderId" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import OrderList from '../components/OrderList.vue'
import CreateOrderModal from '../components/CreateOrder.vue'
import ComplaintModal from '../components/Complaint.vue'

const auth = useAuthStore()
const showCreateOrder = ref(false)
const showComplaint = ref(false)
const selectedOrderId = ref(null)

const handleComplaint = (orderId) => {
  selectedOrderId.value = orderId
  showComplaint.value = true
}

const handleOrderCreated = () => {
  showCreateOrder.value = false
}
</script>

<style scoped>
.header-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding: 1rem;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.header-actions h4 {
  margin: 0;
  color: #333;
}
</style>