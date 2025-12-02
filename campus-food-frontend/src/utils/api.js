// utils/api.js
import axios from 'axios'

const api = axios.create({
  baseURL: 'http://localhost:8080/api',
  timeout: 10000
})

// 请求拦截器 - 自动添加 token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理认证错误
api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    }
    return Promise.reject(error)
  }
)

// 认证相关 API
export const authAPI = {
  login: (data) => api.post('/auth/login', data),
  register: (data) => api.post('/auth/register', data),
}

// 订单相关 API 
export const orderAPI = {
  // 获取订单列表
  list: (status = 'all') => {
    const params = {}
    if (status !== 'all') {
      params.status = status
    }
    return api.get('/orders', { params })
  },
  
  // 创建订单
  create: (data) => api.post('/orders', data),
  
  // 获取完整订单详情（带菜品信息）
  getFullDetails: (orderId) => api.get(`/orders/${orderId}`),
  
  // 获取简单订单详情
  getSimpleDetails: (orderId) => api.get(`/orders/${orderId}/simple`),
  
  // 接单
  take: (orderId) => api.post(`/orders/${orderId}/take`),
  
  // 完成订单
  complete: (orderId) => api.post(`/orders/${orderId}/complete`),
  
  // 取消订单
  cancel: (orderId, reason = '用户取消') => {
    return api.post(`/orders/${orderId}/cancel`, null, {
      params: { reason }
    })
  },
  
  // 通用状态更新（使用 PUT /orders/status）
  updateStatus: (orderId, status, reason = '') => {
    return api.put('/orders/status', {
      order_id: orderId,
      status: status,
      reason: reason
    })
  },
  
  // 获取可用的状态选项
  getAvailableStatusOptions: (orderId) => {
    return api.get(`/orders/${orderId}/available-status`)
  },
  
  // 获取订单统计
  getStats: () => api.get('/orders/stats'),
  
  // 获取进行中的订单
  getRunningOrders: () => api.get('/orders/running'),
  
  // 获取可接订单（跑腿员用）
  getAvailableOrders: () => api.get('/orders/available')
}

// 数据相关 API（食堂、商家、菜品）
export const dataAPI = {
  getCanteens: () => api.get('/data/canteens'),
  getMerchants: (canteenId) => api.get(`/data/merchants?canteen_id=${canteenId}`),
  getDishes: (merchantId) => api.get(`/data/dishes?merchant_id=${merchantId}`),
}

// 投诉相关 API
export const complaintAPI = {
  create: (data) => api.post('/complaints', data),
  list: () => api.get('/complaints'),
}

export default api