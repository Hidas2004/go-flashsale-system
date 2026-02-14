package v1

import (
	"net/http"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
	"github.com/Hidas2004/go-flashsale-system/internal/domain/models"
	"github.com/Hidas2004/go-flashsale-system/internal/usecase"
	"github.com/Hidas2004/go-flashsale-system/pkg/common/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderHandler struct {
	orderUseCase usecase.OrderUseCase
}

func NewOrderHandler(orderUseCase usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{orderUseCase: orderUseCase}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req dtos.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu đầu vào không hợp lệ", err.Error())
		return
	}
	userIDVal, exists := c.Get("x-user-id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "Chưa đăng nhập", "User ID not found in context")
		return
	}
	userID := userIDVal.(uuid.UUID)
	resp, err := h.orderUseCase.CreateFlashSaleOrder(c.Request.Context(), userID, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Đặt hàng thất bại", err.Error())
		return
	}
	response.SuccessResponse(c, http.StatusAccepted, "Đơn hàng đã được tiếp nhận", resp)
}

// GetOrder (Chi tiết đơn hàng)
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ID đơn hàng không hợp lệ", err.Error())
		return
	}

	order, err := h.orderUseCase.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy đơn hàng", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Lấy thông tin đơn hàng thành công", order)
}

// Xem DANH SÁCH tất cả đơn hàng của người dùng.
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userIDVal, exists := c.Get("x-user-id")
	if !exists {
		response.ErrorResponse(c, http.StatusUnauthorized, "Chưa đăng nhập", "User ID not found")
		return
	}
	userID := userIDVal.(uuid.UUID)

	orders, err := h.orderUseCase.GetUserOrders(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Lỗi hệ thống", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Lấy danh sách đơn hàng thành công", gin.H{"orders": orders})
}

func (h *OrderHandler) AdminUpdateStatus(c *gin.Context) {
	// 1. Lấy ID từ URL
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ID đơn hàng không hợp lệ", err.Error())
		return
	}
	// 2. Bind JSON Body
	var req dtos.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}
	// 3. Gọi UseCase
	if err := h.orderUseCase.UpdateOrderStatus(c.Request.Context(), id, models.OrderStatus(req.Status)); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Cập nhật trạng thái thất bại", err.Error())
		return
	}
	response.SuccessResponse(c, http.StatusOK, "Cập nhật trạng thái thành công", nil)
}
func (h *OrderHandler) AdminListOrders(c *gin.Context) {
	// 1. Phân trang (Lấy từ Query Param ?page=1&limit=10)
	page := 1
	limit := 10
	// 2. Gọi UseCase
	orders, total, err := h.orderUseCase.ListOrders(c.Request.Context(), page, limit)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Lỗi lấy danh sách đơn hàng", err.Error())
		return
	}
	response.SuccessResponse(c, http.StatusOK, "Thành công", gin.H{
		"orders": orders,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}
