package v1

import (
	"net/http"

	"github.com/Hidas2004/go-flashsale-system/internal/domain/dtos"
	"github.com/Hidas2004/go-flashsale-system/internal/usecase"
	"github.com/Hidas2004/go-flashsale-system/pkg/common/response"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthHandler(authUseCase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dtos.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid Input", err.Error())
		return
	}
	//usecase trả về token và user
	authResp, err := h.authUseCase.Register(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "Registration Failed", err.Error())
		return
	}
	response.SuccessResponse(c, http.StatusCreated, "Đăng ký thành công", authResp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dtos.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "Invalid Input", err.Error())
		return
	}

	// UseCase đã trả về cả Token và User
	authResp, err := h.authUseCase.Login(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponse(c, http.StatusUnauthorized, "Đăng nhập thất bại", err.Error())
		return
	}

	response.SuccessResponse(c, http.StatusOK, "Đăng nhập thành công", authResp)
}
