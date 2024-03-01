package domain

import (
	"context"
	"errors"
)

type Account struct {
	ID          int64
	UserID      int64
	AmountCents int64
}

type AccountEvent struct {
	ID           int64
	OrderEventID int64
	AccountID    int64
	OrderID      int64
	Status       string
}

type IncomingOrderEvent struct {
	ID          int64
	OrderID     int64
	Status      string
	AmountCents int64
	UserID      int64
}

const (
	AccountOrderStatusCanceled = "CANCELED"
	AccountOrderStatusPaid     = "PAID"
)

type Service interface {
	ProcessOrder(context.Context, IncomingOrderEvent) error
}

type AccountRepository interface {
	AccountEventWithOrderEventIDExists(context.Context, int64) (bool, error)
	GetAccountByUserID(context.Context, int64) (Account, error)
	UpdateAccount(context.Context, Account) error
	CreateAccountEvent(context.Context, AccountEvent) error
	WithinTransaction(context.Context, func(context.Context) error) error
}

type AccountService struct {
	repo AccountRepository
}

func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) ProcessOrder(ctx context.Context, orderEvent IncomingOrderEvent) error {
	if orderEvent.AmountCents < 0 {
		return errors.New("amount must be a positive number")
	}

	return s.repo.WithinTransaction(ctx, func(tctx context.Context) error {
		// Требуется дополнительная проверка на существование записи об обработке
		// данного заказа. Событие о заказе может быть доставлено дважды, ввиду
		// возможного выхода из строя Debezium'a.
		exists, err := s.repo.AccountEventWithOrderEventIDExists(tctx, orderEvent.ID)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}

		account, err := s.repo.GetAccountByUserID(tctx, orderEvent.UserID)
		if err != nil {
			return err
		}

		newAmount := account.AmountCents - orderEvent.AmountCents
		if newAmount < 0 {
			return s.repo.CreateAccountEvent(tctx, AccountEvent{
				AccountID: account.ID,
				OrderID:   orderEvent.OrderID,
				Status:    AccountOrderStatusCanceled,
			})
		}

		account.AmountCents = newAmount
		if err = s.repo.UpdateAccount(tctx, account); err != nil {
			return err
		}

		if err = s.repo.CreateAccountEvent(tctx, AccountEvent{
			AccountID: account.ID,
			OrderID:   orderEvent.OrderID,
			Status:    AccountOrderStatusPaid,
		}); err != nil {
			return err
		}

		return nil
	})
}
