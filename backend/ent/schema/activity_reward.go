package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"time"
)

// ActivityReward holds the schema definition for the ActivityReward entity.
type ActivityReward struct {
	ent.Schema
}

// Fields of the ActivityReward.
func (ActivityReward) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("activity_id").
			Comment("关联的活动ID"),

		field.String("name").
			NotEmpty().
			Comment("奖励名称，如'5元余额'、'7天订阅'"),
		field.String("description").
			Default("").
			Comment("奖励描述"),
		field.String("icon").
			Default("").
			Comment("奖励图标"),

		// 奖励类型和价值
		field.Enum("reward_type").
			Values("balance", "subscription", "coupon", "points", "custom").
			Comment("奖励类型: balance=余额, subscription=订阅, coupon=优惠券, points=积分, custom=自定义"),
		field.String("reward_value").
			Default("").
			Comment("奖励值，JSON格式 {amount:10.5} 或 {group_id:1,days:7}"),

		// 概率和权重
		field.Int("weight").
			Default(100).
			Comment("抽奖权重，越大概率越高（仅抽奖活动使用）"),
		field.Float("probability").
			Optional().
			Comment("概率百分比（0-100），用于显示，可选"),

		// 库存管理
		field.Int64("total_stock").
			Default(0).
			Comment("总库存，0表示无限"),
		field.Int64("remaining_stock").
			Default(0).
			Comment("剩余库存"),

		// 奖励等级（用于展示，如一等奖、二等奖）
		field.Enum("tier").
			Values("grand", "first", "second", "third", "common", "consolation").
			Default("common").
			Comment("奖励等级: grand=特等奖, first=一等奖, second=二等奖, third=三等奖, common=普通奖, consolation=安慰奖"),

		// 状态
		field.Enum("status").
			Values("active", "inactive", "out_of_stock").
			Default("active").
			Comment("奖励状态"),

		field.Int("sort_order").
			Default(0).
			Comment("显示排序"),

		// 统计
		field.Int64("distributed_count").
			Default(0).
			Comment("已发放次数"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the ActivityReward.
func (ActivityReward) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("activity", Activity.Type).
			Ref("rewards").
			Field("activity_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("participations", ActivityParticipation.Type),
	}
}

// Indexes of the ActivityReward.
func (ActivityReward) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("activity_id", "status"),
		index.Fields("activity_id", "sort_order"),
	}
}
