//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type invitationStatsTestRedeemRepo struct {
	invitation *RedeemCode
	redeems    []RedeemCode
}

func (r *invitationStatsTestRedeemRepo) Create(_ context.Context, _ *RedeemCode) error { return nil }
func (r *invitationStatsTestRedeemRepo) CreateBatch(_ context.Context, _ []RedeemCode) error {
	return nil
}

func (r *invitationStatsTestRedeemRepo) GetByID(_ context.Context, id int64) (*RedeemCode, error) {
	if r.invitation == nil || r.invitation.ID != id {
		return nil, ErrRedeemCodeNotFound
	}
	copyCode := *r.invitation
	return &copyCode, nil
}

func (r *invitationStatsTestRedeemRepo) GetByCode(_ context.Context, _ string) (*RedeemCode, error) {
	return nil, ErrRedeemCodeNotFound
}

func (r *invitationStatsTestRedeemRepo) Update(_ context.Context, _ *RedeemCode) error { return nil }
func (r *invitationStatsTestRedeemRepo) Delete(_ context.Context, _ int64) error       { return nil }
func (r *invitationStatsTestRedeemRepo) Use(_ context.Context, _, _ int64) error       { return nil }

func (r *invitationStatsTestRedeemRepo) List(_ context.Context, _ pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *invitationStatsTestRedeemRepo) ListWithFilters(_ context.Context, _ pagination.PaginationParams, _, _, _, _ string) ([]RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *invitationStatsTestRedeemRepo) ListByUser(_ context.Context, _ int64, _ int) ([]RedeemCode, error) {
	return nil, nil
}

func (r *invitationStatsTestRedeemRepo) ListByUserPaginated(_ context.Context, userID int64, params pagination.PaginationParams, codeType string) ([]RedeemCode, *pagination.PaginationResult, error) {
	if codeType != RedeemTypeSubscription {
		return nil, &pagination.PaginationResult{Page: params.Page, PageSize: params.PageSize, Total: 0, Pages: 0}, nil
	}

	filtered := make([]RedeemCode, 0, len(r.redeems))
	for _, c := range r.redeems {
		if c.UsedBy != nil && *c.UsedBy == userID && c.Type == RedeemTypeSubscription {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 {
		return nil, &pagination.PaginationResult{Page: params.Page, PageSize: params.PageSize, Total: 0, Pages: 0}, nil
	}

	start := (params.Page - 1) * params.PageSize
	if start >= len(filtered) {
		return nil, &pagination.PaginationResult{Page: params.Page, PageSize: params.PageSize, Total: int64(len(filtered)), Pages: 1}, nil
	}
	end := start + params.PageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	pageItems := make([]RedeemCode, 0, end-start)
	for _, c := range filtered[start:end] {
		pageItems = append(pageItems, c)
	}

	pages := (len(filtered) + params.PageSize - 1) / params.PageSize
	pageResult := &pagination.PaginationResult{
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    int64(len(filtered)),
		Pages:    pages,
	}
	return pageItems, pageResult, nil
}

func (r *invitationStatsTestRedeemRepo) SumPositiveBalanceByUser(_ context.Context, _ int64) (float64, error) {
	return 0, nil
}

func TestGetInvitationImpactStats_NoUsageWhenUnused(t *testing.T) {
	t.Parallel()

	repo := &invitationStatsTestRedeemRepo{
		invitation: &RedeemCode{
			ID:       1,
			Code:     "INV-UNUSED",
			Type:     RedeemTypeInvitation,
			Category: "campaign-a",
		},
	}

	userRepo := newCheckinTestUserRepo(map[int64]*User{})
	svc := NewRedeemService(repo, userRepo, nil, nil, nil, nil, nil, nil)

	stats, err := svc.GetInvitationImpactStats(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Equal(t, int64(1), stats.InvitationCodeID)
	require.Equal(t, "INV-UNUSED", stats.Code)
	require.Equal(t, "campaign-a", stats.Category)
	require.Nil(t, stats.UsedByUserID)
	require.Equal(t, int64(0), stats.SubscriptionRedeemsTotal)
	require.Len(t, stats.Redeems, 0)
	require.Len(t, stats.ByCategory, 0)
	require.Len(t, stats.ByGroup, 0)
}

func TestGetInvitationImpactStats_AggregatesRedeems(t *testing.T) {
	t.Parallel()

	userID := int64(42)
	repo := &invitationStatsTestRedeemRepo{}
	repo.invitation = &RedeemCode{
		ID:       1,
		Code:     "INV-USED",
		Type:     RedeemTypeInvitation,
		Category: "campaign-b",
	}
	repo.invitation.UsedBy = &userID

	groupID := int64(10)
	now := time.Now()
	repo.redeems = []RedeemCode{
		{
			ID:       2,
			Code:     "SUB-1",
			Type:     RedeemTypeSubscription,
			Category: "starter",
			Value:    1,
			UsedBy:   &userID,
			UsedAt:   &now,
		},
		{
			ID:       3,
			Code:     "SUB-2",
			Type:     RedeemTypeSubscription,
			Category: "pro",
			Value:    2,
			UsedBy:   &userID,
			GroupID:  &groupID,
			Group:    &Group{ID: groupID, Name: "Pro Group"},
			UsedAt:   &now,
		},
	}

	userRepo := newCheckinTestUserRepo(map[int64]*User{
		userID: {
			ID:        userID,
			Role:      RoleUser,
			Email:     "invited@example.com",
			CreatedAt: now,
		},
	})

	svc := NewRedeemService(repo, userRepo, nil, nil, nil, nil, nil, nil)
	stats, err := svc.GetInvitationImpactStats(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, stats)

	require.Equal(t, &userID, stats.UsedByUserID)
	require.Equal(t, "invited@example.com", stats.UsedByEmail)
	require.Equal(t, int64(2), stats.SubscriptionRedeemsTotal)
	require.Len(t, stats.Redeems, 2)
	require.Len(t, stats.ByCategory, 2)
	require.Len(t, stats.ByGroup, 1)

	var totalByCategory int64
	for _, item := range stats.ByCategory {
		totalByCategory += item.RedeemCount
	}
	require.Equal(t, int64(2), totalByCategory)

	require.Equal(t, groupID, stats.ByGroup[0].GroupID)
	require.Equal(t, "Pro Group", stats.ByGroup[0].GroupName)
	require.Equal(t, int64(1), stats.ByGroup[0].RedeemCount)
}
