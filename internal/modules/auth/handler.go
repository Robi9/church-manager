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

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(c *gin.Context) {
	var req authRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	user, err := h.service.Register(req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	response.Success(c, http.StatusCreated, user)
}

func (h *Handler) Login(c *gin.Context) {
	var req authRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	token, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"token": token,
	})
}
