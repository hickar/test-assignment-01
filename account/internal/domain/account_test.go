package domain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hickar/crtex_test_assignment/events"
)

func TestProcessNewOrder(t *testing.T) {
	var (
		expectedAccountEventStatus = events.AccountOrderStatusPaid
		actualAccountEventStatus   events.AccountOrderPaymentStatus

		expectedNewAmount = int64(100000 - 1000)
		actualNewAmount   int64
	)

	repo := newAccountRepoStub(
		func(_ context.Context, _ int64) (bool, error) {
			return false, nil
		},
		func(_ context.Context, userID int64) (Account, error) {
			return Account{
				UserID:      userID,
				AmountCents: 100000,
			}, nil
		},
		func(_ context.Context, account Account) error {
			actualNewAmount = account.AmountCents
			return nil
		},
		func(_ context.Context, event events.AccountOrderPaymentEvent) error {
			actualAccountEventStatus = event.Status
			return nil
		},
		nil,
	)

	orderEvent := events.OrderCreatedEvent{
		AmountCents: 1000,
	}
	service := NewAccountService(repo)
	err := service.ProcessNewOrder(context.Background(), orderEvent)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccountEventStatus, actualAccountEventStatus)
	assert.Equal(t, expectedNewAmount, actualNewAmount)
}

func TestProcessNewOrderWithNonExistentUser(t *testing.T) {
	var actualAccountEventStatus events.AccountOrderPaymentStatus
	expectedAccountEventStatus := events.AccountOrderStatusCanceled

	repo := newAccountRepoStub(
		func(_ context.Context, _ int64) (bool, error) {
			return false, nil
		},
		func(_ context.Context, _ int64) (Account, error) {
			return Account{}, ErrNotFound
		},
		nil,
		func(_ context.Context, event events.AccountOrderPaymentEvent) error {
			actualAccountEventStatus = event.Status
			return nil
		},
		nil,
	)

	orderEvent := events.OrderCreatedEvent{}
	service := NewAccountService(repo)
	err := service.ProcessNewOrder(context.Background(), orderEvent)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccountEventStatus, actualAccountEventStatus)
}

func TestProcessNewOrderWithNotEnoughAmount(t *testing.T) {
	var actualAccountEventStatus events.AccountOrderPaymentStatus
	expectedAccountEventStatus := events.AccountOrderStatusCanceled

	repo := newAccountRepoStub(
		func(_ context.Context, _ int64) (bool, error) {
			return false, nil
		},
		func(_ context.Context, _ int64) (Account, error) {
			return Account{
				AmountCents: 1,
			}, ErrNotFound
		},
		nil,
		func(_ context.Context, event events.AccountOrderPaymentEvent) error {
			actualAccountEventStatus = event.Status
			return nil
		},
		nil,
	)

	orderEvent := events.OrderCreatedEvent{AmountCents: 100000}
	service := NewAccountService(repo)
	err := service.ProcessNewOrder(context.Background(), orderEvent)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccountEventStatus, actualAccountEventStatus)
}

func TestProcessNewOrderWithOrderAlreadyProcessed(t *testing.T) {
	repo := newAccountRepoStub(
		func(_ context.Context, _ int64) (bool, error) {
			return true, nil
		},
		nil,
		nil,
		nil,
		func(ctx context.Context, txfn func(context.Context) error) error {
			return txfn(ctx)
		},
	)

	orderEvent := events.OrderCreatedEvent{AmountCents: 100000}
	service := NewAccountService(repo)
	err := service.ProcessNewOrder(context.Background(), orderEvent)

	assert.ErrorIs(t, err, ErrAlreadyProcessed)
}

func TestProcessNewOrderWithNegativeAmount(t *testing.T) {
	repo := newAccountRepoStub(
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	orderEvent := events.OrderCreatedEvent{AmountCents: -100000}
	service := NewAccountService(repo)
	err := service.ProcessNewOrder(context.Background(), orderEvent)

	assert.ErrorIs(t, err, ErrInvalidData)
}

type accountRepoStub struct {
	orderEventExistsFn   func(context.Context, int64) (bool, error)
	getAccountByIDFn     func(context.Context, int64) (Account, error)
	updateAccountFn      func(context.Context, Account) error
	createAccountEventFn func(context.Context, events.AccountOrderPaymentEvent) error
	withinTxFn           func(context.Context, func(context.Context) error) error
}

func newAccountRepoStub(
	orderEventExistsFn func(context.Context, int64) (bool, error),
	getAccountByIDFn func(context.Context, int64) (Account, error),
	updateAccountFn func(context.Context, Account) error,
	createAccountEventFn func(context.Context, events.AccountOrderPaymentEvent) error,
	withinTxFn func(context.Context, func(context.Context) error) error,
) *accountRepoStub {
	return &accountRepoStub{
		orderEventExistsFn:   orderEventExistsFn,
		getAccountByIDFn:     getAccountByIDFn,
		updateAccountFn:      updateAccountFn,
		createAccountEventFn: createAccountEventFn,
		withinTxFn:           withinTxFn,
	}
}

func (r *accountRepoStub) AccountEventWithOrderEventIDExists(ctx context.Context, eventID int64) (bool, error) {
	return r.orderEventExistsFn(ctx, eventID)
}

func (r *accountRepoStub) GetAccountByUserID(ctx context.Context, userID int64) (Account, error) {
	return r.getAccountByIDFn(ctx, userID)
}

func (r *accountRepoStub) UpdateAccount(ctx context.Context, account Account) error {
	if r.updateAccountFn == nil {
		return nil
	}

	return r.updateAccountFn(ctx, account)
}

func (r *accountRepoStub) CreateAccountEvent(ctx context.Context, event events.AccountOrderPaymentEvent) error {
	if r.createAccountEventFn == nil {
		return nil
	}

	return r.createAccountEventFn(ctx, event)
}

func (r *accountRepoStub) WithinTransaction(ctx context.Context, txfn func(context.Context) error) error {
	if r.withinTxFn == nil {
		return txfn(ctx)
	}

	return r.withinTxFn(ctx, txfn)
}
