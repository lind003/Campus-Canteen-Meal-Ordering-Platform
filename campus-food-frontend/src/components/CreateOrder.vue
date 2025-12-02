<template>
  <div v-if="modelValue" class="modal-overlay" @click="close">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h2>发布代买需求</h2>
        <button type="button" class="btn-close" @click="close">×</button>
      </div>

      

      <!-- 平台说明 -->
      <div class="alert alert-info mb-3">
        <h6 class="alert-heading"> 校园代买平台</h6>
        <p class="mb-0">
          <strong>代买流程：</strong><br>
          1. 您发布代买需求 → 2. 跑腿员接单 → 3. 跑腿员购买并配送 → 4. 您确认收货付款
        </p>
      </div>

      <form @submit.prevent="handleSubmit">
        <!-- 食堂选择 -->
        <div class="form-group">
          <label>选择食堂 <span class="text-danger">*</span>:</label>
          <select v-model="form.canteen_id" @change="loadMerchants" required>
            <option value="">请选择食堂</option>
            <option v-for="canteen in canteens" :key="canteen.canteen_id" :value="canteen.canteen_id">
              {{ canteen.canteen_name }}
            </option>
          </select>
        </div>

        <!-- 商家选择 -->
        <div class="form-group">
          <label>选择商家 <span class="text-danger">*</span>:</label>
          <select v-model="form.merchant_id" @change="loadDishes" :disabled="!form.canteen_id" required>
            <option value="">请选择商家</option>
            <option v-for="merchant in merchants" :key="merchant.merchant_id" :value="merchant.merchant_id">
              {{ merchant.merchant_name }} - {{ merchant.window_no || '未知窗口' }}
            </option>
          </select>
        </div>

         <!-- 显示商家详情 -->
        <div v-if="currentMerchant && currentMerchant.description" class="merchant-info">
          <h6>
            <i class="bi bi-shop"></i>
            餐厅信息
          </h6>
          <div class="merchant-details">
            <p class="merchant-description">
              <i class="bi bi-chat-left-text"></i>
              {{ currentMerchant.description || '暂无商家描述' }}
            </p>
            <p v-if="currentMerchant.business_hours">
              <i class="bi bi-clock"></i>
              营业时间: {{ currentMerchant.business_hours || '未设置' }}
            </p>
            <p>
              <i class="bi bi-telephone"></i>
              联系电话: {{ currentMerchant.phone || '未设置' }}
            </p>
          </div>
        </div>

        <!-- 菜品选择区域 -->
        <div class="form-group">
          <label>选择菜品 <span class="text-danger">*</span>:</label>
          
          <!-- 加载状态 -->
          <div v-if="loadingDishes" class="loading-dishes">
            <div class="spinner-border text-primary" role="status"></div>
            <span>加载菜品中...</span>
          </div>

          <!-- 无菜品 -->
          <div v-else-if="dishes.length === 0 && form.merchant_id" class="no-dishes">
            <i class="bi bi-emoji-frown"></i>
            <p>该商家暂无菜品</p>
          </div>

          <!-- 菜品列表 -->
          <div v-else class="dishes-container">
            <!-- 简单的菜品列表 -->
            <div v-for="dish in dishes" :key="dish.dish_id" class="dish-item">
              <div class="dish-info">
                <div class="dish-header">
                  <h6>{{ dish.dish_name }}</h6>
                  <span class="dish-price">¥{{ dish.price }}</span>
                </div>
                
                <!-- 显示菜品描述 -->
                <p v-if="dish.description" class="dish-description">
                  {{ dish.description }}
                </p>
                
                <!-- 显示辣度 -->
                <div v-if="dish.spicy_level > 0" class="spicy-level">
                  辣度: 
                  <span v-for="i in 3" :key="i" class="spicy-dot" :class="{ 'hot': i <= dish.spicy_level }"></span>
                </div>
                
                <!-- 显示类别 -->
                <div class="dish-category">
                  <span class="category-tag">{{ getCategoryLabel(dish.category) }}</span>
                  <span v-if="!dish.is_available" class="sold-out">售罄</span>
                </div>
              </div>
              
              <!-- 数量选择 -->
              <div class="dish-controls">
                <button 
                  type="button" 
                  @click="decreaseQuantity(dish.dish_id)"
                  class="qty-btn"
                  :disabled="!isDishSelected(dish.dish_id) || getDishQuantity(dish.dish_id) <= 1"
                >-</button>
                
                <span v-if="isDishSelected(dish.dish_id)" class="qty-number">
                  {{ getDishQuantity(dish.dish_id) }}
                </span>
                <span v-else class="qty-placeholder">0</span>
                
                <button 
                  type="button" 
                  @click="increaseQuantity(dish.dish_id)"
                  class="qty-btn"
                  :disabled="!dish.is_available"
                >+</button>
                
                <button 
                  type="button" 
                  @click="toggleDish(dish)"
                  class="select-btn"
                  :class="{ 'selected': isDishSelected(dish.dish_id), 'disabled': !dish.is_available }"
                  :disabled="!dish.is_available"
                >
                  {{ isDishSelected(dish.dish_id) ? '已选' : (dish.is_available ? '选择' : '售罄') }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 订单详情 -->
        <div v-if="selectedDishesList.length > 0" class="order-summary">
          <h5>已选菜品:</h5>
          <div class="selected-dishes">
            <div v-for="item in selectedDishesList" :key="item.dish_id" class="selected-item">
              <span>{{ getDishName(item.dish_id) }}</span>
              <span class="qty-badge">×{{ item.qty }}</span>
              <span class="price">¥{{ (getDishPrice(item.dish_id) * item.qty).toFixed(2) }}</span>
            </div>
          </div>
          <div class="summary-total">
            <div>菜品总计: <strong>¥{{ totalAmount.toFixed(2) }}</strong></div>
            <div v-if="form.tip > 0">小费: <strong>¥{{ form.tip.toFixed(2) }}</strong></div>
            <div class="grand-total">
              订单总额: <strong>¥{{ (totalAmount + form.tip).toFixed(2) }}</strong>
            </div>
          </div>
        </div>

        <!-- 期望送达时间 -->
        <div class="form-group">
          <label>期望送达时间 <span class="text-danger">*</span>:</label>
          <input 
            v-model="form.delivery_time" 
            type="datetime-local" 
            :min="minDeliveryTime"
            required
          >
          <small class="form-text">请至少选择{{ minMinutes }}分钟后的时间</small>
        </div>

        <!-- 配送地址 -->
        <div class="form-group">
          <label>配送地址 <span class="text-danger">*</span>:</label>
          <input 
            v-model="form.delivery_address" 
            type="text" 
            placeholder="例如：计算机学院教学楼A栋101室"
            required
          >
        </div>

        <!-- 联系电话 -->
        <div class="form-group">
          <label>联系电话 <span class="text-danger">*</span>:</label>
          <input 
            v-model="form.contact_phone" 
            type="tel" 
            placeholder="请填写能联系到您的手机号"
            required
          >
        </div>

        <!-- 小费 -->
        <div class="form-group">
          <label>小费（可选）:</label>
          <div class="tip-options">
            <button 
              type="button" 
              v-for="tip in tipOptions" 
              :key="tip"
              :class="{ 'active': form.tip === tip }"
              @click="form.tip = tip"
              class="tip-btn"
            >
              ¥{{ tip }}
            </button>
            <div class="custom-tip">
              <input 
                type="number" 
                v-model.number="form.custom_tip" 
                placeholder="自定义"
                min="0"
                step="0.01"
                @input="form.tip = form.custom_tip || 0"
              >
            </div>
          </div>
          <small class="form-text">适当的小费可以更快吸引跑腿员接单</small>
        </div>

        <!-- 备注 -->
        <div class="form-group">
          <label>备注要求:</label>
          <textarea 
            v-model="form.notes" 
            placeholder="例如：多要辣椒、不要葱、送到后门、打包要求等"
            rows="3"
          ></textarea>
        </div>

        <!-- 紧急程度 -->
        <div class="form-group">
          <label>紧急程度:</label>
          <div class="urgency-buttons">
            <button 
              type="button" 
              v-for="option in urgencyOptions" 
              :key="option.value"
              :class="['urgency-btn', `urgency-${option.value}`, { 'active': form.urgency === option.value }]"
              @click="form.urgency = option.value"
            >
              {{ option.label }}
            </button>
          </div>
        </div>

        <div class="form-actions">
          <button type="submit" :disabled="loading || !canSubmit" class="submit-btn">
            {{ loading ? '发布中...' : '发布代买需求' }}
          </button>
          <button type="button" @click="close" class="cancel-btn">取消</button>
        </div>
      </form>
    </div>
  </div>

</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { dataAPI, orderAPI } from '../utils/api'

const props = defineProps(['modelValue'])
const emit = defineEmits(['update:modelValue', 'orderCreated'])

console.log('CreateOrder 组件加载，props:', props.modelValue)

// 表单数据
const form = ref({
  canteen_id: '',
  merchant_id: '',
  delivery_time: '',
  delivery_address: '',
  contact_phone: '',
  tip: 0,
  custom_tip: '',
  notes: '',
  urgency: 'normal'
})

// 数据列表
const canteens = ref([])
const merchants = ref([])
const dishes = ref([])
const loading = ref(false)
const loadingDishes = ref(false)
const currentMerchant = ref(null)

// 选择的菜品 [{dish_id, qty}]
const selectedDishes = ref({}) // 使用对象存储，key: dish_id, value: qty

// 小费选项
const tipOptions = [2, 3, 5, 8, 10]

// 类别映射
const categoryMap = {
  'breakfast': '早餐',
  'lunch': '午餐', 
  'dinner': '晚餐',
  'snack': '小吃',
  'beverage': '饮品',
  'other': '其他'
}

// 紧急程度选项
const urgencyOptions = [
  { value: 'normal', label: '普通', color: 'info' },
  { value: 'urgent', label: '紧急', color: 'warning' },
  { value: 'very_urgent', label: '非常紧急', color: 'danger' }
]

// 计算属性
const minMinutes = 30
const minDeliveryTime = computed(() => {
  const now = new Date()
  now.setMinutes(now.getMinutes() + minMinutes)
  return now.toISOString().slice(0, 16)
})

const selectedDishesList = computed(() => {
  return Object.entries(selectedDishes.value).map(([dish_id, qty]) => ({
    dish_id: parseInt(dish_id),
    qty
  }))
})

const totalAmount = computed(() => {
  return selectedDishesList.value.reduce((sum, item) => {
    const dish = dishes.value.find(d => d.dish_id === item.dish_id)
    return sum + (dish ? dish.price * item.qty : 0)
  }, 0)
})

const canSubmit = computed(() => {
  return form.value.canteen_id && 
         form.value.merchant_id && 
         selectedDishesList.value.length > 0 &&
         form.value.delivery_time &&
         form.value.delivery_address &&
         form.value.contact_phone &&
         !loading.value
})

// 方法
const close = () => {
  console.log('关闭模态框')
  emit('update:modelValue', false)
}

const isDishSelected = (dishId) => {
  return !!selectedDishes.value[dishId]
}

const getDishQuantity = (dishId) => {
  return selectedDishes.value[dishId] || 0
}

const toggleDish = (dish) => {
  if (!dish.is_available) return
  const dishId = dish.dish_id
  if (selectedDishes.value[dishId]) {
    delete selectedDishes.value[dishId]
  } else {
    selectedDishes.value[dishId] = 1
  }
}

const increaseQuantity = (dishId) => {
  const dish = dishes.value.find(d => d.dish_id === dishId)
  if (!dish || !dish.is_available) return
  
  if (selectedDishes.value[dishId]) {
    selectedDishes.value[dishId]++
  } else {
    selectedDishes.value[dishId] = 1
  }
}

const decreaseQuantity = (dishId) => {
  if (selectedDishes.value[dishId] && selectedDishes.value[dishId] > 1) {
    selectedDishes.value[dishId]--
  } else {
    delete selectedDishes.value[dishId]
  }
}

const getDishName = (dishId) => {
  const dish = dishes.value.find(d => d.dish_id === dishId)
  return dish ? dish.dish_name : ''
}

const getDishPrice = (dishId) => {
  const dish = dishes.value.find(d => d.dish_id === dishId)
  return dish ? dish.price : 0
}
const getCategoryLabel = (category) => {
  return categoryMap[category] || category
}
// 加载食堂
const loadCanteens = async () => {
  try {
    const response = await dataAPI.getCanteens()
    canteens.value = response.data.data || response.data
  } catch (error) {
    console.error('加载食堂失败:', error)
    alert('加载食堂列表失败')
  }
}
const loadMerchantDetails = async () => {
  if (!form.value.merchant_id) return
  
  console.log('开始加载商家详情，商家ID:', form.value.merchant_id)
  console.log('所有商家:', merchants.value)

  // 找到当前选中的商家
  currentMerchant.value = merchants.value.find(m => m.merchant_id == form.value.merchant_id)
  
   console.log('找到的商家:', currentMerchant.value)
  
  if (currentMerchant.value) {
    console.log('商家详情:', {
      description: currentMerchant.value.description,
      business_hours: currentMerchant.value.business_hours,
      phone: currentMerchant.value.phone
    })
  }

  // 加载菜品
  await loadDishes()
}
// 加载商家
const loadMerchants = async () => {
  if (!form.value.canteen_id) return
  
  try {
    const response = await dataAPI.getMerchants(form.value.canteen_id)
    merchants.value = response.data.data || response.data
    console.log('商家数据:', merchants.value) // 测试
    form.value.merchant_id = ''
    dishes.value = []
    selectedDishes.value = {}
  } catch (error) {
    console.error('加载商家失败:', error)
    alert('加载商家列表失败')
  }
}
// 监听商家选择变化
watch(() => form.value.merchant_id, (newVal) => {
  if (newVal) {
    console.log('选中的商家ID:', newVal)
    loadMerchantDetails()
  }
})
// 加载菜品
const loadDishes = async () => {
  if (!form.value.merchant_id) return
  
  loadingDishes.value = true
  try {
    const response = await dataAPI.getDishes(form.value.merchant_id)
    dishes.value = response.data.data || response.data
    
    // 详细查看菜品数据结构
    console.log('菜品数据完整结构:', dishes.value)
    if (dishes.value.length > 0) {
      console.log('第一个菜品对象:', dishes.value[0])
      console.log('第一个菜品字段:', Object.keys(dishes.value[0]))
      console.log('是否有description字段:', 'description' in dishes.value[0])
      console.log('description值:', dishes.value[0].description)
      console.log('是否有category字段:', 'category' in dishes.value[0])
      console.log('是否有is_available字段:', 'is_available' in dishes.value[0])
    }
    
    selectedDishes.value = {}
  } catch (error) {
    console.error('加载菜品失败:', error)
    alert('加载菜品列表失败')
  } finally {
    loadingDishes.value = false
  }
}

// 提交订单
const handleSubmit = async () => {
  if (!canSubmit.value) return

  // 准备订单数据
  const orderData = {
    canteen_id: parseInt(form.value.canteen_id),
    merchant_id: parseInt(form.value.merchant_id),
    fetch_time: form.value.delivery_time.replace(' ', 'T') + ':00',
    tip: parseFloat(form.value.tip) || 0,
    delivery_address: form.value.delivery_address,
    contact_phone: form.value.contact_phone,
    order_details: selectedDishesList.value,
    notes: form.value.notes || '',
    urgency: form.value.urgency
  }

  console.log('提交代买需求:', orderData)

  loading.value = true
  try {
    // 使用现有的 orderAPI.create
    const response = await orderAPI.create(orderData)
    
    console.log('代买需求发布成功:', response.data)
    alert('代买需求发布成功！等待跑腿员接单～')
    emit('orderCreated')
    close()
    
    resetForm()
  } catch (error) {
    console.error('发布失败:', error)
    if (error.response?.data?.message) {
      alert('发布失败: ' + error.response.data.message)
    } else {
      alert('发布失败，请检查网络连接')
    }
  } finally {
    loading.value = false
  }
}

// 重置表单
const resetForm = () => {
  form.value = {
    canteen_id: '',
    merchant_id: '',
    delivery_time: '',
    delivery_address: '',
    contact_phone: '',
    tip: 0,
    custom_tip: '',
    notes: '',
    urgency: 'normal'
  }
  selectedDishes.value = {}
  dishes.value = []
  merchants.value = []
}

// 监听模态框显示/隐藏
watch(() => props.modelValue, (newVal) => {
  console.log('模态框显示状态变化:', newVal)
  if (newVal) {
    loadCanteens()
  } else {
    resetForm()
  }
})

onMounted(() => {
  if (props.modelValue) {
    loadCanteens()
  }
})
</script>

<style scoped>
/* 模态框基础样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 9999;
  overflow-y: auto;
  padding: 20px;
}

.modal-content {
  background: white;
  padding: 2rem;
  border-radius: 12px;
  width: 90%;
  max-width: 800px;
  max-height: 85vh;
  overflow-y: auto;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
  animation: modalFadeIn 0.3s ease;
  position: relative;
  margin: auto;
}

/* 模态框头部 */
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 2px solid #f0f0f0;
  position: sticky;
  top: 0;
  background: white;
  z-index: 10;
}

.modal-header h2 {
  margin: 0;
  color: #333;
  font-size: 1.5rem;
  font-weight: 600;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.8rem;
  cursor: pointer;
  color: #666;
  padding: 0.5rem;
  line-height: 1;
  transition: color 0.2s;
}

.btn-close:hover {
  color: #333;
  transform: scale(1.1);
}

/* 模态框动画 */
@keyframes modalFadeIn {
  from {
    opacity: 0;
    transform: translateY(-30px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* 表单容器 */
form {
  max-height: calc(85vh - 150px);
  overflow-y: auto;
  padding-right: 5px;
}

/* 自定义滚动条 */
form::-webkit-scrollbar {
  width: 6px;
}

form::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

form::-webkit-scrollbar-thumb {
  background: #888;
  border-radius: 3px;
}

form::-webkit-scrollbar-thumb:hover {
  background: #555;
}

/* 表单组 */
.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #333;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 2px solid #e0e0e0;
  border-radius: 6px;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  border-color: #007bff;
  outline: none;
  box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.1);
}

/* 操作按钮 */
.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 2px solid #f0f0f0;
  position: sticky;
  bottom: 0;
  background: white;
  z-index: 10;
}

.submit-btn,
.cancel-btn {
  flex: 1;
  padding: 1rem;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.submit-btn {
  background: linear-gradient(135deg, #007bff, #0056b3);
  color: white;
}

.submit-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #0056b3, #004085);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 91, 187, 0.3);
}

.submit-btn:disabled {
  background: #cccccc;
  cursor: not-allowed;
  transform: none;
}

.cancel-btn {
  background: #6c757d;
  color: white;
}

.cancel-btn:hover {
  background: #5a6268;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(108, 117, 125, 0.3);
}

/* 响应式调整 */
@media (max-width: 768px) {
  .modal-content {
    width: 95%;
    padding: 1.5rem;
    max-height: 90vh;
  }
  
  .form-actions {
    flex-direction: column;
  }
  
  .modal-header h2 {
    font-size: 1.3rem;
  }
}

/* 平台说明 */
.alert {
  border-radius: 8px;
  border: none;
}

.alert-info {
  background: linear-gradient(135deg, #e3f2fd, #bbdefb);
  border-left: 4px solid #1976d2;
}

.alert-heading {
  color: #1976d2;
  font-weight: 600;
}


</style>