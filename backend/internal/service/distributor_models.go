package service

import "time"

const (
	DistributorOrderStatusIssued   = "issued"
	DistributorOrderStatusRedeemed = "redeemed"
	DistributorOrderStatusRevoked  = "revoked"

	DistributorLedgerTypeAdminTopup   = "admin_topup"
	DistributorLedgerTypeAdminRefund  = "admin_refund"
	DistributorLedgerTypeRedeemBuy    = "redeem_purchase"
	DistributorLedgerTypeRevokeRefund = "revoke_refund"
)

type DistributorProfile struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Enabled         bool      `json:"enabled"`
	BalanceCNYCents int64     `json:"balance_cny_cents"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	User *User `json:"user,omitempty"`
}

type DistributorOffer struct {
	ID                int64     `json:"id"`
	DistributorUserID int64     `json:"distributor_user_id"`
	Name              string    `json:"name"`
	TargetGroupID     int64     `json:"target_group_id"`
	ValidityDays      int       `json:"validity_days"`
	CostCNYCents      int64     `json:"cost_cny_cents"`
	Enabled           bool      `json:"enabled"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	Group *Group `json:"group,omitempty"`
}

type DistributorOrder struct {
	ID                int64      `json:"id"`
	DistributorUserID int64      `json:"distributor_user_id"`
	OfferID           int64      `json:"offer_id"`
	RedeemCodeID      int64      `json:"redeem_code_id"`
	CostCNYCents      int64      `json:"cost_cny_cents"`
	SellPriceCNYCents int64      `json:"sell_price_cny_cents"`
	Status            string     `json:"status"`
	Memo              string     `json:"memo"`
	IssuedAt          time.Time  `json:"issued_at"`
	RedeemedAt        *time.Time `json:"redeemed_at,omitempty"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	RevokedByAdmin    bool       `json:"revoked_by_admin"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	Offer      *DistributorOffer `json:"offer,omitempty"`
	RedeemCode *RedeemCode       `json:"redeem_code,omitempty"`
}

type DistributorWalletLedger struct {
	ID                   int64     `json:"id"`
	DistributorUserID    int64     `json:"distributor_user_id"`
	Type                 string    `json:"type"`
	AmountCNYCents       int64     `json:"amount_cny_cents"`
	BalanceAfterCNYCents int64     `json:"balance_after_cny_cents"`
	OperatorUserID       *int64    `json:"operator_user_id,omitempty"`
	OrderID              *int64    `json:"order_id,omitempty"`
	Notes                string    `json:"notes"`
	CreatedAt            time.Time `json:"created_at"`
}

type DistributorUserStats struct {
	DistributorUserID int64 `json:"distributor_user_id"`
	OrdersTotal       int64 `json:"orders_total"`
	IssuedCount       int64 `json:"issued_count"`
	RedeemedCount     int64 `json:"redeemed_count"`
	RevokedCount      int64 `json:"revoked_count"`
	SellAmountCNY     int64 `json:"sell_amount_cny"`
	CostAmountCNY     int64 `json:"cost_amount_cny"`
	RefundAmountCNY   int64 `json:"refund_amount_cny"`
	NetSellAmountCNY  int64 `json:"net_sell_amount_cny"`
	GrossProfitCNY    int64 `json:"gross_profit_cny"`
}
