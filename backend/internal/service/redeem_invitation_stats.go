package service

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// InvitationRedeemCategoryStat aggregates subscription-type redeem usage by category
// for users who registered via a specific invitation code.
type InvitationRedeemCategoryStat struct {
	Category    string `json:"category"`
	RedeemCount int64  `json:"redeem_count"`
}

// InvitationRedeemGroupStat aggregates subscription-type redeem usage by group.
type InvitationRedeemGroupStat struct {
	GroupID     int64  `json:"group_id"`
	GroupName   string `json:"group_name"`
	RedeemCount int64  `json:"redeem_count"`
}

// InvitationRedeemDetail describes a single subscription-type redeem code
// used by a user who registered via the invitation.
type InvitationRedeemDetail struct {
	RedeemCodeID int64      `json:"redeem_code_id"`
	Code         string     `json:"code"`
	Type         string     `json:"type"`
	Category     string     `json:"category"`
	Value        float64    `json:"value"`
	GroupID      *int64     `json:"group_id,omitempty"`
	GroupName    string     `json:"group_name,omitempty"`
	UsedAt       *time.Time `json:"used_at,omitempty"`
}

// InvitationRedeemImpactStats captures how a specific invitation redeem code
// (used during registration) correlates with later subscription activations
// via subscription-type redeem codes.
type InvitationRedeemImpactStats struct {
	InvitationCodeID         int64                          `json:"invitation_code_id"`
	Code                     string                         `json:"code"`
	Category                 string                         `json:"category"`
	UsedByUserID             *int64                         `json:"used_by_user_id,omitempty"`
	UsedByEmail              string                         `json:"used_by_email,omitempty"`
	RegisteredAt             *time.Time                     `json:"registered_at,omitempty"`
	SubscriptionRedeemsTotal int64                          `json:"subscription_redeems_total"`
	Redeems                  []InvitationRedeemDetail       `json:"redeems"`
	ByCategory               []InvitationRedeemCategoryStat `json:"by_category"`
	ByGroup                  []InvitationRedeemGroupStat    `json:"by_group"`
}

// GetInvitationImpactStats computes, for a given invitation redeem code, how the
// registered user later used subscription-type redeem codes to activate other
// subscriptions. It aggregates by redeem category and subscription group.
func (s *RedeemService) GetInvitationImpactStats(ctx context.Context, codeID int64) (*InvitationRedeemImpactStats, error) {
	code, err := s.redeemRepo.GetByID(ctx, codeID)
	if err != nil {
		return nil, err
	}

	stats := &InvitationRedeemImpactStats{
		InvitationCodeID: code.ID,
		Code:             code.Code,
		Category:         strings.TrimSpace(code.Category),
	}

	// 如果邀请码尚未被使用，没有注册用户，直接返回基本信息
	if code.UsedBy == nil {
		return stats, nil
	}
	stats.UsedByUserID = code.UsedBy
	userID := *code.UsedBy

	// 拉取注册用户的基础信息（邮箱、注册时间）
	if s.userRepo != nil {
		if user, err := s.userRepo.GetByID(ctx, userID); err == nil && user != nil {
			stats.UsedByEmail = strings.TrimSpace(user.Email)
			registeredAt := user.CreatedAt
			stats.RegisteredAt = &registeredAt
		}
	}

	params := pagination.PaginationParams{Page: 1, PageSize: 200}
	redeems := make([]InvitationRedeemDetail, 0)
	byCategory := make(map[string]*InvitationRedeemCategoryStat)
	byGroup := make(map[int64]*InvitationRedeemGroupStat)

	for {
		codes, pageResult, err := s.redeemRepo.ListByUserPaginated(ctx, userID, params, RedeemTypeSubscription)
		if err != nil {
			return nil, err
		}
		if len(codes) == 0 {
			break
		}

		for i := range codes {
			c := codes[i]
			stats.SubscriptionRedeemsTotal++

			detail := InvitationRedeemDetail{
				RedeemCodeID: c.ID,
				Code:         c.Code,
				Type:         c.Type,
				Category:     strings.TrimSpace(c.Category),
				Value:        c.Value,
			}
			if c.GroupID != nil {
				gid := *c.GroupID
				detail.GroupID = c.GroupID
				if c.Group != nil {
					detail.GroupName = c.Group.Name
				}
				groupStat, ok := byGroup[gid]
				if !ok {
					groupStat = &InvitationRedeemGroupStat{GroupID: gid, GroupName: detail.GroupName}
					byGroup[gid] = groupStat
				}
				groupStat.RedeemCount++
			}
			if c.UsedAt != nil {
				usedAt := *c.UsedAt
				detail.UsedAt = &usedAt
			}

			category := strings.TrimSpace(c.Category)
			if category == "" {
				category = "未分组"
			}
			catStat, ok := byCategory[category]
			if !ok {
				catStat = &InvitationRedeemCategoryStat{Category: category}
				byCategory[category] = catStat
			}
			catStat.RedeemCount++

			redeems = append(redeems, detail)
		}

		if pageResult == nil || params.Page >= pageResult.Pages {
			break
		}
		params.Page++
	}

	stats.Redeems = redeems
	stats.ByCategory = make([]InvitationRedeemCategoryStat, 0, len(byCategory))
	for _, v := range byCategory {
		stats.ByCategory = append(stats.ByCategory, *v)
	}
	stats.ByGroup = make([]InvitationRedeemGroupStat, 0, len(byGroup))
	for _, v := range byGroup {
		stats.ByGroup = append(stats.ByGroup, *v)
	}

	return stats, nil
}
