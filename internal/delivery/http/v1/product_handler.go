package v1

import (
	"fmt"
	"net/http"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
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

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req dtos.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu không hợp lệ", err.Error())
		return
	}
	//gọi usecase
	product, err := h.productUseCase.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		fmt.Println("❌ Error creating product:", err) // [DEBUG] In lỗi ra terminal
		response.ErrorResponse(c, http.StatusInternalServerError, "Lỗi khi tạo sản phẩm", err.Error())
		return
	}
	response.SuccessResponse(c, http.StatusCreated, "Tạo sản phẩm thành công", product)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	// [1] Parse ID từ URL param
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ID sản phẩm không hợp lệ", err.Error())
		return
	}
	// [2] Bind JSON body
	var req dtos.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Dữ liệu cập nhật không hợp lệ", err.Error())
		return
	}
	// [3] Call Business Logic
	updatedProduct, err := h.productUseCase.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Lỗi cập nhật sản phẩm", err.Error())
		return
	}
	response.SuccessResponse(c, http.StatusOK, "Cập nhật thành công", updatedProduct)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "ID sản phẩm không hợp lệ", err.Error())
		return
	}
	if err := h.productUseCase.DeleteProduct(c.Request.Context(), id); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Lỗi xóa sản phẩm", err.Error())
		return
	}
	response.SuccessResponse(c, http.StatusOK, "Xóa sản phẩm thành công", nil)
}
