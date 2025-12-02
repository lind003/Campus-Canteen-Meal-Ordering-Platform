package handlers

import (
	//"bytes"
	"campus-food-backend/internal/models"
	"campus-food-backend/internal/services"
	"campus-food-backend/internal/utils"
	"fmt"

	//"io"

	"github.com/gin-gonic/gin"
)

type ComplaintHandler struct {
	complaintService *services.ComplaintService
}

func NewComplaintHandler(complaintService *services.ComplaintService) *ComplaintHandler {
	return &ComplaintHandler{
		complaintService: complaintService,
	}
}

// CreateComplaint 创建投诉
// @Summary 创建投诉
// @Description 创建订单投诉
// @Tags 投诉
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateComplaintRequest true "投诉信息"
// @Success 200 {object} utils.Response
// @Router /complaints [post]
func (h *ComplaintHandler) CreateComplaint(c *gin.Context) {
	var req models.CreateComplaintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("CreateComplaint - 参数绑定错误: %v\n", err)
		fmt.Printf("CreateComplaint - 请求体: %s\n", getRequestBody(c))
		utils.Error(c, "参数错误: "+err.Error())
		return
	}

	fmt.Printf("CreateComplaint - 接收到的请求: %+v\n", req)

	complainerID := c.GetInt("user_id")
	complaint, err := h.complaintService.CreateComplaint(complainerID, req)
	if err != nil {
		fmt.Printf("CreateComplaint - 服务层错误: %v\n", err)
		utils.Error(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "投诉提交成功", complaint)
}

// GetComplaints 获取投诉列表
// @Summary 获取投诉列表
// @Description 获取投诉列表（管理员）
// @Tags 投诉
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /complaints [get]
func (h *ComplaintHandler) GetComplaints(c *gin.Context) {
	complaints, err := h.complaintService.GetComplaints()
	if err != nil {
		utils.InternalError(c, "获取投诉列表失败")
		return
	}

	utils.Success(c, complaints)
}
