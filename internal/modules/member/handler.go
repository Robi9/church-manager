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

func (h *Handler) Update(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.New("invalid member id"))
		return
	}

	var payload Member

	if err := c.ShouldBindJSON(&payload); err != nil {
		response.Error(c, http.StatusBadRequest, err)
		return
	}

	result, err := h.service.Update(id, payload)
	if err != nil {
		response.Error(c, http.StatusNotFound, err)
		return
	}

	response.Success(c, http.StatusOK, result)
}

func (h *Handler) Delete(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.New("invalid member id"))
		return
	}

	err = h.service.SoftDelete(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"message": "member deactivated successfully",
	})
}

func (h *Handler) Import(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.New("file is required"))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	userIDValue, _ := c.Get("user_id")
	userID := int64(userIDValue.(float64))

	result, err := h.service.ImportCSV(file, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err)
		return
	}

	response.Success(c, http.StatusOK, result)
}
