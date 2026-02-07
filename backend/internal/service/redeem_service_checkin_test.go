//go:build unit

package service

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestRedeemServiceDailyCheckinSuccess(t *testing.T) {
	t.Parallel()

	redeemRepo := newCheckinTestRedeemRepo()
	userRepo := newCheckinTestUserRepo(map[int64]*User{
		1: {
			ID:      1,
			Role:    RoleUser,
			Balance: 10,
		},
	})
	settingService := newCheckinTestSettingService(true, 6.23, 8.66)

	svc := NewRedeemService(redeemRepo, userRepo, nil, settingService, nil, nil, nil, nil)
	result, err := svc.DailyCheckin(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.GreaterOrEqual(t, result.RewardAmount, 6.23)
	require.LessOrEqual(t, result.RewardAmount, 8.66)
	require.InDelta(t, 10+result.RewardAmount, result.NewBalance, 1e-9)

	updatedUser, err := userRepo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.InDelta(t, result.NewBalance, updatedUser.Balance, 1e-9)
}

func TestRedeemServiceDailyCheckinDuplicateSameDay(t *testing.T) {
	t.Parallel()

	redeemRepo := newCheckinTestRedeemRepo()
	userRepo := newCheckinTestUserRepo(map[int64]*User{
		1: {
			ID:      1,
			Role:    RoleUser,
			Balance: 5,
		},
	})
	settingService := newCheckinTestSettingService(true, 1, 3)

	svc := NewRedeemService(redeemRepo, userRepo, nil, settingService, nil, nil, nil, nil)
	first, err := svc.DailyCheckin(context.Background(), 1)
	require.NoError(t, err)

	_, err = svc.DailyCheckin(context.Background(), 1)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrDailyCheckinDone))

	updatedUser, err := userRepo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.InDelta(t, first.NewBalance, updatedUser.Balance, 1e-9)
}

func TestRedeemServiceDailyCheckinWhenDisabled(t *testing.T) {
	t.Parallel()

	redeemRepo := newCheckinTestRedeemRepo()
	userRepo := newCheckinTestUserRepo(map[int64]*User{
		1: {
			ID:      1,
			Role:    RoleUser,
			Balance: 0,
		},
	})
	settingService := newCheckinTestSettingService(false, 1, 3)

	svc := NewRedeemService(redeemRepo, userRepo, nil, settingService, nil, nil, nil, nil)
	_, err := svc.DailyCheckin(context.Background(), 1)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrDailyCheckinDisabled))
}

func TestRedeemServiceDailyCheckinAdminForbidden(t *testing.T) {
	t.Parallel()

	redeemRepo := newCheckinTestRedeemRepo()
	userRepo := newCheckinTestUserRepo(map[int64]*User{
		1: {
			ID:      1,
			Role:    RoleAdmin,
			Balance: 9,
		},
	})
	settingService := newCheckinTestSettingService(true, 1, 3)

	svc := NewRedeemService(redeemRepo, userRepo, nil, settingService, nil, nil, nil, nil)
	_, err := svc.DailyCheckin(context.Background(), 1)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrDailyCheckinRole))
}

func TestRedeemServiceGetDailyCheckinStatus(t *testing.T) {
	t.Parallel()

	redeemRepo := newCheckinTestRedeemRepo()
	userRepo := newCheckinTestUserRepo(map[int64]*User{
		1: {
			ID:      1,
			Role:    RoleUser,
			Balance: 2,
		},
	})
	settingService := newCheckinTestSettingService(true, 2.5, 3.5)

	svc := NewRedeemService(redeemRepo, userRepo, nil, settingService, nil, nil, nil, nil)

	statusBefore, err := svc.GetDailyCheckinStatus(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, statusBefore.Enabled)
	require.False(t, statusBefore.CheckedInToday)
	require.InDelta(t, 2.5, statusBefore.RewardMin, 1e-9)
	require.InDelta(t, 3.5, statusBefore.RewardMax, 1e-9)
	require.Nil(t, statusBefore.RewardAmount)

	result, err := svc.DailyCheckin(context.Background(), 1)
	require.NoError(t, err)

	statusAfter, err := svc.GetDailyCheckinStatus(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, statusAfter.CheckedInToday)
	require.NotNil(t, statusAfter.RewardAmount)
	require.InDelta(t, result.RewardAmount, *statusAfter.RewardAmount, 1e-9)
}

type checkinTestRedeemRepo struct {
	nextID int64
	byCode map[string]*RedeemCode
}

func newCheckinTestRedeemRepo() *checkinTestRedeemRepo {
	return &checkinTestRedeemRepo{
		nextID: 1,
		byCode: make(map[string]*RedeemCode),
	}
}

func (r *checkinTestRedeemRepo) Create(_ context.Context, code *RedeemCode) error {
	if _, exists := r.byCode[code.Code]; exists {
		return errors.New("duplicate key value violates unique constraint")
	}

	copyCode := *code
	copyCode.ID = r.nextID
	if copyCode.CreatedAt.IsZero() {
		copyCode.CreatedAt = time.Now()
	}
	r.nextID++
	r.byCode[copyCode.Code] = &copyCode

	code.ID = copyCode.ID
	code.CreatedAt = copyCode.CreatedAt
	return nil
}

func (r *checkinTestRedeemRepo) CreateBatch(_ context.Context, _ []RedeemCode) error { return nil }

func (r *checkinTestRedeemRepo) GetByID(_ context.Context, id int64) (*RedeemCode, error) {
	for _, item := range r.byCode {
		if item.ID == id {
			copyItem := *item
			return &copyItem, nil
		}
	}
	return nil, ErrRedeemCodeNotFound
}

func (r *checkinTestRedeemRepo) GetByCode(_ context.Context, code string) (*RedeemCode, error) {
	item, ok := r.byCode[code]
	if !ok {
		return nil, ErrRedeemCodeNotFound
	}
	copyItem := *item
	return &copyItem, nil
}

func (r *checkinTestRedeemRepo) Update(_ context.Context, _ *RedeemCode) error { return nil }

func (r *checkinTestRedeemRepo) Delete(_ context.Context, _ int64) error { return nil }

func (r *checkinTestRedeemRepo) Use(_ context.Context, _, _ int64) error { return nil }

func (r *checkinTestRedeemRepo) List(_ context.Context, _ pagination.PaginationParams) ([]RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *checkinTestRedeemRepo) ListWithFilters(_ context.Context, _ pagination.PaginationParams, _, _, _ string) ([]RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *checkinTestRedeemRepo) ListByUser(_ context.Context, userID int64, _ int) ([]RedeemCode, error) {
	items := make([]RedeemCode, 0)
	for _, item := range r.byCode {
		if item.UsedBy != nil && *item.UsedBy == userID {
			items = append(items, *item)
		}
	}
	return items, nil
}

func (r *checkinTestRedeemRepo) ListByUserPaginated(_ context.Context, _ int64, _ pagination.PaginationParams, _ string) ([]RedeemCode, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *checkinTestRedeemRepo) SumPositiveBalanceByUser(_ context.Context, _ int64) (float64, error) {
	return 0, nil
}

type checkinTestUserRepo struct {
	users map[int64]*User
}

func newCheckinTestUserRepo(users map[int64]*User) *checkinTestUserRepo {
	return &checkinTestUserRepo{users: users}
}

func (r *checkinTestUserRepo) Create(_ context.Context, _ *User) error { return nil }

func (r *checkinTestUserRepo) GetByID(_ context.Context, id int64) (*User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	copyUser := *user
	return &copyUser, nil
}

func (r *checkinTestUserRepo) GetByEmail(_ context.Context, _ string) (*User, error) {
	return nil, ErrUserNotFound
}

func (r *checkinTestUserRepo) GetFirstAdmin(_ context.Context) (*User, error) {
	return nil, ErrUserNotFound
}

func (r *checkinTestUserRepo) Update(_ context.Context, user *User) error {
	existing, ok := r.users[user.ID]
	if !ok {
		return ErrUserNotFound
	}
	*existing = *user
	return nil
}

func (r *checkinTestUserRepo) Delete(_ context.Context, _ int64) error { return nil }

func (r *checkinTestUserRepo) List(_ context.Context, _ pagination.PaginationParams) ([]User, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *checkinTestUserRepo) ListWithFilters(_ context.Context, _ pagination.PaginationParams, _ UserListFilters) ([]User, *pagination.PaginationResult, error) {
	return nil, nil, nil
}

func (r *checkinTestUserRepo) UpdateBalance(_ context.Context, id int64, amount float64) error {
	user, ok := r.users[id]
	if !ok {
		return ErrUserNotFound
	}
	user.Balance += amount
	return nil
}

func (r *checkinTestUserRepo) DeductBalance(_ context.Context, _ int64, _ float64) error { return nil }

func (r *checkinTestUserRepo) UpdateConcurrency(_ context.Context, _ int64, _ int) error { return nil }

func (r *checkinTestUserRepo) ExistsByEmail(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (r *checkinTestUserRepo) RemoveGroupFromAllowedGroups(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

func (r *checkinTestUserRepo) UpdateTotpSecret(_ context.Context, _ int64, _ *string) error {
	return nil
}

func (r *checkinTestUserRepo) EnableTotp(_ context.Context, _ int64) error { return nil }

func (r *checkinTestUserRepo) DisableTotp(_ context.Context, _ int64) error { return nil }

type checkinTestSettingRepo struct {
	values map[string]string
}

func newCheckinTestSettingRepo(values map[string]string) *checkinTestSettingRepo {
	return &checkinTestSettingRepo{values: values}
}

func newCheckinTestSettingService(enabled bool, minAmount, maxAmount float64) *SettingService {
	values := map[string]string{
		SettingKeyDailyCheckinEnabled:   strconv.FormatBool(enabled),
		SettingKeyDailyCheckinRewardMin: strconv.FormatFloat(minAmount, 'f', 8, 64),
		SettingKeyDailyCheckinRewardMax: strconv.FormatFloat(maxAmount, 'f', 8, 64),
	}
	cfg := &config.Config{
		Default: config.DefaultConfig{
			UserConcurrency: 1,
			UserBalance:     0,
		},
	}
	return NewSettingService(newCheckinTestSettingRepo(values), cfg)
}

func (r *checkinTestSettingRepo) Get(_ context.Context, key string) (*Setting, error) {
	value, ok := r.values[key]
	if !ok {
		return nil, ErrSettingNotFound
	}
	return &Setting{Key: key, Value: value}, nil
}

func (r *checkinTestSettingRepo) GetValue(_ context.Context, key string) (string, error) {
	value, ok := r.values[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}

func (r *checkinTestSettingRepo) Set(_ context.Context, key, value string) error {
	r.values[key] = value
	return nil
}

func (r *checkinTestSettingRepo) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.values[key]; ok {
			result[key] = value
		}
	}
	return result, nil
}

func (r *checkinTestSettingRepo) SetMultiple(_ context.Context, settings map[string]string) error {
	for key, value := range settings {
		r.values[key] = value
	}
	return nil
}

func (r *checkinTestSettingRepo) GetAll(_ context.Context) (map[string]string, error) {
	result := make(map[string]string, len(r.values))
	for key, value := range r.values {
		result[key] = value
	}
	return result, nil
}

func (r *checkinTestSettingRepo) Delete(_ context.Context, key string) error {
	delete(r.values, key)
	return nil
}
