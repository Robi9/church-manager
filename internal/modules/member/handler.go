package member

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Robi9/church-manager/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	var m Member

	if err := c.ShouldBindJSON(&m); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	userIDFloat := userIDValue.(float64)
	m.CreatedBy = int64(userIDFloat)

	result, err := h.service.Create(m)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err)
		return
	}

	response.Success(c, http.StatusCreated, result)
}

func (h *Handler) Find(c *gin.Context) {
	name := c.Query("name")
	status := c.Query("status")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	filters := map[string]string{
		"name":   name,
		"status": status,
	}

	data, meta, err := h.service.Find(filters, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"data": data,
		"meta": meta,
	})
}
