package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/thomasaqx/finance-app/internal/domain"
)

// TransactionJob encapsulates the data needed to process a transaction asynchronously.
type TransactionJob struct {
	Transaction *domain.Transaction
	Result      chan<- error
}

// TransactionService implements business logic for transactions.
// It uses a buffered channel + goroutine pool to process transactions concurrently.
type TransactionService struct {
	txRepo      domain.TransactionRepository
	accountRepo domain.AccountRepository
	cache       domain.CacheRepository
	publisher   domain.EventPublisher

	jobQueue chan TransactionJob
	wg       sync.WaitGroup
}

func NewTransactionService(
	txRepo domain.TransactionRepository,
	accountRepo domain.AccountRepository,
	cache domain.CacheRepository,
	publisher domain.EventPublisher,
	workers int,
) *TransactionService {
	s := &TransactionService{
		txRepo:      txRepo,
		accountRepo: accountRepo,
		cache:       cache,
		publisher:   publisher,
		jobQueue:    make(chan TransactionJob, 100),
	}
	s.startWorkers(workers)
	return s
}

// startWorkers spins up a goroutine pool that drains the job queue.
func (s *TransactionService) startWorkers(n int) {
	for i := 0; i < n; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for job := range s.jobQueue {
				job.Result <- s.processTransaction(context.Background(), job.Transaction)
			}
		}()
	}
}

// Stop shuts the worker pool down gracefully.
func (s *TransactionService) Stop() {
	close(s.jobQueue)
	s.wg.Wait()
}

// CreateAsync enqueues a transaction for background processing.
// Returns a channel the caller can read to get the result.
func (s *TransactionService) CreateAsync(t *domain.Transaction) <-chan error {
	result := make(chan error, 1)
	s.jobQueue <- TransactionJob{Transaction: t, Result: result}
	return result
}

// Create creates a transaction synchronously.
func (s *TransactionService) Create(ctx context.Context, t *domain.Transaction) error {
	return s.processTransaction(ctx, t)
}

func (s *TransactionService) processTransaction(ctx context.Context, t *domain.Transaction) error {
	if err := t.Validate(); err != nil {
		return fmt.Errorf("validation: %w", err)
	}

	t.Status = domain.TransactionStatusPending
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	account, err := s.accountRepo.GetByID(ctx, t.AccountID)
	if err != nil {
		t.MarkFailed()
		return fmt.Errorf("get account: %w", err)
	}

	switch t.Type {
	case domain.TransactionTypeExpense:
		if err := account.Debit(t.Amount); err != nil {
			t.MarkFailed()
			return err
		}
	case domain.TransactionTypeIncome:
		if err := account.Credit(t.Amount); err != nil {
			t.MarkFailed()
			return err
		}
	case domain.TransactionTypeTransfer:
		if err := account.Debit(t.Amount); err != nil {
			t.MarkFailed()
			return err
		}
		toAccount, err := s.accountRepo.GetByID(ctx, *t.ToAccountID)
		if err != nil {
			t.MarkFailed()
			return fmt.Errorf("get destination account: %w", err)
		}
		_ = toAccount.Credit(t.Amount)
		if err := s.accountRepo.UpdateBalance(ctx, toAccount.ID, toAccount.Balance); err != nil {
			t.MarkFailed()
			return fmt.Errorf("update destination balance: %w", err)
		}
	}

	if err := s.accountRepo.UpdateBalance(ctx, account.ID, account.Balance); err != nil {
		t.MarkFailed()
		return fmt.Errorf("update source balance: %w", err)
	}

	if err := s.txRepo.Create(ctx, t); err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	t.MarkCompleted()
	_ = s.txRepo.Update(ctx, t)
	_ = s.cache.Delete(ctx, cacheKeyAccount(t.AccountID))
	_ = s.publisher.Publish(ctx, "transaction.created", t)
	return nil
}

func (s *TransactionService) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	return s.txRepo.GetByID(ctx, id)
}

func (s *TransactionService) List(ctx context.Context, filter domain.TransactionFilter) ([]*domain.Transaction, error) {
	if filter.Limit == 0 {
		filter.Limit = 20
	}
	return s.txRepo.List(ctx, filter)
}

// Summary returns income, expense, and net balance for an account concurrently.
func (s *TransactionService) Summary(ctx context.Context, accountID int64) (income, expense, net float64, err error) {
	type result struct {
		value float64
		err   error
	}

	incomeCh := make(chan result, 1)
	expenseCh := make(chan result, 1)

	go func() {
		v, e := s.txRepo.SumByType(ctx, accountID, domain.TransactionTypeIncome)
		incomeCh <- result{v, e}
	}()

	go func() {
		v, e := s.txRepo.SumByType(ctx, accountID, domain.TransactionTypeExpense)
		expenseCh <- result{v, e}
	}()

	ir := <-incomeCh
	er := <-expenseCh

	if ir.err != nil {
		return 0, 0, 0, ir.err
	}
	if er.err != nil {
		return 0, 0, 0, er.err
	}
	return ir.value, er.value, ir.value - er.value, nil
}
