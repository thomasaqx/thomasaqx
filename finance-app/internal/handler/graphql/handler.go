package graphql

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"

	"github.com/thomasaqx/finance-app/internal/domain"
	"github.com/thomasaqx/finance-app/internal/service"
)

// Handler exposes a GraphQL endpoint via Gin.
type Handler struct {
	schema graphql.Schema
}

func NewHandler(accountSvc *service.AccountService, txSvc *service.TransactionService) (*Handler, error) {
	schema, err := buildSchema(accountSvc, txSvc)
	if err != nil {
		return nil, err
	}
	return &Handler{schema: schema}, nil
}

// Register wires the /graphql route.
func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.POST("/graphql", h.Handle)
	rg.GET("/graphql", h.Handle)
}

func (h *Handler) Handle(c *gin.Context) {
	var params struct {
		Query     string                 `json:"query" form:"query"`
		Variables map[string]interface{} `json:"variables"`
	}
	if err := c.ShouldBind(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result := graphql.Do(graphql.Params{
		Schema:         h.schema,
		RequestString:  params.Query,
		VariableValues: params.Variables,
		Context:        c.Request.Context(),
	})
	c.JSON(http.StatusOK, result)
}

func buildSchema(accountSvc *service.AccountService, txSvc *service.TransactionService) (graphql.Schema, error) {
	accountType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Account",
		Fields: graphql.Fields{
			"id":       &graphql.Field{Type: graphql.Int},
			"user_id":  &graphql.Field{Type: graphql.Int},
			"name":     &graphql.Field{Type: graphql.String},
			"type":     &graphql.Field{Type: graphql.String},
			"balance":  &graphql.Field{Type: graphql.Float},
			"currency": &graphql.Field{Type: graphql.String},
		},
	})

	summaryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Summary",
		Fields: graphql.Fields{
			"income":  &graphql.Field{Type: graphql.Float},
			"expense": &graphql.Field{Type: graphql.Float},
			"net":     &graphql.Field{Type: graphql.Float},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"account": {
				Type:        accountType,
				Description: "Get account by ID",
				Args: graphql.FieldConfigArgument{
					"id": {Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(int)
					return accountSvc.GetByID(p.Context, int64(id))
				},
			},
			"accounts": {
				Type:        graphql.NewList(accountType),
				Description: "List accounts for a user",
				Args: graphql.FieldConfigArgument{
					"user_id": {Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, _ := p.Args["user_id"].(int)
					return accountSvc.GetByUserID(p.Context, int64(userID))
				},
			},
			"summary": {
				Type:        summaryType,
				Description: "Get income/expense/net summary for an account",
				Args: graphql.FieldConfigArgument{
					"account_id": {Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					aid, _ := p.Args["account_id"].(int)
					income, expense, net, err := txSvc.Summary(p.Context, int64(aid))
					if err != nil {
						return nil, err
					}
					return map[string]float64{"income": income, "expense": expense, "net": net}, nil
				},
			},
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createAccount": {
				Type:        accountType,
				Description: "Create a new account",
				Args: graphql.FieldConfigArgument{
					"user_id":  {Type: graphql.NewNonNull(graphql.Int)},
					"name":     {Type: graphql.NewNonNull(graphql.String)},
					"type":     {Type: graphql.NewNonNull(graphql.String)},
					"currency": {Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					a := &domain.Account{
						UserID:   int64(p.Args["user_id"].(int)),
						Name:     p.Args["name"].(string),
						Type:     domain.AccountType(p.Args["type"].(string)),
						Currency: "BRL",
					}
					if cur, ok := p.Args["currency"].(string); ok && cur != "" {
						a.Currency = cur
					}
					if err := accountSvc.Create(p.Context, a); err != nil {
						return nil, err
					}
					return a, nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
}
