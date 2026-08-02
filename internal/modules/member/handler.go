package member

import (
	"encoding/csv"
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

func (h *Handler) DownloadTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv")
	c.Header(
		"Content-Disposition",
		`attachment; filename="modelo_membros.csv"`,
	)

	csvContent := `Nome,Telefone,Status,Membro desde,Batizado,Data do batismo,Cargo na igreja,Estado civil,Congregação,Igreja de origem,Fez o curso de membresia,Data do curso de membresia,Contactado no WhatsApp,Data do contato,Endereço,Número,Complemento,Bairro,Cidade,Estado
Maria,88999999999,Ativo,2024-01-01,Sim,2024-02-01,Líder,Solteiro,Sede,Batista,Sim,2024-02-15,Sim,2024-03-01,Rua das Flores,123,Apto 1,Centro,Fortaleza,CE
`

	c.String(http.StatusOK, csvContent)
}

func (h *Handler) DownloadImportErrors(c *gin.Context) {
	jobID := c.Param("jobID")

	job, ok := GetImportJob(jobID)
	if !ok {
		c.JSON(404, gin.H{
			"error": "import job not found",
		})
		return
	}

	c.Header(
		"Content-Type",
		"text/csv",
	)

	c.Header(
		"Content-Disposition",
		"attachment; filename=import_errors.csv",
	)

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	headers := append(job.Headers, "erro")

	err := writer.Write(headers)
	if err != nil {
		return
	}

	for _, errRow := range job.ErrorRows {
		row := normalizeRow(
			errRow.Data,
			len(job.Headers),
		)

		row = append(
			row,
			errRow.Error,
		)

		err := writer.Write(row)
		if err != nil {
			return
		}
	}
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	member, err := h.service.GetByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, errors.New("member not found"))
		return
	}

	response.Success(c, http.StatusOK, member)
}
