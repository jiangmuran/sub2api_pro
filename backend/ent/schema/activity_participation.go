package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

// ActivityParticipation holds the schema definition for the ActivityParticipation entity.
type ActivityParticipation struct {
	ent.Schema
}

// Fields of the ActivityParticipation.
func (ActivityParticipation) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("activity_id").
			Comment("关联的活动ID"),
		field.Int64("user_id").
			Comment("参与用户ID"),

		// 参与时间
		field.Time("participated_at").
			Default(time.Now).
			Comment("参与时间"),

		// 窗口时间（用于统计每日/每周/每月参与次数）
		field.Time("daily_window").
			Comment("所属日窗口（当天0点）"),
		field.Time("weekly_window").
			Comment("所属周窗口（当周一0点）"),
		field.Time("monthly_window").
			Comment("所属月窗口（当月1号0点）"),

		// 参与结果
		field.Enum("result").
			Values("success", "failed", "pending").
			Default("success").
			Comment("参与结果"),

		// 获得奖励（可能多个）
		field.JSON("rewards_received", []map[string]interface{}{}).
			Optional().
			Comment("获得的奖励列表 [{reward_id:1,type:'balance',value:'10.5'}]"),

		field.Int64("reward_id").
			Optional().
			Nillable().
			Comment("主要奖励ID（抽奖活动）"),

		// 消耗
		field.String("cost_balance").
			Default("0").
			Comment("消耗的余额（某些活动需要付费参与）"),

		// 额外数据
		field.JSON("extra_data", map[string]interface{}{}).
			Optional().
			Comment("额外数据，如签到连续天数、任务完成详情等"),

		// IP和设备信息（防作弊）
		field.String("ip_address").
			Default("").
			Comment("参与时的IP地址"),
		field.String("user_agent").
			Default("").
			Comment("User Agent"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the ActivityParticipation.
func (ActivityParticipation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("activity", Activity.Type).
			Ref("participations").
			Field("activity_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("user", User.Type).
			Ref("activity_participations").
			Field("user_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("reward", ActivityReward.Type).
			Ref("participations").
			Field("reward_id").
			Unique(),
	}
}

// Indexes of the ActivityParticipation.
func (ActivityParticipation) Indexes() []ent.Index {
	return []ent.Index{
		// 查询用户的活动参与历史
		index.Fields("user_id", "activity_id", "participated_at"),
		// 统计活动参与次数
		index.Fields("activity_id", "participated_at"),
		// 每日/每周/每月参与次数统计
		index.Fields("user_id", "activity_id", "daily_window"),
		index.Fields("user_id", "activity_id", "weekly_window"),
		index.Fields("user_id", "activity_id", "monthly_window"),
		// 防作弊检测
		index.Fields("ip_address", "participated_at"),
	}
}
