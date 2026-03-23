package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

// Activity holds the schema definition for the Activity entity.
type Activity struct {
	ent.Schema
}

// Fields of the Activity.
func (Activity) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name").
			NotEmpty().
			Comment("活动名称"),
		field.String("description").
			Default("").
			Comment("活动描述"),
		field.String("icon").
			Default("").
			Comment("活动图标URL或icon名称"),
		field.Enum("type").
			Values("check_in", "lottery", "redeem", "task", "newbie", "limited_time").
			Comment("活动类型: check_in=签到, lottery=抽奖, redeem=兑换, task=任务, newbie=新手礼包, limited_time=限时活动"),
		field.Enum("status").
			Values("draft", "active", "paused", "ended", "archived").
			Default("draft").
			Comment("活动状态"),
		field.Int("sort_order").
			Default(0).
			Comment("排序权重，越大越靠前"),

		// 时间配置
		field.Time("starts_at").
			Optional().
			Nillable().
			Comment("活动开始时间，null表示立即开始"),
		field.Time("ends_at").
			Optional().
			Nillable().
			Comment("活动结束时间，null表示永久有效"),

		// 可见性规则（JSON配置）
		field.JSON("visibility_rules", map[string]interface{}{}).
			Optional().
			Comment("可见性规则 {user_tags:[],min_balance:0,subscription_required:bool,min_register_days:0}"),

		// 参与规则配置
		field.JSON("participation_config", map[string]interface{}{}).
			Optional().
			Comment("参与配置 {max_per_day:1,max_per_week:7,max_per_month:30,max_total:0,cost_balance:0}"),

		// 活动特定配置（不同类型活动的特殊配置）
		field.JSON("activity_config", map[string]interface{}{}).
			Optional().
			Comment("活动特定配置，根据type不同而不同"),

		// 统计数据
		field.Int64("total_participations").
			Default(0).
			Comment("总参与次数"),
		field.Int64("total_rewards_distributed").
			Default(0).
			Comment("总发放奖励次数"),

		// 元数据
		field.String("created_by").
			Default("system").
			Comment("创建者"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Activity.
func (Activity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("rewards", ActivityReward.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("participations", ActivityParticipation.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the Activity.
func (Activity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type", "status"),
		index.Fields("status", "sort_order"),
		index.Fields("starts_at", "ends_at"),
	}
}
