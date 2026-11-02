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
		if err.Error() == "invalid credentials" {
			response.Error(c, http.StatusUnauthorized, err)
		} else {
			response.Error(c, http.StatusInternalServerError, err)
		}
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"token":      token,
		"expires_in": int(SessionLifetime.Seconds()),
	})
}

func (h *Handler) Me(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{"user_id": c.GetInt64("user_id")})
}

func (h *Handler) Logout(c *gin.Context) {
	if err := h.service.repo.RevokeSession(c.Request.Context(), c.GetString("session_hash")); err != nil {
		response.Error(c, 500, err)
		return
	}
	response.Success(c, 200, gin.H{"message": "signed out"})
}

func (h *Handler) LogoutAll(c *gin.Context) {
	if err := h.service.repo.RevokeAllSessions(c.Request.Context(), c.GetInt64("user_id")); err != nil {
		response.Error(c, 500, err)
		return
	}
	response.Success(c, 200, gin.H{"message": "all sessions revoked"})
}
