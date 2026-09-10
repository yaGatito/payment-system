package postgres

import (
	"context"
	"payment-system/internal/domain"
	"payment-system/sql/sqlcgen"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type WalletRepository interface {
	AddCard(ctx context.Context, card domain.Card) error
	GetCardByID(ctx context.Context, cardID uuid.UUID) (domain.Card, error)
	GetCards(ctx context.Context, ownerID uuid.UUID) []domain.Card
	RemoveCard(ctx context.Context, cardID uuid.UUID) error

	AddPayment(ctx context.Context, payment domain.Payment) error
	GetPayments(ctx context.Context, ownerID uuid.UUID, limit, offset int32) ([]domain.Payment, error)
}

type WalletRepoPostgreSQL struct {
	queries sqlcgen.Querier
}

var _ WalletRepository = (*WalletRepoPostgreSQL)(nil)

func NewWalletRepoPostgreSQL(querier sqlcgen.Querier) WalletRepository {
	w := WalletRepoPostgreSQL{}
	w.queries = querier
	return &w
}

func (r *WalletRepoPostgreSQL) AddCard(ctx context.Context, card domain.Card) error {
	arg := sqlcgen.AddCardParams{
		OwnerID: card.CustomerID,
		Type:    card.Type,
		Last4:   card.Last4,
	}
	_, err := r.queries.AddCard(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}

func (r *WalletRepoPostgreSQL) GetCardByID(ctx context.Context, cardID uuid.UUID) (domain.Card, error) {
	row, err := r.queries.GetCardByID(ctx, cardID)
	if err != nil {
		return domain.Card{}, err
	}
	return domain.Card{
		CardID:     row.ID,
		CustomerID: row.OwnerID,
		Token:      row.Token,
		Type:       row.Type,
		Last4:      row.Last4,
		LinkedAt:   row.Linkedat.Time,
	}, nil
}

func (r *WalletRepoPostgreSQL) GetCards(ctx context.Context, ownerID uuid.UUID) []domain.Card {
	rows, err := r.queries.GetCards(ctx, ownerID)
	if err != nil {
		return nil
	}
	cards := make([]domain.Card, len(rows))
	for i, row := range rows {
		cards[i] = domain.Card{
			CardID:     row.ID,
			CustomerID: ownerID,
			Type:       row.Type,
			Last4:      row.Last4,
			LinkedAt:   row.Linkedat.Time,
		}
	}
	return cards
}

func (r *WalletRepoPostgreSQL) RemoveCard(ctx context.Context, cardID uuid.UUID) error {
	err := r.queries.RemoveCard(ctx, cardID)
	if err != nil {
		return err
	}
	return nil
}

func (r *WalletRepoPostgreSQL) AddPayment(ctx context.Context, payment domain.Payment) error {
	var amount pgtype.Numeric
	err := amount.Scan(strconv.FormatFloat(payment.Amount, 'f', -1, 64))
	if err != nil {
		return err
	}

	arg := sqlcgen.AddPaymentParams{
		OwnerID:  payment.CustomerID,
		CardID:   payment.CardID,
		Amount:   amount,
		Currency: payment.Currency,
		Status:   payment.Status,
	}
	err = r.queries.AddPayment(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}

func (r *WalletRepoPostgreSQL) GetPayments(ctx context.Context, ownerID uuid.UUID, limit, offset int32) ([]domain.Payment, error) {
	args := sqlcgen.GetPaymentsParams{
		OwnerID: ownerID,
		Limit:   limit,
		Offset:  offset,
	}
	rows, err := r.queries.GetPayments(ctx, args)
	if err != nil {
		return nil, err
	}
	payments := make([]domain.Payment, len(rows))
	for i, row := range rows {
		amount, err := row.Amount.Float64Value()
		if err != nil {
			return nil, err
		}
		payments[i] = domain.Payment{
			OrderID:    row.ID,
			CustomerID: ownerID,
			CardID:     row.CardID,
			Amount:     amount.Float64,
			Currency:   row.Currency,
			Status:     row.Status,
			CreatedAt:  row.CreatedAt.Time,
		}
	}
	return payments, nil
}
