//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type authInvitationRedeemRepoStub struct {
	codes []RedeemCode
	err   error
}

func (s *authInvitationRedeemRepoStub) Create(context.Context, *RedeemCode) error       { return nil }
func (s *authInvitationRedeemRepoStub) CreateBatch(context.Context, []RedeemCode) error { return nil }
func (s *authInvitationRedeemRepoStub) GetByID(context.Context, int64) (*RedeemCode, error) {
	return nil, ErrRedeemCodeNotFound
}
func (s *authInvitationRedeemRepoStub) GetByCode(context.Context, string) (*RedeemCode, error) {
	return nil, ErrRedeemCodeNotFound
}
func (s *authInvitationRedeemRepoStub) Update(context.Context, *RedeemCode) error { return nil }
func (s *authInvitationRedeemRepoStub) Delete(context.Context, int64) error       { return nil }
func (s *authInvitationRedeemRepoStub) Use(context.Context, int64, int64) error   { return nil }
func (s *authInvitationRedeemRepoStub) List(context.Context, pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (s *authInvitationRedeemRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, string) ([]RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (s *authInvitationRedeemRepoStub) ListByUser(context.Context, int64, int) ([]RedeemCode, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.codes, nil
}
func (s *authInvitationRedeemRepoStub) ListByUserPaginated(context.Context, int64, pagination.PaginationParams, string) ([]RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (s *authInvitationRedeemRepoStub) SumPositiveBalanceByUser(context.Context, int64) (float64, error) {
	return 0, nil
}

func TestAuthService_GetInvitationCodeByUserID_ReturnsInvitationCode(t *testing.T) {
	service := &AuthService{
		redeemRepo: &authInvitationRedeemRepoStub{codes: []RedeemCode{
			{Code: "BALANCE50", Type: RedeemTypeBalance},
			{Code: "INV-ABC", Type: RedeemTypeInvitation},
		}},
	}

	code, err := service.GetInvitationCodeByUserID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, "INV-ABC", code)
}

func TestAuthService_GetInvitationCodeByUserID_ReturnsEmptyWhenMissing(t *testing.T) {
	service := &AuthService{
		redeemRepo: &authInvitationRedeemRepoStub{codes: []RedeemCode{{Code: "BALANCE50", Type: RedeemTypeBalance}}},
	}

	code, err := service.GetInvitationCodeByUserID(context.Background(), 1)
	require.NoError(t, err)
	require.Empty(t, code)
}
