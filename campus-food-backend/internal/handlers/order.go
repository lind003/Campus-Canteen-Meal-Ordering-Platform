package handlers

import (
	"bytes"
	"campus-food-backend/internal/models"
	"campus-food-backend/internal/services"
	"campus-food-backend/internal/utils"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder 创建订单
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt("user_id")
	order, err := h.orderService.CreateOrder(userID, req)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "订单创建成功", order)
}

// GetOrders 获取订单列表
func (h *OrderHandler) GetOrders(c *gin.Context) {
	status := c.DefaultQuery("status", "all")
	userID := c.GetInt("user_id")

	orders, err := h.orderService.GetOrders(userID, status)
	if err != nil {
		utils.InternalError(c, "获取订单列表失败")
		return
	}

	utils.Success(c, orders)
}

// GetFullOrderDetails 获取完整订单详情
func (h *OrderHandler) GetFullOrderDetails(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, "订单ID格式错误")
		return
	}

	orderWithDetails, err := h.orderService.GetOrderDetails(orderID)
	if err != nil {
		utils.Error(c, "获取订单详情失败: "+err.Error())
		return
	}

	utils.Success(c, orderWithDetails)
}

// GetOrderDetails 获取简单订单详情
func (h *OrderHandler) GetOrderDetails(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, "订单ID格式错误")
		return
	}

	order, err := h.orderService.GetOrderByID(orderID)
	if err != nil {
		utils.Error(c, "获取订单详情失败: "+err.Error())
		return
	}

	utils.Success(c, order)
}

// TakeOrder 接单
func (h *OrderHandler) TakeOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, "订单ID格式错误")
		return
	}

	runnerID := c.GetInt("user_id")
	err = h.orderService.TakeOrder(orderID, runnerID)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "接单成功", nil)
}

// CompleteOrder 完成订单
func (h *OrderHandler) CompleteOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, "订单ID格式错误")
		return
	}

	runnerID := c.GetInt("user_id")
	err = h.orderService.CompleteOrder(orderID, runnerID)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "订单完成成功", nil)
}

// CancelOrder 取消订单
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, "订单ID格式错误")
		return
	}

	reason := c.DefaultQuery("reason", "用户取消")
	userID := c.GetInt("user_id")

	err = h.orderService.CancelOrder(userID, orderID, reason)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "订单取消成功", nil)
}

// UpdateOrderStatus 更新订单状态
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, "参数错误: "+err.Error())
		return
	}

	userID := c.GetInt("user_id")
	err := h.orderService.UpdateOrderStatus(userID, req)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "订单状态更新成功", nil)
}

// GetAvailableStatusOptions 获取可用状态选项
func (h *OrderHandler) GetAvailableStatusOptions(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, "订单ID格式错误")
		return
	}

	userID := c.GetInt("user_id")
	options, err := h.orderService.GetAvailableStatusOptions(userID, orderID)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"available_statuses": options,
	})
}

// GetMyRunningOrders 获取进行中的订单
func (h *OrderHandler) GetMyRunningOrders(c *gin.Context) {
	userID := c.GetInt("user_id")

	orders, err := h.orderService.GetMyRunningOrders(userID)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, orders)
}

// GetAvailableOrders 获取可接订单
func (h *OrderHandler) GetAvailableOrders(c *gin.Context) {
	userID := c.GetInt("user_id")

	orders, err := h.orderService.GetAvailableOrders(userID)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, orders)
}

// GetOrderStats 获取订单统计
func (h *OrderHandler) GetOrderStats(c *gin.Context) {
	userID := c.GetInt("user_id")

	stats, err := h.orderService.GetOrderStats(userID)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	utils.Success(c, stats)
}

// 辅助函数获取请求体
func getRequestBody(c *gin.Context) string {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "无法读取请求体"
	}
	// 重新设置请求体，以便后续使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	return string(body)
}
