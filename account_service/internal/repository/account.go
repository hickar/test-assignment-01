package repository

import (
	"context"
	"database/sql"

	"github.com/hickar/crtex_test_assignment/account_service/internal/domain"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) AccountEventWithOrderEventIDExists(ctx context.Context, orderEventID int64) (bool, error) {
	tx := getTxFromContextOrDB(ctx, r.db)

	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM account_events WHERE order_event_id = $1);`

	err := tx.QueryRowContext(ctx, query, orderEventID).Scan(&exists)
	return exists, err
}

func (r *AccountRepository) GetAccountByUserID(ctx context.Context, userID int64) (domain.Account, error) {
	tx := getTxFromContextOrDB(ctx, r.db)

	query := `SELECT id, user_id, amount_cents FROM accounts WHERE user_id = $1;`

	var account domain.Account
	if err := tx.QueryRowContext(ctx, query, userID).Scan(
		&account.ID,
		&account.UserID,
		&account.AmountCents,
	); err != nil {
		return account, err
	}

	return account, nil
}

func (r *AccountRepository) UpdateAccount(ctx context.Context, account domain.Account) error {
	tx := getTxFromContextOrDB(ctx, r.db)

	query := `UPDATE accounts SET amount_cents = $2 WHERE id = $1;`

	_, err := tx.ExecContext(ctx, query, account.ID, account.AmountCents)
	return err
}

func (r *AccountRepository) CreateAccountEvent(ctx context.Context, event domain.AccountEvent) error {
	tx := getTxFromContextOrDB(ctx, r.db)

	query := `INSERT INTO account_events (order_event_id, account_id, order_id, status) VALUES ($1, $2, $3, $4);`

	_, err := tx.ExecContext(ctx, query, event.OrderEventID, event.AccountID, event.OrderEventID, event.Status)
	return err
}

const transactionCtxKey = "ctxtransaction"

func (r *AccountRepository) WithinTransaction(ctx context.Context, txfn func(context.Context) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	//nolint:staticcheck
	tctx := context.WithValue(ctx, transactionCtxKey, tx)
	if err = txfn(tctx); err != nil {
		return tx.Rollback()
	}

	return tx.Commit()
}

func getTxFromContextOrDB(ctx context.Context, db *sql.DB) Querier {
	txval := ctx.Value(transactionCtxKey)
	tx, ok := txval.(*sql.Tx)
	if !ok {
		return db
	}

	return tx
}

type Querier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}
