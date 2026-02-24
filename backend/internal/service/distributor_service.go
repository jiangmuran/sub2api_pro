package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrDistributorProfileNotFound   = infraerrors.NotFound("DISTRIBUTOR_PROFILE_NOT_FOUND", "distributor profile not found")
	ErrDistributorOfferNotFound     = infraerrors.NotFound("DISTRIBUTOR_OFFER_NOT_FOUND", "distributor offer not found")
	ErrDistributorOrderNotFound     = infraerrors.NotFound("DISTRIBUTOR_ORDER_NOT_FOUND", "distributor order not found")
	ErrDistributorDisabled          = infraerrors.Forbidden("DISTRIBUTOR_DISABLED", "distributor is disabled")
	ErrDistributorInsufficientCNY   = infraerrors.BadRequest("DISTRIBUTOR_INSUFFICIENT_BALANCE", "insufficient distributor cny balance")
	ErrDistributorOrderNotRevocable = infraerrors.Conflict("DISTRIBUTOR_ORDER_NOT_REVOCABLE", "order cannot be revoked")
)

type DistributorAdminSummary struct {
	UnsettledCNYCents         int64                  `json:"unsettled_cny_cents"`
	DeltaSinceLastSettleCNY   int64                  `json:"delta_since_last_settle_cny"`
	LastSettledAt             *time.Time             `json:"last_settled_at,omitempty"`
	LastSettledAmountCNYCents int64                  `json:"last_settled_amount_cny_cents"`
	ByUser                    []DistributorUserStats `json:"by_user"`
}

type DistributorService struct {
	db        *sql.DB
	userRepo  UserRepository
	groupRepo GroupRepository
}

func NewDistributorService(db *sql.DB, userRepo UserRepository, groupRepo GroupRepository) *DistributorService {
	return &DistributorService{db: db, userRepo: userRepo, groupRepo: groupRepo}
}

func (s *DistributorService) ensureProfile(ctx context.Context, userID int64) error {
	_, err := s.GetProfileByUserID(ctx, userID)
	return err
}

func (s *DistributorService) GetProfileByUserID(ctx context.Context, userID int64) (*DistributorProfile, error) {
	const q = `
SELECT p.id, p.user_id, p.enabled, p.balance_cny_cents, p.notes, p.created_at, p.updated_at,
       u.id, u.email, u.username, u.role, u.balance, u.concurrency, u.status, u.created_at, u.updated_at
FROM distributor_profiles p
JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
WHERE p.user_id = $1
`
	row := s.db.QueryRowContext(ctx, q, userID)
	var p DistributorProfile
	var u User
	if err := row.Scan(
		&p.ID, &p.UserID, &p.Enabled, &p.BalanceCNYCents, &p.Notes, &p.CreatedAt, &p.UpdatedAt,
		&u.ID, &u.Email, &u.Username, &u.Role, &u.Balance, &u.Concurrency, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDistributorProfileNotFound
		}
		return nil, err
	}
	p.User = &u
	return &p, nil
}

func (s *DistributorService) UpsertProfileByEmail(ctx context.Context, email string, enabled bool, notes string) (*DistributorProfile, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, infraerrors.BadRequest("DISTRIBUTOR_EMAIL_REQUIRED", "email is required")
	}
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	notes = strings.TrimSpace(notes)

	const upsert = `
INSERT INTO distributor_profiles (user_id, enabled, notes, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (user_id)
DO UPDATE SET enabled = EXCLUDED.enabled, notes = EXCLUDED.notes, updated_at = NOW()
`
	if _, err := s.db.ExecContext(ctx, upsert, user.ID, enabled, notes); err != nil {
		return nil, err
	}
	return s.GetProfileByUserID(ctx, user.ID)
}

func (s *DistributorService) ListProfiles(ctx context.Context, params pagination.PaginationParams, search string) ([]DistributorProfile, *pagination.PaginationResult, error) {
	search = strings.TrimSpace(search)
	where := ""
	args := []any{}
	if search != "" {
		where = " WHERE (u.email ILIKE $1 OR u.username ILIKE $1) "
		args = append(args, "%"+search+"%")
	}

	countQ := `SELECT COUNT(1) FROM distributor_profiles p JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL` + where
	var total int64
	if err := s.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, nil, err
	}

	listQ := `
SELECT p.id, p.user_id, p.enabled, p.balance_cny_cents, p.notes, p.created_at, p.updated_at,
       u.id, u.email, u.username, u.role, u.balance, u.concurrency, u.status, u.created_at, u.updated_at
FROM distributor_profiles p
JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL` + where + `
ORDER BY p.updated_at DESC
LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, params.Limit(), params.Offset())
	rows, err := s.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	out := make([]DistributorProfile, 0)
	for rows.Next() {
		var p DistributorProfile
		var u User
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Enabled, &p.BalanceCNYCents, &p.Notes, &p.CreatedAt, &p.UpdatedAt,
			&u.ID, &u.Email, &u.Username, &u.Role, &u.Balance, &u.Concurrency, &u.Status, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		p.User = &u
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return out, buildPaginationResult(total, params), nil
}

func (s *DistributorService) AdjustBalance(ctx context.Context, distributorUserID, operatorUserID int64, operation string, amountCNYCents int64, notes string) (*DistributorProfile, error) {
	if amountCNYCents <= 0 {
		return nil, infraerrors.BadRequest("DISTRIBUTOR_AMOUNT_INVALID", "amount must be > 0")
	}
	notes = strings.TrimSpace(notes)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var balance int64
	var enabled bool
	if err := tx.QueryRowContext(ctx, `SELECT balance_cny_cents, enabled FROM distributor_profiles WHERE user_id = $1 FOR UPDATE`, distributorUserID).Scan(&balance, &enabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDistributorProfileNotFound
		}
		return nil, err
	}

	delta := amountCNYCents
	ledgerType := DistributorLedgerTypeAdminTopup
	if operation == "refund" || operation == "subtract" {
		delta = -amountCNYCents
		ledgerType = DistributorLedgerTypeAdminRefund
	}
	newBalance := balance + delta
	if newBalance < 0 {
		return nil, ErrDistributorInsufficientCNY
	}
	if _, err := tx.ExecContext(ctx, `UPDATE distributor_profiles SET balance_cny_cents = $2, updated_at = NOW() WHERE user_id = $1`, distributorUserID, newBalance); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO distributor_wallet_ledger(distributor_user_id, type, amount_cny_cents, balance_after_cny_cents, operator_user_id, notes)
VALUES($1,$2,$3,$4,$5,$6)
`, distributorUserID, ledgerType, delta, newBalance, nullableInt64(operatorUserID), notes); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	_ = enabled
	return s.GetProfileByUserID(ctx, distributorUserID)
}

func (s *DistributorService) CreateOffer(ctx context.Context, distributorUserID int64, name string, targetGroupID int64, validityDays int, costCNYCents int64, notes string, enabled bool) (*DistributorOffer, error) {
	if err := s.ensureProfile(ctx, distributorUserID); err != nil {
		return nil, err
	}
	if validityDays <= 0 {
		return nil, infraerrors.BadRequest("DISTRIBUTOR_VALIDITY_INVALID", "validity_days must be > 0")
	}
	if costCNYCents <= 0 {
		return nil, infraerrors.BadRequest("DISTRIBUTOR_COST_INVALID", "cost must be > 0")
	}
	g, err := s.groupRepo.GetByID(ctx, targetGroupID)
	if err != nil {
		return nil, err
	}
	if g.SubscriptionType != SubscriptionTypeSubscription {
		return nil, infraerrors.BadRequest("DISTRIBUTOR_GROUP_INVALID", "target group must be subscription type")
	}
	name = strings.TrimSpace(name)
	notes = strings.TrimSpace(notes)

	const q = `
INSERT INTO distributor_offers(distributor_user_id,name,target_group_id,validity_days,cost_cny_cents,enabled,notes,created_at,updated_at)
VALUES($1,$2,$3,$4,$5,$6,$7,NOW(),NOW()) RETURNING id
`
	var id int64
	if err := s.db.QueryRowContext(ctx, q, distributorUserID, name, targetGroupID, validityDays, costCNYCents, enabled, notes).Scan(&id); err != nil {
		return nil, err
	}
	return s.GetOfferByID(ctx, id)
}

func (s *DistributorService) UpdateOffer(ctx context.Context, id int64, name string, targetGroupID int64, validityDays int, costCNYCents int64, notes string, enabled bool) (*DistributorOffer, error) {
	if validityDays <= 0 {
		return nil, infraerrors.BadRequest("DISTRIBUTOR_VALIDITY_INVALID", "validity_days must be > 0")
	}
	if costCNYCents <= 0 {
		return nil, infraerrors.BadRequest("DISTRIBUTOR_COST_INVALID", "cost must be > 0")
	}
	g, err := s.groupRepo.GetByID(ctx, targetGroupID)
	if err != nil {
		return nil, err
	}
	if g.SubscriptionType != SubscriptionTypeSubscription {
		return nil, infraerrors.BadRequest("DISTRIBUTOR_GROUP_INVALID", "target group must be subscription type")
	}
	res, err := s.db.ExecContext(ctx, `
UPDATE distributor_offers
SET name=$2, target_group_id=$3, validity_days=$4, cost_cny_cents=$5, enabled=$6, notes=$7, updated_at=NOW()
WHERE id=$1
`, id, strings.TrimSpace(name), targetGroupID, validityDays, costCNYCents, enabled, strings.TrimSpace(notes))
	if err != nil {
		return nil, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil, ErrDistributorOfferNotFound
	}
	return s.GetOfferByID(ctx, id)
}

func (s *DistributorService) GetOfferByID(ctx context.Context, id int64) (*DistributorOffer, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, distributor_user_id, name, target_group_id, validity_days, cost_cny_cents, enabled, notes, created_at, updated_at
FROM distributor_offers WHERE id = $1
`, id)
	var o DistributorOffer
	if err := row.Scan(&o.ID, &o.DistributorUserID, &o.Name, &o.TargetGroupID, &o.ValidityDays, &o.CostCNYCents, &o.Enabled, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDistributorOfferNotFound
		}
		return nil, err
	}
	return &o, nil
}

func (s *DistributorService) ListOffers(ctx context.Context, distributorUserID int64, onlyEnabled bool) ([]DistributorOffer, error) {
	args := []any{distributorUserID}
	where := " WHERE distributor_user_id = $1"
	if onlyEnabled {
		where += " AND enabled = TRUE"
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, distributor_user_id, name, target_group_id, validity_days, cost_cny_cents, enabled, notes, created_at, updated_at
FROM distributor_offers`+where+` ORDER BY updated_at DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]DistributorOffer, 0)
	for rows.Next() {
		var o DistributorOffer
		if err := rows.Scan(&o.ID, &o.DistributorUserID, &o.Name, &o.TargetGroupID, &o.ValidityDays, &o.CostCNYCents, &o.Enabled, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *DistributorService) DeleteOffer(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM distributor_offers WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrDistributorOfferNotFound
	}
	return nil
}

func (s *DistributorService) PurchaseOrder(ctx context.Context, distributorUserID, offerID, sellPriceCNYCents int64, memo string) (*DistributorOrder, error) {
	memo = strings.TrimSpace(memo)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var profileEnabled bool
	var balance int64
	if err := tx.QueryRowContext(ctx, `SELECT enabled, balance_cny_cents FROM distributor_profiles WHERE user_id=$1 FOR UPDATE`, distributorUserID).Scan(&profileEnabled, &balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDistributorProfileNotFound
		}
		return nil, err
	}
	if !profileEnabled {
		return nil, ErrDistributorDisabled
	}

	var offer DistributorOffer
	if err := tx.QueryRowContext(ctx, `
SELECT id, distributor_user_id, name, target_group_id, validity_days, cost_cny_cents, enabled, notes, created_at, updated_at
FROM distributor_offers WHERE id=$1 AND distributor_user_id=$2 FOR UPDATE
`, offerID, distributorUserID).Scan(&offer.ID, &offer.DistributorUserID, &offer.Name, &offer.TargetGroupID, &offer.ValidityDays, &offer.CostCNYCents, &offer.Enabled, &offer.Notes, &offer.CreatedAt, &offer.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDistributorOfferNotFound
		}
		return nil, err
	}
	if !offer.Enabled {
		return nil, ErrDistributorDisabled
	}
	if sellPriceCNYCents <= 0 {
		sellPriceCNYCents = offer.CostCNYCents
	}
	if balance < offer.CostCNYCents {
		return nil, ErrDistributorInsufficientCNY
	}

	code, err := GenerateRedeemCode()
	if err != nil {
		return nil, err
	}
	redeemNotes := EncodeRedeemNotes(memo, "distributor")
	var redeemCodeID int64
	if err := tx.QueryRowContext(ctx, `
INSERT INTO redeem_codes(code,type,value,status,notes,group_id,validity_days,created_at,updated_at)
VALUES($1,$2,$3,$4,$5,$6,$7,NOW(),NOW()) RETURNING id
`, code, RedeemTypeSubscription, 0, StatusUnused, redeemNotes, offer.TargetGroupID, offer.ValidityDays).Scan(&redeemCodeID); err != nil {
		return nil, err
	}

	newBalance := balance - offer.CostCNYCents
	if _, err := tx.ExecContext(ctx, `UPDATE distributor_profiles SET balance_cny_cents=$2, updated_at=NOW() WHERE user_id=$1`, distributorUserID, newBalance); err != nil {
		return nil, err
	}

	var orderID int64
	if err := tx.QueryRowContext(ctx, `
INSERT INTO distributor_orders(distributor_user_id,offer_id,redeem_code_id,cost_cny_cents,sell_price_cny_cents,status,memo,issued_at,created_at,updated_at)
VALUES($1,$2,$3,$4,$5,$6,$7,NOW(),NOW(),NOW()) RETURNING id
`, distributorUserID, offer.ID, redeemCodeID, offer.CostCNYCents, sellPriceCNYCents, DistributorOrderStatusIssued, memo).Scan(&orderID); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO distributor_wallet_ledger(distributor_user_id,type,amount_cny_cents,balance_after_cny_cents,order_id,notes)
VALUES($1,$2,$3,$4,$5,$6)
`, distributorUserID, DistributorLedgerTypeRedeemBuy, -offer.CostCNYCents, newBalance, orderID, memo); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetOrderByID(ctx, orderID, true), nil
}

func (s *DistributorService) GetOrderByID(ctx context.Context, orderID int64, includeEmail bool) *DistributorOrder {
	order, _ := s.getOrderByID(ctx, orderID, includeEmail)
	return order
}

func (s *DistributorService) getOrderByID(ctx context.Context, orderID int64, includeEmail bool) (*DistributorOrder, error) {
	q := s.orderQuery(includeEmail) + ` WHERE o.id = $1`
	rows, err := s.db.QueryContext(ctx, q, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders, err := s.scanOrders(rows, includeEmail)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, ErrDistributorOrderNotFound
	}
	return &orders[0], nil
}

func (s *DistributorService) RevokeOrder(ctx context.Context, actorUserID int64, isAdmin bool, orderID int64, notes string) (*DistributorOrder, error) {
	notes = strings.TrimSpace(notes)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var distributorUserID int64
	var redeemCodeID int64
	var status string
	var cost int64
	if err := tx.QueryRowContext(ctx, `SELECT distributor_user_id, redeem_code_id, status, cost_cny_cents FROM distributor_orders WHERE id=$1 FOR UPDATE`, orderID).Scan(&distributorUserID, &redeemCodeID, &status, &cost); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDistributorOrderNotFound
		}
		return nil, err
	}
	if !isAdmin && actorUserID != distributorUserID {
		return nil, infraerrors.Forbidden("DISTRIBUTOR_FORBIDDEN", "forbidden")
	}
	if status != DistributorOrderStatusIssued {
		return nil, ErrDistributorOrderNotRevocable
	}

	var redeemStatus string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM redeem_codes WHERE id=$1 FOR UPDATE`, redeemCodeID).Scan(&redeemStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRedeemCodeNotFound
		}
		return nil, err
	}
	if redeemStatus != StatusUnused {
		return nil, ErrDistributorOrderNotRevocable
	}

	if _, err := tx.ExecContext(ctx, `UPDATE redeem_codes SET status=$2, updated_at=NOW() WHERE id=$1`, redeemCodeID, StatusRevoked); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE distributor_orders SET status=$2, revoked_at=NOW(), revoked_by_admin=$3, updated_at=NOW(), memo = CASE WHEN $4 <> '' THEN $4 ELSE memo END WHERE id=$1`, orderID, DistributorOrderStatusRevoked, isAdmin, notes); err != nil {
		return nil, err
	}

	var balance int64
	if err := tx.QueryRowContext(ctx, `SELECT balance_cny_cents FROM distributor_profiles WHERE user_id=$1 FOR UPDATE`, distributorUserID).Scan(&balance); err != nil {
		return nil, err
	}
	newBalance := balance + cost
	if _, err := tx.ExecContext(ctx, `UPDATE distributor_profiles SET balance_cny_cents=$2, updated_at=NOW() WHERE user_id=$1`, distributorUserID, newBalance); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO distributor_wallet_ledger(distributor_user_id,type,amount_cny_cents,balance_after_cny_cents,operator_user_id,order_id,notes)
VALUES($1,$2,$3,$4,$5,$6,$7)
`, distributorUserID, DistributorLedgerTypeRevokeRefund, cost, newBalance, nullableInt64(actorUserID), orderID, notes); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.getOrderByID(ctx, orderID, isAdmin)
}

func (s *DistributorService) MarkOrderRedeemedByRedeemCodeID(ctx context.Context, redeemCodeID, redeemedByUserID int64) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE distributor_orders
SET status=$2, redeemed_at=NOW(), updated_at=NOW()
WHERE redeem_code_id=$1 AND status=$3
`, redeemCodeID, DistributorOrderStatusRedeemed, DistributorOrderStatusIssued)
	return err
}

func (s *DistributorService) ListOrdersForDistributor(ctx context.Context, distributorUserID int64, params pagination.PaginationParams, status, search string) ([]DistributorOrder, *pagination.PaginationResult, error) {
	return s.listOrders(ctx, params, distributorUserID, status, search, false)
}

func (s *DistributorService) ListOrdersAdmin(ctx context.Context, params pagination.PaginationParams, distributorUserID int64, status, search string) ([]DistributorOrder, *pagination.PaginationResult, error) {
	return s.listOrders(ctx, params, distributorUserID, status, search, true)
}

func (s *DistributorService) listOrders(ctx context.Context, params pagination.PaginationParams, distributorUserID int64, status, search string, includeEmail bool) ([]DistributorOrder, *pagination.PaginationResult, error) {
	status = strings.TrimSpace(status)
	search = strings.TrimSpace(search)
	args := make([]any, 0)
	where := make([]string, 0)
	if distributorUserID > 0 {
		args = append(args, distributorUserID)
		where = append(where, fmt.Sprintf("o.distributor_user_id = $%d", len(args)))
	}
	if status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("o.status = $%d", len(args)))
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		cond := fmt.Sprintf("(r.code ILIKE $%d OR o.memo ILIKE $%d)", len(args), len(args))
		if includeEmail {
			cond = fmt.Sprintf("(r.code ILIKE $%d OR o.memo ILIKE $%d OR ru.email ILIKE $%d)", len(args), len(args), len(args))
		}
		where = append(where, cond)
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	countQ := `SELECT COUNT(1) FROM distributor_orders o JOIN redeem_codes r ON r.id=o.redeem_code_id LEFT JOIN users ru ON ru.id=r.used_by` + whereSQL
	var total int64
	if err := s.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, nil, err
	}

	args = append(args, params.Limit(), params.Offset())
	listQ := s.orderQuery(includeEmail) + whereSQL + fmt.Sprintf(" ORDER BY o.issued_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := s.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	items, err := s.scanOrders(rows, includeEmail)
	if err != nil {
		return nil, nil, err
	}
	return items, buildPaginationResult(total, params), nil
}

func (s *DistributorService) orderQuery(includeEmail bool) string {
	selectUser := ""
	if includeEmail {
		selectUser = ", ru.email"
	}
	return `
SELECT o.id, o.distributor_user_id, o.offer_id, o.redeem_code_id, o.cost_cny_cents, o.sell_price_cny_cents,
       o.status, o.memo, o.issued_at, o.redeemed_at, o.revoked_at, o.revoked_by_admin, o.created_at, o.updated_at,
       r.code, r.status, r.used_at, r.used_by` + selectUser + `
FROM distributor_orders o
JOIN redeem_codes r ON r.id = o.redeem_code_id
LEFT JOIN users ru ON ru.id = r.used_by
`
}

func (s *DistributorService) scanOrders(rows *sql.Rows, includeEmail bool) ([]DistributorOrder, error) {
	out := make([]DistributorOrder, 0)
	for rows.Next() {
		var o DistributorOrder
		var code string
		var redeemStatus string
		var usedAt sql.NullTime
		var usedBy sql.NullInt64
		var redeemedAt sql.NullTime
		var revokedAt sql.NullTime
		var usedByEmail sql.NullString
		base := []any{&o.ID, &o.DistributorUserID, &o.OfferID, &o.RedeemCodeID, &o.CostCNYCents, &o.SellPriceCNYCents, &o.Status, &o.Memo, &o.IssuedAt, &redeemedAt, &revokedAt, &o.RevokedByAdmin, &o.CreatedAt, &o.UpdatedAt, &code, &redeemStatus, &usedAt, &usedBy}
		if includeEmail {
			base = append(base, &usedByEmail)
		}
		if err := rows.Scan(base...); err != nil {
			return nil, err
		}
		if redeemedAt.Valid {
			t := redeemedAt.Time
			o.RedeemedAt = &t
		}
		if revokedAt.Valid {
			t := revokedAt.Time
			o.RevokedAt = &t
		}
		r := &RedeemCode{ID: o.RedeemCodeID, Code: code, Status: redeemStatus}
		if usedAt.Valid {
			t := usedAt.Time
			r.UsedAt = &t
		}
		if usedBy.Valid {
			v := usedBy.Int64
			r.UsedBy = &v
		}
		if includeEmail && usedByEmail.Valid {
			r.User = &User{Email: usedByEmail.String}
		}
		o.RedeemCode = r
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *DistributorService) GetAdminSummary(ctx context.Context) (*DistributorAdminSummary, error) {
	summary := &DistributorAdminSummary{}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(balance_cny_cents),0) FROM distributor_profiles`).Scan(&summary.UnsettledCNYCents); err != nil {
		return nil, err
	}
	var lastAt sql.NullTime
	if err := s.db.QueryRowContext(ctx, `SELECT amount_cny_cents, created_at FROM distributor_settlement_checkpoints ORDER BY id DESC LIMIT 1`).Scan(&summary.LastSettledAmountCNYCents, &lastAt); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if lastAt.Valid {
		t := lastAt.Time
		summary.LastSettledAt = &t
	}
	summary.DeltaSinceLastSettleCNY = summary.UnsettledCNYCents - summary.LastSettledAmountCNYCents

	rows, err := s.db.QueryContext(ctx, `
SELECT o.distributor_user_id,
       COUNT(1) AS total_orders,
       COUNT(1) FILTER (WHERE o.status = 'issued') AS issued_count,
       COUNT(1) FILTER (WHERE o.status = 'redeemed') AS redeemed_count,
       COUNT(1) FILTER (WHERE o.status = 'revoked') AS revoked_count,
       COALESCE(SUM(o.sell_price_cny_cents),0) AS sell_amount,
       COALESCE(SUM(o.cost_cny_cents),0) AS cost_amount,
       COALESCE(SUM(o.cost_cny_cents) FILTER (WHERE o.status='revoked'),0) AS refund_amount,
       COALESCE(SUM(o.sell_price_cny_cents) FILTER (WHERE o.status<>'revoked'),0) AS net_sell,
       COALESCE(SUM(o.sell_price_cny_cents) FILTER (WHERE o.status<>'revoked'),0) -
       (COALESCE(SUM(o.cost_cny_cents),0) - COALESCE(SUM(o.cost_cny_cents) FILTER (WHERE o.status='revoked'),0)) AS gross_profit
FROM distributor_orders o
GROUP BY o.distributor_user_id
ORDER BY gross_profit DESC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item DistributorUserStats
		if err := rows.Scan(&item.DistributorUserID, &item.OrdersTotal, &item.IssuedCount, &item.RedeemedCount, &item.RevokedCount, &item.SellAmountCNY, &item.CostAmountCNY, &item.RefundAmountCNY, &item.NetSellAmountCNY, &item.GrossProfitCNY); err != nil {
			return nil, err
		}
		summary.ByUser = append(summary.ByUser, item)
	}
	return summary, rows.Err()
}

func (s *DistributorService) MarkSettled(ctx context.Context, operatorUserID int64, notes string) (*DistributorAdminSummary, error) {
	summary, err := s.GetAdminSummary(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO distributor_settlement_checkpoints(amount_cny_cents, created_by, notes)
VALUES($1,$2,$3)
`, summary.UnsettledCNYCents, nullableInt64(operatorUserID), strings.TrimSpace(notes)); err != nil {
		return nil, err
	}
	return s.GetAdminSummary(ctx)
}

func nullableInt64(v int64) any {
	if v <= 0 {
		return nil
	}
	return v
}

func buildPaginationResult(total int64, params pagination.PaginationParams) *pagination.PaginationResult {
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.Limit()
	pages := 0
	if total > 0 {
		pages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return &pagination.PaginationResult{Total: total, Page: page, PageSize: pageSize, Pages: pages}
}
