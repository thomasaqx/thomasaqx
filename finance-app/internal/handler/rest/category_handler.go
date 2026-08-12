package rest

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/thomasaqx/finance-app/internal/domain"
	"github.com/thomasaqx/finance-app/internal/service"
)

// CategoryHandler handles HTTP requests for categories and budgets.
type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// RegisterRoutes wires up category routes.
func (h *CategoryHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", h.CreateCategory)
	rg.GET("", h.GetByUser)
}

// RegisterBudgetRoutes wires up budget routes.
func (h *CategoryHandler) RegisterBudgetRoutes(rg *gin.RouterGroup) {
	rg.POST("", h.CreateBudget)
	rg.GET("", h.GetBudgets)
}

// CreateCategory godoc
// POST /api/v1/categories
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var cat domain.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateCategory(c.Request.Context(), &cat); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

// GetByUser godoc
// GET /api/v1/categories?user_id=X
func (h *CategoryHandler) GetByUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id query param required"})
		return
	}
	cats, err := h.svc.GetCategoriesByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cats)
}

// CreateBudget godoc
// POST /api/v1/budgets
func (h *CategoryHandler) CreateBudget(c *gin.Context) {
	var b domain.Budget
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateBudget(c.Request.Context(), &b); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, b)
}

// GetBudgets godoc
// GET /api/v1/budgets?user_id=X&year=2024&month=3
func (h *CategoryHandler) GetBudgets(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)
	year, _ := strconv.Atoi(c.Query("year"))
	month, _ := strconv.Atoi(c.Query("month"))

	if userID == 0 || year == 0 || month == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id, year, and month are required"})
		return
	}
	budgets, err := h.svc.GetBudgets(c.Request.Context(), userID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, budgets)
}
