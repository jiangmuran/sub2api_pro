//go:build unit

package service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

type promoStatsTestRepo struct {
	code   *PromoCode
	usages []PromoCodeUsage
}

func (r *promoStatsTestRepo) Create(_ context.Context, _ *PromoCode) error { return nil }
func (r *promoStatsTestRepo) GetByID(_ context.Context, id int64) (*PromoCode, error) {
	if r.code == nil || r.code.ID != id {
		return nil, ErrPromoCodeNotFound
	}
	copyCode := *r.code
	return &copyCode, nil
}
func (r *promoStatsTestRepo) GetByCode(_ context.Context, _ string) (*PromoCode, error) {
	return nil, ErrPromoCodeNotFound
}
func (r *promoStatsTestRepo) GetByCodeForUpdate(_ context.Context, _ string) (*PromoCode, error) {
	return nil, ErrPromoCodeNotFound
}
func (r *promoStatsTestRepo) Update(_ context.Context, _ *PromoCode) error { return nil }
func (r *promoStatsTestRepo) Delete(_ context.Context, _ int64) error      { return nil }
func (r *promoStatsTestRepo) List(_ context.Context, _ pagination.PaginationParams) ([]PromoCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *promoStatsTestRepo) ListWithFilters(_ context.Context, _ pagination.PaginationParams, _, _ string) ([]PromoCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (r *promoStatsTestRepo) CreateUsage(_ context.Context, _ *PromoCodeUsage) error { return nil }
func (r *promoStatsTestRepo) GetUsageByPromoCodeAndUser(_ context.Context, _, _ int64) (*PromoCodeUsage, error) {
	return nil, nil
}
func (r *promoStatsTestRepo) GetLatestUsageByUser(_ context.Context, _ int64) (*PromoCodeUsage, error) {
	if len(r.usages) == 0 {
		return nil, nil
	}
	copyUsage := r.usages[0]
	return &copyUsage, nil
}

func (r *promoStatsTestRepo) ListUsagesByPromoCode(_ context.Context, promoCodeID int64, params pagination.PaginationParams) ([]PromoCodeUsage, *pagination.PaginationResult, error) {
	if r.code == nil || r.code.ID != promoCodeID {
		return nil, &pagination.PaginationResult{Page: params.Page, PageSize: params.PageSize, Total: 0, Pages: 0}, nil
	}

	if len(r.usages) == 0 {
		return nil, &pagination.PaginationResult{Page: params.Page, PageSize: params.PageSize, Total: 0, Pages: 0}, nil
	}

	start := (params.Page - 1) * params.PageSize
	if start >= len(r.usages) {
		return nil, &pagination.PaginationResult{Page: params.Page, Pages: 1, Total: int64(len(r.usages))}, nil
	}
	end := start + params.PageSize
	if end > len(r.usages) {
		end = len(r.usages)
	}

	pageItems := make([]PromoCodeUsage, 0, end-start)
	for _, u := range r.usages[start:end] {
		pageItems = append(pageItems, u)
	}

	pages := (len(r.usages) + params.PageSize - 1) / params.PageSize
	pageResult := &pagination.PaginationResult{
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    int64(len(r.usages)),
		Pages:    pages,
	}
	return pageItems, pageResult, nil
}

func (r *promoStatsTestRepo) IncrementUsedCount(_ context.Context, _ int64) error { return nil }

func newPromoStatsTestEntClient(t *testing.T) *dbent.Client {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_fk=1", t.Name())

	db, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestPromoServiceGetUsageStatsAggregatesCorrectly(t *testing.T) {
	t.Parallel()

	entClient := newPromoStatsTestEntClient(t)

	codeID := int64(1)
	promoCode := &PromoCode{
		ID:          codeID,
		Code:        "PROMO-TEST",
		BonusAmount: 10,
		MaxUses:     0,
		UsedCount:   3,
	}

	now := time.Now()
	today := now
	sevenDaysAgo := now.AddDate(0, 0, -3)
	thirtyDaysAgo := now.AddDate(0, 0, -10)

	repo := &promoStatsTestRepo{
		code: promoCode,
		usages: []PromoCodeUsage{
			{PromoCodeID: codeID, UserID: 1, BonusAmount: 10, UsedAt: today},
			{PromoCodeID: codeID, UserID: 2, BonusAmount: 10, UsedAt: sevenDaysAgo},
			{PromoCodeID: codeID, UserID: 1, BonusAmount: 10, UsedAt: thirtyDaysAgo},
		},
	}

	svc := NewPromoService(repo, nil, nil, entClient, nil)
	stats, err := svc.GetUsageStats(context.Background(), codeID)
	require.NoError(t, err)
	require.NotNil(t, stats)

	require.Equal(t, int64(3), stats.TotalUses)
	require.InDelta(t, 30.0, stats.TotalBonusAmount, 1e-9)
	require.Equal(t, int64(2), stats.UniqueUsers)
	require.GreaterOrEqual(t, stats.UsesToday, int64(1))
	require.GreaterOrEqual(t, stats.UsesLast7Days, int64(2))
	require.GreaterOrEqual(t, stats.UsesLast30Days, int64(3))
}
