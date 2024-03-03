package domain

import (
	"context"
	"errors"

	"github.com/hickar/crtex_test_assignment/events"
)

type Account struct {
	ID          int64
	UserID      int64
	AmountCents int64
}

type Service interface {
	ProcessNewOrder(context.Context, events.OrderCreatedEvent) error
}

type AccountRepository interface {
	AccountEventWithOrderEventIDExists(context.Context, int64) (bool, error)
	GetAccountByUserID(context.Context, int64) (Account, error)
	UpdateAccount(context.Context, Account) error
	CreateAccountEvent(context.Context, events.AccountOrderPaymentEvent) error
	WithinTransaction(context.Context, func(context.Context) error) error
}

type AccountService struct {
	repo AccountRepository
}

func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) ProcessNewOrder(ctx context.Context, orderEvent events.OrderCreatedEvent) error {
	if orderEvent.AmountCents < 0 {
		return ErrInvalidData
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
			return ErrAlreadyProcessed
		}

		account, err := s.repo.GetAccountByUserID(tctx, orderEvent.UserID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return s.cancelAccountPayment(tctx, orderEvent, account)
			}

			return err
		}

		newAmount := account.AmountCents - orderEvent.AmountCents
		if newAmount < 0 {
			return s.cancelAccountPayment(tctx, orderEvent, account)
		}

		account.AmountCents = newAmount
		if err = s.repo.UpdateAccount(tctx, account); err != nil {
			return err
		}

		if err = s.repo.CreateAccountEvent(tctx, events.AccountOrderPaymentEvent{
			AccountID:    account.ID,
			OrderID:      orderEvent.OrderID,
			OrderEventID: orderEvent.ID,
			Status:       events.AccountOrderStatusPaid,
		}); err != nil {
			return err
		}

		return nil
	})
}

func (s *AccountService) cancelAccountPayment(ctx context.Context, event events.OrderCreatedEvent, account Account) error {
	return s.repo.CreateAccountEvent(ctx, events.AccountOrderPaymentEvent{
		AccountID:    account.ID,
		OrderID:      event.OrderID,
		OrderEventID: event.ID,
		Status:       events.AccountOrderStatusCanceled,
	})
}
