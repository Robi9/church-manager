package auth

import (
	"net/http"

	"github.com/Robi9/church-manager/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// bind do body
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	// chama service
	token, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err)
		return
	}

	// sucesso
	response.Success(c, http.StatusOK, gin.H{
		"token": token,
	})
}
