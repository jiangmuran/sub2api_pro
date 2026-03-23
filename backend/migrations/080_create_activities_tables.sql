-- 创建活动系统相关表

-- 活动主表
CREATE TABLE IF NOT EXISTS activities (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    icon VARCHAR(255) DEFAULT '',
    type VARCHAR(50) NOT NULL CHECK (type IN ('check_in', 'lottery', 'redeem', 'task', 'newbie', 'limited_time')),
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'paused', 'ended', 'archived')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    starts_at TIMESTAMP WITH TIME ZONE,
    ends_at TIMESTAMP WITH TIME ZONE,
    visibility_rules JSONB,
    participation_config JSONB,
    activity_config JSONB,
    total_participations BIGINT NOT NULL DEFAULT 0,
    total_rewards_distributed BIGINT NOT NULL DEFAULT 0,
    created_by VARCHAR(255) NOT NULL DEFAULT 'system',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_activities_type_status ON activities(type, status);
CREATE INDEX IF NOT EXISTS idx_activities_status_sort ON activities(status, sort_order DESC);
CREATE INDEX IF NOT EXISTS idx_activities_time_range ON activities(starts_at, ends_at);

COMMENT ON TABLE activities IS '活动主表';
COMMENT ON COLUMN activities.type IS '活动类型: check_in=签到, lottery=抽奖, redeem=兑换, task=任务, newbie=新手礼包, limited_time=限时活动';
COMMENT ON COLUMN activities.status IS '活动状态';
COMMENT ON COLUMN activities.visibility_rules IS '可见性规则 JSON';
COMMENT ON COLUMN activities.participation_config IS '参与配置 JSON';
COMMENT ON COLUMN activities.activity_config IS '活动特定配置 JSON';

-- 活动奖励表
CREATE TABLE IF NOT EXISTS activity_rewards (
    id BIGSERIAL PRIMARY KEY,
    activity_id BIGINT NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    icon VARCHAR(255) DEFAULT '',
    reward_type VARCHAR(50) NOT NULL CHECK (reward_type IN ('balance', 'subscription', 'coupon', 'points', 'custom')),
    reward_value TEXT DEFAULT '',
    weight INTEGER NOT NULL DEFAULT 100,
    probability DOUBLE PRECISION,
    total_stock BIGINT NOT NULL DEFAULT 0,
    remaining_stock BIGINT NOT NULL DEFAULT 0,
    tier VARCHAR(50) NOT NULL DEFAULT 'common' CHECK (tier IN ('grand', 'first', 'second', 'third', 'common', 'consolation')),
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'out_of_stock')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    distributed_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_activity_rewards_activity_status ON activity_rewards(activity_id, status);
CREATE INDEX IF NOT EXISTS idx_activity_rewards_activity_sort ON activity_rewards(activity_id, sort_order);

COMMENT ON TABLE activity_rewards IS '活动奖励配置表';
COMMENT ON COLUMN activity_rewards.reward_type IS '奖励类型: balance=余额, subscription=订阅, coupon=优惠券, points=积分';
COMMENT ON COLUMN activity_rewards.weight IS '抽奖权重';
COMMENT ON COLUMN activity_rewards.tier IS '奖励等级';

-- 活动参与记录表
CREATE TABLE IF NOT EXISTS activity_participations (
    id BIGSERIAL PRIMARY KEY,
    activity_id BIGINT NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    participated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    daily_window TIMESTAMP WITH TIME ZONE NOT NULL,
    weekly_window TIMESTAMP WITH TIME ZONE NOT NULL,
    monthly_window TIMESTAMP WITH TIME ZONE NOT NULL,
    result VARCHAR(50) NOT NULL DEFAULT 'success' CHECK (result IN ('success', 'failed', 'pending')),
    rewards_received JSONB,
    reward_id BIGINT REFERENCES activity_rewards(id) ON DELETE SET NULL,
    cost_balance DECIMAL(20,8) NOT NULL DEFAULT 0,
    extra_data JSONB,
    ip_address VARCHAR(255) DEFAULT '',
    user_agent TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_activity_participations_user_activity_time ON activity_participations(user_id, activity_id, participated_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_participations_activity_time ON activity_participations(activity_id, participated_at DESC);
CREATE INDEX IF NOT EXISTS idx_activity_participations_user_daily ON activity_participations(user_id, activity_id, daily_window);
CREATE INDEX IF NOT EXISTS idx_activity_participations_user_weekly ON activity_participations(user_id, activity_id, weekly_window);
CREATE INDEX IF NOT EXISTS idx_activity_participations_user_monthly ON activity_participations(user_id, activity_id, monthly_window);
CREATE INDEX IF NOT EXISTS idx_activity_participations_ip_time ON activity_participations(ip_address, participated_at DESC);

COMMENT ON TABLE activity_participations IS '活动参与记录表';
COMMENT ON COLUMN activity_participations.daily_window IS '所属日窗口（当天0点UTC）';
COMMENT ON COLUMN activity_participations.weekly_window IS '所属周窗口（当周一0点UTC）';
COMMENT ON COLUMN activity_participations.monthly_window IS '所属月窗口（当月1号0点UTC）';
COMMENT ON COLUMN activity_participations.rewards_received IS '获得的奖励列表 JSON';
