package v1

import (
	"net/http"

	"github.com/Hidas2004/go-flashsale-system/internal/usecase"
	"github.com/Hidas2004/go-flashsale-system/pkg/common/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	productUseCase usecase.ProductUseCase
}

func NewProductHandler(productUseCase usecase.ProductUseCase) *ProductHandler {
	return &ProductHandler{productUseCase: productUseCase}
}

func (h *ProductHandler) GetFlashSaleProducts(c *gin.Context) {
	products, err := h.productUseCase.FindFlashSaleProducts(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Lỗi lấy danh sách sản phẩm", err.Error())
		return
	}
	response.SuccessResponse(c, http.StatusOK, "Thành công", gin.H{"products": products})
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ID không hợp lệ", err.Error())
		return
	}

	product, err := h.productUseCase.FindByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "Không tìm thấy sản phẩm", err.Error())
		return
	}

	// stock, _ := h.productUseCase.GetProductStock(c.Request.Context(), id)
	response.SuccessResponse(c, http.StatusOK, "Thành công", gin.H{
		"product":       product,
		"current_stock": 0, // TODO: Implement GetProductStock
	})
}
