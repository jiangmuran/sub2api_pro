package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/activity"
	"github.com/Wei-Shaw/sub2api/ent/activityparticipation"
	"github.com/Wei-Shaw/sub2api/ent/activityreward"
)

// ActivityAdminService 活动管理服务
type ActivityAdminService struct {
	client *ent.Client
}

// NewActivityAdminService 创建活动管理服务
func NewActivityAdminService(client *ent.Client) *ActivityAdminService {
	return &ActivityAdminService{client: client}
}

// ===== 活动管理 =====

// CreateActivity 创建活动
func (s *ActivityAdminService) CreateActivity(ctx context.Context, req *CreateActivityRequest) (*ent.Activity, error) {
	creator := s.client.Activity.Create().
		SetName(req.Name).
		SetDescription(req.Description).
		SetIcon(req.Icon).
		SetType(activity.Type(req.Type)).
		SetStatus(activity.Status(req.Status)).
		SetSortOrder(req.SortOrder).
		SetCreatedBy(req.CreatedBy)

	if req.StartsAt != nil {
		creator.SetStartsAt(*req.StartsAt)
	}
	if req.EndsAt != nil {
		creator.SetEndsAt(*req.EndsAt)
	}
	if req.VisibilityRules != nil {
		creator.SetVisibilityRules(req.VisibilityRules)
	}
	if req.ParticipationConfig != nil {
		creator.SetParticipationConfig(req.ParticipationConfig)
	}
	if req.ActivityConfig != nil {
		creator.SetActivityConfig(req.ActivityConfig)
	}

	return creator.Save(ctx)
}

// CreateActivityRequest 创建活动请求
type CreateActivityRequest struct {
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	Icon                string                 `json:"icon"`
	Type                string                 `json:"type"`
	Status              string                 `json:"status"`
	SortOrder           int                    `json:"sort_order"`
	StartsAt            *time.Time             `json:"starts_at"`
	EndsAt              *time.Time             `json:"ends_at"`
	VisibilityRules     map[string]interface{} `json:"visibility_rules"`
	ParticipationConfig map[string]interface{} `json:"participation_config"`
	ActivityConfig      map[string]interface{} `json:"activity_config"`
	CreatedBy           string                 `json:"created_by"`
}

// UpdateActivity 更新活动
func (s *ActivityAdminService) UpdateActivity(ctx context.Context, id int64, req *UpdateActivityRequest) (*ent.Activity, error) {
	updater := s.client.Activity.UpdateOneID(id)

	if req.Name != nil {
		updater.SetName(*req.Name)
	}
	if req.Description != nil {
		updater.SetDescription(*req.Description)
	}
	if req.Icon != nil {
		updater.SetIcon(*req.Icon)
	}
	if req.Type != nil {
		updater.SetType(activity.Type(*req.Type))
	}
	if req.Status != nil {
		updater.SetStatus(activity.Status(*req.Status))
	}
	if req.SortOrder != nil {
		updater.SetSortOrder(*req.SortOrder)
	}
	if req.StartsAt != nil {
		updater.SetStartsAt(*req.StartsAt)
	}
	if req.EndsAt != nil {
		updater.SetEndsAt(*req.EndsAt)
	}
	if req.VisibilityRules != nil {
		updater.SetVisibilityRules(req.VisibilityRules)
	}
	if req.ParticipationConfig != nil {
		updater.SetParticipationConfig(req.ParticipationConfig)
	}
	if req.ActivityConfig != nil {
		updater.SetActivityConfig(req.ActivityConfig)
	}

	return updater.Save(ctx)
}

// UpdateActivityRequest 更新活动请求
type UpdateActivityRequest struct {
	Name                *string                `json:"name"`
	Description         *string                `json:"description"`
	Icon                *string                `json:"icon"`
	Type                *string                `json:"type"`
	Status              *string                `json:"status"`
	SortOrder           *int                   `json:"sort_order"`
	StartsAt            *time.Time             `json:"starts_at"`
	EndsAt              *time.Time             `json:"ends_at"`
	VisibilityRules     map[string]interface{} `json:"visibility_rules"`
	ParticipationConfig map[string]interface{} `json:"participation_config"`
	ActivityConfig      map[string]interface{} `json:"activity_config"`
}

// DeleteActivity 删除活动
func (s *ActivityAdminService) DeleteActivity(ctx context.Context, id int64) error {
	return s.client.Activity.DeleteOneID(id).Exec(ctx)
}

// ListActivities 列出所有活动（管理端）
func (s *ActivityAdminService) ListActivities(ctx context.Context, filter *ActivityFilter, limit, offset int) ([]*ent.Activity, int, error) {
	query := s.client.Activity.Query()

	if filter != nil {
		if filter.Type != nil {
			query.Where(activity.TypeEQ(activity.Type(*filter.Type)))
		}
		if filter.Status != nil {
			query.Where(activity.StatusEQ(activity.Status(*filter.Status)))
		}
		if filter.Keyword != nil && *filter.Keyword != "" {
			query.Where(activity.NameContains(*filter.Keyword))
		}
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	activities, err := query.
		WithRewards(func(rq *ent.ActivityRewardQuery) {
			rq.Order(ent.Asc(activityreward.FieldSortOrder))
		}).
		Order(ent.Desc(activity.FieldSortOrder), ent.Desc(activity.FieldCreatedAt)).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

// ActivityFilter 活动过滤条件
type ActivityFilter struct {
	Type    *string `json:"type"`
	Status  *string `json:"status"`
	Keyword *string `json:"keyword"`
}

// ===== 奖励管理 =====

// CreateReward 创建奖励
func (s *ActivityAdminService) CreateReward(ctx context.Context, req *CreateRewardRequest) (*ent.ActivityReward, error) {
	creator := s.client.ActivityReward.Create().
		SetActivityID(req.ActivityID).
		SetName(req.Name).
		SetDescription(req.Description).
		SetIcon(req.Icon).
		SetRewardType(activityreward.RewardType(req.RewardType)).
		SetRewardValue(req.RewardValue).
		SetWeight(req.Weight).
		SetTotalStock(req.TotalStock).
		SetRemainingStock(req.RemainingStock).
		SetTier(activityreward.Tier(req.Tier)).
		SetStatus(activityreward.Status(req.Status)).
		SetSortOrder(req.SortOrder)

	if req.Probability != nil {
		creator.SetProbability(*req.Probability)
	}

	return creator.Save(ctx)
}

// CreateRewardRequest 创建奖励请求
type CreateRewardRequest struct {
	ActivityID     int64    `json:"activity_id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Icon           string   `json:"icon"`
	RewardType     string   `json:"reward_type"`
	RewardValue    string   `json:"reward_value"`
	Weight         int      `json:"weight"`
	Probability    *float64 `json:"probability"`
	TotalStock     int64    `json:"total_stock"`
	RemainingStock int64    `json:"remaining_stock"`
	Tier           string   `json:"tier"`
	Status         string   `json:"status"`
	SortOrder      int      `json:"sort_order"`
}

// UpdateReward 更新奖励
func (s *ActivityAdminService) UpdateReward(ctx context.Context, id int64, req *UpdateRewardRequest) (*ent.ActivityReward, error) {
	updater := s.client.ActivityReward.UpdateOneID(id)

	if req.Name != nil {
		updater.SetName(*req.Name)
	}
	if req.Description != nil {
		updater.SetDescription(*req.Description)
	}
	if req.Icon != nil {
		updater.SetIcon(*req.Icon)
	}
	if req.RewardType != nil {
		updater.SetRewardType(activityreward.RewardType(*req.RewardType))
	}
	if req.RewardValue != nil {
		updater.SetRewardValue(*req.RewardValue)
	}
	if req.Weight != nil {
		updater.SetWeight(*req.Weight)
	}
	if req.Probability != nil {
		updater.SetProbability(*req.Probability)
	}
	if req.TotalStock != nil {
		updater.SetTotalStock(*req.TotalStock)
	}
	if req.RemainingStock != nil {
		updater.SetRemainingStock(*req.RemainingStock)
	}
	if req.Tier != nil {
		updater.SetTier(activityreward.Tier(*req.Tier))
	}
	if req.Status != nil {
		updater.SetStatus(activityreward.Status(*req.Status))
	}
	if req.SortOrder != nil {
		updater.SetSortOrder(*req.SortOrder)
	}

	return updater.Save(ctx)
}

// UpdateRewardRequest 更新奖励请求
type UpdateRewardRequest struct {
	Name           *string  `json:"name"`
	Description    *string  `json:"description"`
	Icon           *string  `json:"icon"`
	RewardType     *string  `json:"reward_type"`
	RewardValue    *string  `json:"reward_value"`
	Weight         *int     `json:"weight"`
	Probability    *float64 `json:"probability"`
	TotalStock     *int64   `json:"total_stock"`
	RemainingStock *int64   `json:"remaining_stock"`
	Tier           *string  `json:"tier"`
	Status         *string  `json:"status"`
	SortOrder      *int     `json:"sort_order"`
}

// DeleteReward 删除奖励
func (s *ActivityAdminService) DeleteReward(ctx context.Context, id int64) error {
	return s.client.ActivityReward.DeleteOneID(id).Exec(ctx)
}

// ===== 统计数据 =====

// GetActivityStats 获取活动统计数据
func (s *ActivityAdminService) GetActivityStats(ctx context.Context, activityID int64, startDate, endDate *time.Time) (*ActivityStats, error) {
	query := s.client.ActivityParticipation.Query().
		Where(activityparticipation.ActivityID(activityID))

	if startDate != nil {
		query.Where(activityparticipation.ParticipatedAtGTE(*startDate))
	}
	if endDate != nil {
		query.Where(activityparticipation.ParticipatedAtLTE(*endDate))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, err
	}

	successCount, err := query.Clone().
		Where(activityparticipation.ResultEQ(activityparticipation.ResultSuccess)).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	// 统计每日参与数
	participations, err := query.Clone().
		Where(activityparticipation.ResultEQ(activityparticipation.ResultSuccess)).
		Order(ent.Asc(activityparticipation.FieldDailyWindow)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	dailyStats := make(map[string]int)
	for _, p := range participations {
		day := p.DailyWindow.Format("2006-01-02")
		dailyStats[day]++
	}

	// 获取独立用户数
	uniqueUsers, err := query.Clone().
		Where(activityparticipation.ResultEQ(activityparticipation.ResultSuccess)).
		GroupBy(activityparticipation.FieldUserID).
		Aggregate(ent.Count()).
		Ints(ctx)
	if err != nil {
		return nil, err
	}

	uniqueUserCount := 0
	if len(uniqueUsers) > 0 {
		uniqueUserCount = len(uniqueUsers)
	}

	return &ActivityStats{
		TotalParticipations: total,
		SuccessCount:        successCount,
		UniqueUsers:         uniqueUserCount,
		DailyStats:          dailyStats,
	}, nil
}

// ActivityStats 活动统计
type ActivityStats struct {
	TotalParticipations int            `json:"total_participations"`
	SuccessCount        int            `json:"success_count"`
	UniqueUsers         int            `json:"unique_users"`
	DailyStats          map[string]int `json:"daily_stats"`
}

// GetDashboardStats 获取仪表盘统计
func (s *ActivityAdminService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	// 活动总数
	totalActivities, err := s.client.Activity.Query().Count(ctx)
	if err != nil {
		return nil, err
	}

	// 激活的活动数
	activeActivities, err := s.client.Activity.Query().
		Where(activity.StatusEQ(activity.StatusActive)).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	// 今日参与总数
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	todayParticipations, err := s.client.ActivityParticipation.Query().
		Where(activityparticipation.ParticipatedAtGTE(todayStart)).
		Count(ctx)
	if err != nil {
		return nil, err
	}

	// 总参与数
	totalParticipations, err := s.client.ActivityParticipation.Query().Count(ctx)
	if err != nil {
		return nil, err
	}

	// 最受欢迎的活动（Top 5）
	type ActivityCount struct {
		ActivityID int64
		Count      int
	}

	// 获取最近7天的热门活动
	sevenDaysAgo := today.AddDate(0, 0, -7)
	topActivities, err := s.client.ActivityParticipation.Query().
		Where(activityparticipation.ParticipatedAtGTE(sevenDaysAgo)).
		GroupBy(activityparticipation.FieldActivityID).
		Aggregate(ent.Count()).
		Ints(ctx)
	if err != nil {
		return nil, err
	}

	return &DashboardStats{
		TotalActivities:     totalActivities,
		ActiveActivities:    activeActivities,
		TodayParticipations: todayParticipations,
		TotalParticipations: totalParticipations,
		TopActivitiesCount:  len(topActivities),
	}, nil
}

// DashboardStats 仪表盘统计
type DashboardStats struct {
	TotalActivities     int `json:"total_activities"`
	ActiveActivities    int `json:"active_activities"`
	TodayParticipations int `json:"today_participations"`
	TotalParticipations int `json:"total_participations"`
	TopActivitiesCount  int `json:"top_activities_count"`
}
