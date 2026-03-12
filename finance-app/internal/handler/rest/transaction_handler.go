package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/thomasaqx/finance-app/internal/domain"
	"github.com/thomasaqx/finance-app/internal/service"
)

// TransactionHandler handles HTTP requests for transactions.
type TransactionHandler struct {
	svc *service.TransactionService
}

func NewTransactionHandler(svc *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

// RegisterRoutes wires up transaction routes.
func (h *TransactionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", h.Create)
	rg.POST("/async", h.CreateAsync)
	rg.GET("/:id", h.GetByID)
	rg.GET("", h.List)
	rg.GET("/summary/:account_id", h.Summary)
}

// Create godoc
// POST /api/v1/transactions
func (h *TransactionHandler) Create(c *gin.Context) {
	var t domain.Transaction
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Create(c.Request.Context(), &t); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

// CreateAsync godoc
// POST /api/v1/transactions/async — enqueues the transaction for background processing.
func (h *TransactionHandler) CreateAsync(c *gin.Context) {
	var t domain.Transaction
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resultCh := h.svc.CreateAsync(&t)
	// Wait for the worker result (demonstrates channel usage).
	if err := <-resultCh; err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, t)
}

// GetByID godoc
// GET /api/v1/transactions/:id
func (h *TransactionHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	t, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}

// List godoc
// GET /api/v1/transactions?account_id=X&type=expense&limit=20&offset=0
func (h *TransactionHandler) List(c *gin.Context) {
	accountID, _ := strconv.ParseInt(c.Query("account_id"), 10, 64)
	if accountID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id query param required"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filter := domain.TransactionFilter{
		AccountID: accountID,
		Type:      domain.TransactionType(c.Query("type")),
		Limit:     limit,
		Offset:    offset,
	}
	list, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// Summary godoc
// GET /api/v1/transactions/summary/:account_id
func (h *TransactionHandler) Summary(c *gin.Context) {
	accountID, err := strconv.ParseInt(c.Param("account_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account_id"})
		return
	}
	income, expense, net, err := h.svc.Summary(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"account_id": accountID,
		"income":     income,
		"expense":    expense,
		"net":        net,
	})
}
