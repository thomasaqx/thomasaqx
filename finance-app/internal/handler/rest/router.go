package rest

import (
	"github.com/gin-gonic/gin"

	gqlhandler "github.com/thomasaqx/finance-app/internal/handler/graphql"
	"github.com/thomasaqx/finance-app/internal/service"
)

// NewRouter builds the Gin engine with all routes registered.
func NewRouter(
	accountSvc *service.AccountService,
	txSvc *service.TransactionService,
	catSvc *service.CategoryService,
) (*gin.Engine, error) {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		NewAccountHandler(accountSvc).RegisterRoutes(api.Group("/accounts"))
		NewTransactionHandler(txSvc).RegisterRoutes(api.Group("/transactions"))
		catHandler := NewCategoryHandler(catSvc)
		catHandler.RegisterRoutes(api.Group("/categories"))
		catHandler.RegisterBudgetRoutes(api.Group("/budgets"))
	}

	// GraphQL endpoint
	gql, err := gqlhandler.NewHandler(accountSvc, txSvc)
	if err != nil {
		return nil, err
	}
	gql.Register(api)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r, nil
}
