package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/activity"
	"github.com/Wei-Shaw/sub2api/ent/activityparticipation"
	"github.com/Wei-Shaw/sub2api/ent/activityreward"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/ent/usersubscription"
	"github.com/Wei-Shaw/sub2api/internal/domain"
)

// ActivityService 活动服务
type ActivityService struct {
	client *ent.Client
}

// NewActivityService 创建活动服务
func NewActivityService(client *ent.Client) *ActivityService {
	return &ActivityService{client: client}
}

// ===== 活动查询 =====

// ListActivitiesForUser 获取用户可见的活动列表
func (s *ActivityService) ListActivitiesForUser(ctx context.Context, userID int64) ([]*ent.Activity, error) {
	now := time.Now()

	// 查询所有激活的活动
	activities, err := s.client.Activity.Query().
		Where(
			activity.StatusEQ(activity.StatusActive),
			activity.Or(
				activity.StartsAtIsNil(),
				activity.StartsAtLTE(now),
			),
			activity.Or(
				activity.EndsAtIsNil(),
				activity.EndsAtGTE(now),
			),
		).
		WithRewards(func(rq *ent.ActivityRewardQuery) {
			rq.Where(activityreward.StatusEQ(activityreward.StatusActive)).
				Order(ent.Asc(activityreward.FieldSortOrder))
		}).
		Order(ent.Desc(activity.FieldSortOrder), ent.Desc(activity.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities: %w", err)
	}

	// 获取用户信息用于可见性过滤
	u, err := s.client.User.Query().
		Where(user.ID(userID)).
		WithSubscriptions(func(sq *ent.UserSubscriptionQuery) {
			sq.Where(usersubscription.StatusEQ(domain.SubscriptionStatusActive))
		}).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// 过滤可见性
	var visibleActivities []*ent.Activity
	for _, act := range activities {
		if s.isActivityVisibleToUser(act, u) {
			visibleActivities = append(visibleActivities, act)
		}
	}

	return visibleActivities, nil
}

// isActivityVisibleToUser 检查活动对用户是否可见
func (s *ActivityService) isActivityVisibleToUser(act *ent.Activity, u *ent.User) bool {
	if act.VisibilityRules == nil || len(act.VisibilityRules) == 0 {
		return true // 无规则则对所有人可见
	}

	rules := act.VisibilityRules

	// 检查最小余额
	if minBalance, ok := rules["min_balance"].(float64); ok && u.Balance < minBalance {
		return false
	}

	// 检查是否需要订阅
	if subscriptionRequired, ok := rules["subscription_required"].(bool); ok && subscriptionRequired {
		if len(u.Edges.Subscriptions) == 0 {
			return false
		}
	}

	// 检查注册时间
	if minRegisterDays, ok := rules["min_register_days"].(float64); ok {
		daysSinceReg := time.Since(u.CreatedAt).Hours() / 24
		if daysSinceReg < minRegisterDays {
			return false
		}
	}

	// 检查用户标签（如果实现了标签系统）
	// if requiredTags, ok := rules["user_tags"].([]interface{}); ok && len(requiredTags) > 0 {
	// 	// TODO: 实现用户标签检查
	// }

	return true
}

// GetActivityDetail 获取活动详情
func (s *ActivityService) GetActivityDetail(ctx context.Context, activityID int64) (*ent.Activity, error) {
	return s.client.Activity.Query().
		Where(activity.ID(activityID)).
		WithRewards(func(rq *ent.ActivityRewardQuery) {
			rq.Order(ent.Asc(activityreward.FieldSortOrder))
		}).
		Only(ctx)
}

// ===== 活动参与 =====

// ParticipateInActivity 参与活动（通用入口）
func (s *ActivityService) ParticipateInActivity(ctx context.Context, userID, activityID int64, ipAddress, userAgent string) (*ParticipationResult, error) {
	// 获取活动
	act, err := s.GetActivityDetail(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	// 检查活动状态
	if act.Status != activity.StatusActive {
		return nil, fmt.Errorf("activity is not active")
	}

	// 检查时间范围
	now := time.Now()
	if act.StartsAt != nil && now.Before(*act.StartsAt) {
		return nil, fmt.Errorf("activity not started yet")
	}
	if act.EndsAt != nil && now.After(*act.EndsAt) {
		return nil, fmt.Errorf("activity has ended")
	}

	// 检查参与次数限制
	if err := s.checkParticipationLimit(ctx, userID, activityID, act.ParticipationConfig); err != nil {
		return nil, err
	}

	// 根据活动类型执行不同逻辑
	switch act.Type {
	case activity.TypeCheckIn:
		return s.handleCheckIn(ctx, userID, activityID, act, ipAddress, userAgent)
	case activity.TypeLottery:
		return s.handleLottery(ctx, userID, activityID, act, ipAddress, userAgent)
	case activity.TypeRedeem:
		return s.handleRedeem(ctx, userID, activityID, act, ipAddress, userAgent)
	case activity.TypeTask:
		return s.handleTask(ctx, userID, activityID, act, ipAddress, userAgent)
	case activity.TypeNewbie:
		return s.handleNewbie(ctx, userID, activityID, act, ipAddress, userAgent)
	default:
		return nil, fmt.Errorf("unsupported activity type: %s", act.Type)
	}
}

// ParticipationResult 参与结果
type ParticipationResult struct {
	Success       bool                       `json:"success"`
	Message       string                     `json:"message"`
	Rewards       []*RewardInfo              `json:"rewards"`
	ExtraData     map[string]interface{}     `json:"extra_data"`
	Participation *ent.ActivityParticipation `json:"participation"`
}

// RewardInfo 奖励信息
type RewardInfo struct {
	RewardID    int64                  `json:"reward_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	Type        string                 `json:"type"`
	Value       map[string]interface{} `json:"value"`
	Tier        string                 `json:"tier"`
}

// checkParticipationLimit 检查参与次数限制
func (s *ActivityService) checkParticipationLimit(ctx context.Context, userID, activityID int64, config map[string]interface{}) error {
	if config == nil {
		return nil
	}

	now := time.Now()
	dailyWindow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekStart := now.AddDate(0, 0, -int(now.Weekday()-time.Monday))
	weeklyWindow := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)
	monthlyWindow := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// 检查每日限制
	if maxPerDay, ok := config["max_per_day"].(float64); ok && maxPerDay > 0 {
		count, err := s.client.ActivityParticipation.Query().
			Where(
				activityparticipation.UserID(userID),
				activityparticipation.ActivityID(activityID),
				activityparticipation.DailyWindow(dailyWindow),
				activityparticipation.ResultEQ(activityparticipation.ResultSuccess),
			).
			Count(ctx)
		if err != nil {
			return fmt.Errorf("failed to count daily participations: %w", err)
		}
		if count >= int(maxPerDay) {
			return fmt.Errorf("daily participation limit reached")
		}
	}

	// 检查每周限制
	if maxPerWeek, ok := config["max_per_week"].(float64); ok && maxPerWeek > 0 {
		count, err := s.client.ActivityParticipation.Query().
			Where(
				activityparticipation.UserID(userID),
				activityparticipation.ActivityID(activityID),
				activityparticipation.WeeklyWindow(weeklyWindow),
				activityparticipation.ResultEQ(activityparticipation.ResultSuccess),
			).
			Count(ctx)
		if err != nil {
			return fmt.Errorf("failed to count weekly participations: %w", err)
		}
		if count >= int(maxPerWeek) {
			return fmt.Errorf("weekly participation limit reached")
		}
	}

	// 检查每月限制
	if maxPerMonth, ok := config["max_per_month"].(float64); ok && maxPerMonth > 0 {
		count, err := s.client.ActivityParticipation.Query().
			Where(
				activityparticipation.UserID(userID),
				activityparticipation.ActivityID(activityID),
				activityparticipation.MonthlyWindow(monthlyWindow),
				activityparticipation.ResultEQ(activityparticipation.ResultSuccess),
			).
			Count(ctx)
		if err != nil {
			return fmt.Errorf("failed to count monthly participations: %w", err)
		}
		if count >= int(maxPerMonth) {
			return fmt.Errorf("monthly participation limit reached")
		}
	}

	return nil
}

// handleCheckIn 处理签到活动
func (s *ActivityService) handleCheckIn(ctx context.Context, userID, activityID int64, act *ent.Activity, ipAddress, userAgent string) (*ParticipationResult, error) {
	// 签到活动通常给固定奖励
	rewards := act.Edges.Rewards
	if len(rewards) == 0 {
		return nil, fmt.Errorf("no rewards configured for check-in activity")
	}

	// 计算连续签到天数
	consecutiveDays, err := s.getConsecutiveCheckInDays(ctx, userID, activityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get consecutive days: %w", err)
	}

	// 选择奖励（可以根据连续天数选择不同奖励）
	var selectedReward *ent.ActivityReward
	if act.ActivityConfig != nil {
		if milestones, ok := act.ActivityConfig["check_in_milestones"].(map[string]interface{}); ok {
			// 如果配置了里程碑奖励，按天数匹配
			dayStr := fmt.Sprintf("%d", consecutiveDays+1)
			if rewardID, ok := milestones[dayStr].(float64); ok {
				for _, r := range rewards {
					if r.ID == int64(rewardID) {
						selectedReward = r
						break
					}
				}
			}
		}
	}

	// 默认选择第一个奖励
	if selectedReward == nil {
		selectedReward = rewards[0]
	}

	// 发放奖励
	rewardInfos, err := s.distributeRewards(ctx, userID, []*ent.ActivityReward{selectedReward})
	if err != nil {
		return nil, fmt.Errorf("failed to distribute rewards: %w", err)
	}

	// 创建参与记录
	now := time.Now()
	dailyWindow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekStart := now.AddDate(0, 0, -int(now.Weekday()-time.Monday))
	weeklyWindow := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)
	monthlyWindow := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	participation, err := s.client.ActivityParticipation.Create().
		SetUserID(userID).
		SetActivityID(activityID).
		SetParticipatedAt(now).
		SetDailyWindow(dailyWindow).
		SetWeeklyWindow(weeklyWindow).
		SetMonthlyWindow(monthlyWindow).
		SetResult(activityparticipation.ResultSuccess).
		SetRewardsReceived(s.rewardInfosToJSON(rewardInfos)).
		SetNillableRewardID(&selectedReward.ID).
		SetExtraData(map[string]interface{}{
			"consecutive_days": consecutiveDays + 1,
		}).
		SetIPAddress(ipAddress).
		SetUserAgent(userAgent).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create participation: %w", err)
	}

	// 更新活动统计
	_ = s.client.Activity.UpdateOneID(activityID).
		AddTotalParticipations(1).
		AddTotalRewardsDistributed(int64(len(rewardInfos))).
		Exec(ctx)

	return &ParticipationResult{
		Success:       true,
		Message:       fmt.Sprintf("签到成功！已连续签到 %d 天", consecutiveDays+1),
		Rewards:       rewardInfos,
		ExtraData:     map[string]interface{}{"consecutive_days": consecutiveDays + 1},
		Participation: participation,
	}, nil
}

// getConsecutiveCheckInDays 获取连续签到天数
func (s *ActivityService) getConsecutiveCheckInDays(ctx context.Context, userID, activityID int64) (int, error) {
	// 查询最近的签到记录
	participations, err := s.client.ActivityParticipation.Query().
		Where(
			activityparticipation.UserID(userID),
			activityparticipation.ActivityID(activityID),
			activityparticipation.ResultEQ(activityparticipation.ResultSuccess),
		).
		Order(ent.Desc(activityparticipation.FieldDailyWindow)).
		Limit(100). // 最多查100天
		All(ctx)
	if err != nil {
		return 0, err
	}

	if len(participations) == 0 {
		return 0, nil
	}

	// 计算连续天数
	consecutiveDays := 0
	yesterday := time.Now().AddDate(0, 0, -1)
	expectedWindow := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC)

	for _, p := range participations {
		if p.DailyWindow.Equal(expectedWindow) {
			consecutiveDays++
			expectedWindow = expectedWindow.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return consecutiveDays, nil
}

// handleLottery 处理抽奖活动
func (s *ActivityService) handleLottery(ctx context.Context, userID, activityID int64, act *ent.Activity, ipAddress, userAgent string) (*ParticipationResult, error) {
	rewards := act.Edges.Rewards
	if len(rewards) == 0 {
		return nil, fmt.Errorf("no rewards configured for lottery activity")
	}

	// 过滤可用奖励（有库存的）
	var availableRewards []*ent.ActivityReward
	for _, r := range rewards {
		if r.Status == activityreward.StatusActive && (r.TotalStock == 0 || r.RemainingStock > 0) {
			availableRewards = append(availableRewards, r)
		}
	}

	if len(availableRewards) == 0 {
		return nil, fmt.Errorf("no available rewards")
	}

	// 基于权重随机抽取
	selectedReward := s.selectRewardByWeight(availableRewards)

	// 扣除库存
	if selectedReward.TotalStock > 0 {
		affected, err := s.client.ActivityReward.Update().
			Where(
				activityreward.ID(selectedReward.ID),
				activityreward.RemainingStockGT(0),
			).
			AddRemainingStock(-1).
			AddDistributedCount(1).
			Save(ctx)
		if err != nil || affected == 0 {
			return nil, fmt.Errorf("reward out of stock")
		}
	} else {
		_ = s.client.ActivityReward.UpdateOneID(selectedReward.ID).
			AddDistributedCount(1).
			Exec(ctx)
	}

	// 发放奖励
	rewardInfos, err := s.distributeRewards(ctx, userID, []*ent.ActivityReward{selectedReward})
	if err != nil {
		return nil, fmt.Errorf("failed to distribute rewards: %w", err)
	}

	// 创建参与记录
	now := time.Now()
	dailyWindow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekStart := now.AddDate(0, 0, -int(now.Weekday()-time.Monday))
	weeklyWindow := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)
	monthlyWindow := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	participation, err := s.client.ActivityParticipation.Create().
		SetUserID(userID).
		SetActivityID(activityID).
		SetParticipatedAt(now).
		SetDailyWindow(dailyWindow).
		SetWeeklyWindow(weeklyWindow).
		SetMonthlyWindow(monthlyWindow).
		SetResult(activityparticipation.ResultSuccess).
		SetRewardsReceived(s.rewardInfosToJSON(rewardInfos)).
		SetNillableRewardID(&selectedReward.ID).
		SetIPAddress(ipAddress).
		SetUserAgent(userAgent).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create participation: %w", err)
	}

	// 更新活动统计
	_ = s.client.Activity.UpdateOneID(activityID).
		AddTotalParticipations(1).
		AddTotalRewardsDistributed(int64(len(rewardInfos))).
		Exec(ctx)

	return &ParticipationResult{
		Success:       true,
		Message:       fmt.Sprintf("恭喜获得：%s", selectedReward.Name),
		Rewards:       rewardInfos,
		Participation: participation,
	}, nil
}

// selectRewardByWeight 根据权重选择奖励
func (s *ActivityService) selectRewardByWeight(rewards []*ent.ActivityReward) *ent.ActivityReward {
	totalWeight := 0
	for _, r := range rewards {
		totalWeight += r.Weight
	}

	randNum := rand.Intn(totalWeight)
	cumulative := 0

	for _, r := range rewards {
		cumulative += r.Weight
		if randNum < cumulative {
			return r
		}
	}

	return rewards[len(rewards)-1]
}

// handleRedeem 处理兑换活动
func (s *ActivityService) handleRedeem(ctx context.Context, userID, activityID int64, act *ent.Activity, ipAddress, userAgent string) (*ParticipationResult, error) {
	// 兑换活动需要检查兑换条件
	config := act.ActivityConfig
	if config == nil {
		return nil, fmt.Errorf("redeem activity config not found")
	}

	// 检查余额要求
	if costBalance, ok := config["cost_balance"].(float64); ok && costBalance > 0 {
		u, err := s.client.User.Get(ctx, userID)
		if err != nil {
			return nil, err
		}
		if u.Balance < costBalance {
			return nil, fmt.Errorf("insufficient balance")
		}

		// 扣除余额
		_, err = s.client.User.UpdateOneID(userID).
			AddBalance(-costBalance).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to deduct balance: %w", err)
		}
	}

	// 发放奖励
	rewards := act.Edges.Rewards
	if len(rewards) == 0 {
		return nil, fmt.Errorf("no rewards configured")
	}

	rewardInfos, err := s.distributeRewards(ctx, userID, rewards)
	if err != nil {
		return nil, fmt.Errorf("failed to distribute rewards: %w", err)
	}

	// 创建参与记录
	now := time.Now()
	dailyWindow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekStart := now.AddDate(0, 0, -int(now.Weekday()-time.Monday))
	weeklyWindow := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)
	monthlyWindow := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	participation, err := s.client.ActivityParticipation.Create().
		SetUserID(userID).
		SetActivityID(activityID).
		SetParticipatedAt(now).
		SetDailyWindow(dailyWindow).
		SetWeeklyWindow(weeklyWindow).
		SetMonthlyWindow(monthlyWindow).
		SetResult(activityparticipation.ResultSuccess).
		SetRewardsReceived(s.rewardInfosToJSON(rewardInfos)).
		SetIPAddress(ipAddress).
		SetUserAgent(userAgent).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create participation: %w", err)
	}

	// 更新活动统计
	_ = s.client.Activity.UpdateOneID(activityID).
		AddTotalParticipations(1).
		AddTotalRewardsDistributed(int64(len(rewardInfos))).
		Exec(ctx)

	return &ParticipationResult{
		Success:       true,
		Message:       "兑换成功",
		Rewards:       rewardInfos,
		Participation: participation,
	}, nil
}

// handleTask 处理任务活动
func (s *ActivityService) handleTask(ctx context.Context, userID, activityID int64, act *ent.Activity, ipAddress, userAgent string) (*ParticipationResult, error) {
	// TODO: 实现任务系统
	return nil, fmt.Errorf("task activity not implemented yet")
}

// handleNewbie 处理新手礼包
func (s *ActivityService) handleNewbie(ctx context.Context, userID, activityID int64, act *ent.Activity, ipAddress, userAgent string) (*ParticipationResult, error) {
	// 新手礼包只能领取一次
	count, err := s.client.ActivityParticipation.Query().
		Where(
			activityparticipation.UserID(userID),
			activityparticipation.ActivityID(activityID),
		).
		Count(ctx)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("newbie package already claimed")
	}

	// 发放所有奖励
	rewards := act.Edges.Rewards
	if len(rewards) == 0 {
		return nil, fmt.Errorf("no rewards configured")
	}

	rewardInfos, err := s.distributeRewards(ctx, userID, rewards)
	if err != nil {
		return nil, fmt.Errorf("failed to distribute rewards: %w", err)
	}

	// 创建参与记录
	now := time.Now()
	dailyWindow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekStart := now.AddDate(0, 0, -int(now.Weekday()-time.Monday))
	weeklyWindow := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)
	monthlyWindow := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	participation, err := s.client.ActivityParticipation.Create().
		SetUserID(userID).
		SetActivityID(activityID).
		SetParticipatedAt(now).
		SetDailyWindow(dailyWindow).
		SetWeeklyWindow(weeklyWindow).
		SetMonthlyWindow(monthlyWindow).
		SetResult(activityparticipation.ResultSuccess).
		SetRewardsReceived(s.rewardInfosToJSON(rewardInfos)).
		SetIPAddress(ipAddress).
		SetUserAgent(userAgent).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create participation: %w", err)
	}

	// 更新活动统计
	_ = s.client.Activity.UpdateOneID(activityID).
		AddTotalParticipations(1).
		AddTotalRewardsDistributed(int64(len(rewardInfos))).
		Exec(ctx)

	return &ParticipationResult{
		Success:       true,
		Message:       "新手礼包领取成功",
		Rewards:       rewardInfos,
		Participation: participation,
	}, nil
}

// distributeRewards 发放奖励
func (s *ActivityService) distributeRewards(ctx context.Context, userID int64, rewards []*ent.ActivityReward) ([]*RewardInfo, error) {
	var rewardInfos []*RewardInfo

	for _, reward := range rewards {
		var rewardValue map[string]interface{}
		if reward.RewardValue != "" {
			if err := json.Unmarshal([]byte(reward.RewardValue), &rewardValue); err != nil {
				return nil, fmt.Errorf("invalid reward value JSON: %w", err)
			}
		}

		switch reward.RewardType {
		case activityreward.RewardTypeBalance:
			// 发放余额
			if amount, ok := rewardValue["amount"].(float64); ok {
				_, err := s.client.User.UpdateOneID(userID).
					AddBalance(amount).
					Save(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to add balance: %w", err)
				}
			}

		case activityreward.RewardTypeSubscription:
			// 发放订阅
			if groupID, ok := rewardValue["group_id"].(float64); ok {
				if days, ok := rewardValue["days"].(float64); ok {
					now := time.Now()
					_, err := s.client.UserSubscription.Create().
						SetUserID(userID).
						SetGroupID(int64(groupID)).
						SetStartsAt(now).
						SetExpiresAt(now.AddDate(0, 0, int(days))).
						SetStatus(domain.SubscriptionStatusActive).
						SetNotes("activity_system").
						Save(ctx)
					if err != nil {
						return nil, fmt.Errorf("failed to create subscription: %w", err)
					}
				}
			}

		case activityreward.RewardTypeCoupon:
			// TODO: 实现优惠券系统
		case activityreward.RewardTypePoints:
			// TODO: 实现积分系统
		}

		rewardInfos = append(rewardInfos, &RewardInfo{
			RewardID:    reward.ID,
			Name:        reward.Name,
			Description: reward.Description,
			Icon:        reward.Icon,
			Type:        string(reward.RewardType),
			Value:       rewardValue,
			Tier:        string(reward.Tier),
		})
	}

	return rewardInfos, nil
}

// rewardInfosToJSON 将奖励信息转换为JSON
func (s *ActivityService) rewardInfosToJSON(infos []*RewardInfo) []map[string]interface{} {
	var result []map[string]interface{}
	for _, info := range infos {
		result = append(result, map[string]interface{}{
			"reward_id": info.RewardID,
			"name":      info.Name,
			"type":      info.Type,
			"value":     info.Value,
			"tier":      info.Tier,
		})
	}
	return result
}

// ===== 用户参与历史 =====

// GetUserParticipations 获取用户参与历史
func (s *ActivityService) GetUserParticipations(ctx context.Context, userID int64, limit, offset int) ([]*ent.ActivityParticipation, int, error) {
	total, err := s.client.ActivityParticipation.Query().
		Where(activityparticipation.UserID(userID)).
		Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	participations, err := s.client.ActivityParticipation.Query().
		Where(activityparticipation.UserID(userID)).
		WithActivity().
		WithReward().
		Order(ent.Desc(activityparticipation.FieldParticipatedAt)).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return participations, total, nil
}
