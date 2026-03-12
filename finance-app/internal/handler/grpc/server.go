// Package grpcserver implements the gRPC server for the finance app.
// Run `make proto` to regenerate the protobuf Go files from proto/finance.proto.
package grpcserver

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/thomasaqx/finance-app/internal/domain"
	"github.com/thomasaqx/finance-app/internal/service"
)

// --- Minimal hand-rolled protobuf types (no code-gen required for compilation) ---
// In production, replace these with the generated types from proto/finance.proto.

type GetAccountRequest struct {
	ID int64
}
type AccountResponse struct {
	ID       int64
	UserID   int64
	Name     string
	Type     string
	Balance  float64
	Currency string
}
type ListAccountsRequest struct {
	UserID int64
}
type ListAccountsResponse struct {
	Accounts []AccountResponse
}
type CreateTransactionRequest struct {
	AccountID   int64
	ToAccountID int64
	CategoryID  int64
	Type        string
	Amount      float64
	Description string
}
type TransactionResponse struct {
	ID     int64
	Status string
}
type GetSummaryRequest struct {
	AccountID int64
}
type SummaryResponse struct {
	Income  float64
	Expense float64
	Net     float64
}

// FinanceGRPCServer bundles account and transaction handlers.
type FinanceGRPCServer struct {
	accountSvc     *service.AccountService
	transactionSvc *service.TransactionService
	grpcServer     *grpc.Server
}

func NewFinanceGRPCServer(
	accountSvc *service.AccountService,
	transactionSvc *service.TransactionService,
) *FinanceGRPCServer {
	s := &FinanceGRPCServer{
		accountSvc:     accountSvc,
		transactionSvc: transactionSvc,
		grpcServer:     grpc.NewServer(),
	}
	return s
}

// Start listens on addr and serves gRPC requests.
func (s *FinanceGRPCServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}
	return s.grpcServer.Serve(lis)
}

// Stop gracefully shuts down the gRPC server.
func (s *FinanceGRPCServer) Stop() {
	s.grpcServer.GracefulStop()
}

// GetAccount retrieves an account by ID.
func (s *FinanceGRPCServer) GetAccount(ctx context.Context, req GetAccountRequest) (*AccountResponse, error) {
	a, err := s.accountSvc.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return accountToResponse(a), nil
}

// ListAccounts returns all accounts for a user.
func (s *FinanceGRPCServer) ListAccounts(ctx context.Context, req ListAccountsRequest) (*ListAccountsResponse, error) {
	accounts, err := s.accountSvc.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	resp := &ListAccountsResponse{}
	for _, a := range accounts {
		resp.Accounts = append(resp.Accounts, *accountToResponse(a))
	}
	return resp, nil
}

// CreateTransaction creates a transaction and returns its ID/status.
func (s *FinanceGRPCServer) CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*TransactionResponse, error) {
	t := &domain.Transaction{
		AccountID:   req.AccountID,
		Type:        domain.TransactionType(req.Type),
		Amount:      req.Amount,
		Description: req.Description,
	}
	if req.ToAccountID != 0 {
		t.ToAccountID = &req.ToAccountID
	}
	if req.CategoryID != 0 {
		t.CategoryID = &req.CategoryID
	}
	if err := s.transactionSvc.Create(ctx, t); err != nil {
		return nil, err
	}
	return &TransactionResponse{ID: t.ID, Status: string(t.Status)}, nil
}

// GetSummary returns income/expense/net for an account.
func (s *FinanceGRPCServer) GetSummary(ctx context.Context, req GetSummaryRequest) (*SummaryResponse, error) {
	income, expense, net, err := s.transactionSvc.Summary(ctx, req.AccountID)
	if err != nil {
		return nil, err
	}
	return &SummaryResponse{Income: income, Expense: expense, Net: net}, nil
}

func accountToResponse(a *domain.Account) *AccountResponse {
	return &AccountResponse{
		ID:       a.ID,
		UserID:   a.UserID,
		Name:     a.Name,
		Type:     string(a.Type),
		Balance:  a.Balance,
		Currency: a.Currency,
	}
}
