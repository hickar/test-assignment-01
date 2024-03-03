package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hickar/crtex_test_assignment/events"

	"github.com/hickar/crtex_test_assignment/account/internal/domain"
)

type AccountRepository struct {
	db *pgxpool.Pool
}

func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: pool}
}

func (r *AccountRepository) AccountEventWithOrderEventIDExists(ctx context.Context, orderEventID int64) (bool, error) {
	tx := getTxFromContextOrDB(ctx, r.db)

	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM account_events WHERE order_event_id = $1);`

	err := tx.QueryRow(ctx, query, orderEventID).Scan(&exists)
	return exists, err
}

func (r *AccountRepository) GetAccountByUserID(ctx context.Context, userID int64) (domain.Account, error) {
	tx := getTxFromContextOrDB(ctx, r.db)

	query := `SELECT id, user_id, amount_cents FROM accounts WHERE user_id = $1;`

	var account domain.Account
	if err := tx.QueryRow(ctx, query, userID).Scan(
		&account.ID,
		&account.UserID,
		&account.AmountCents,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return account, domain.ErrNotFound
		}

		return account, err
	}

	return account, nil
}

func (r *AccountRepository) UpdateAccount(ctx context.Context, account domain.Account) error {
	tx := getTxFromContextOrDB(ctx, r.db)

	query := `UPDATE accounts SET amount_cents = $2 WHERE id = $1;`

	_, err := tx.Exec(ctx, query, account.ID, account.AmountCents)
	return err
}

func (r *AccountRepository) CreateAccountEvent(ctx context.Context, event events.AccountOrderPaymentEvent) error {
	tx := getTxFromContextOrDB(ctx, r.db)

	query := `INSERT INTO account_events (order_event_id, account_id, order_id, status) VALUES ($1, $2, $3, $4);`

	var accountID *int64
	if event.AccountID != 0 {
		accountID = &event.AccountID
	}

	_, err := tx.Exec(ctx, query, event.OrderEventID, accountID, event.OrderID, event.Status)
	return err
}

type txContextKey string

const transactionCtxKey txContextKey = "ctxtransaction"

func (r *AccountRepository) WithinTransaction(ctx context.Context, txfn func(context.Context) error) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.RepeatableRead,
	})
	if err != nil {
		return err
	}

	tctx := context.WithValue(ctx, transactionCtxKey, tx)
	if err = txfn(tctx); err != nil {
		return tx.Rollback(ctx)
	}

	return tx.Commit(ctx)
}

func getTxFromContextOrDB(ctx context.Context, db *pgxpool.Pool) Querier {
	txval := ctx.Value(transactionCtxKey)
	tx, ok := txval.(pgx.Tx)
	if !ok {
		return db
	}

	return tx
}

type Querier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}
