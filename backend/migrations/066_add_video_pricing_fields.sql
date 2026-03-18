-- Add general video generation pricing fields for OpenAI-compatible platforms (e.g., Grok)
-- +goose Up
ALTER TABLE groups
ADD COLUMN IF NOT EXISTS video_price_per_request DECIMAL(20,8),
ADD COLUMN IF NOT EXISTS video_price_per_request_hd DECIMAL(20,8);

COMMENT ON COLUMN groups.video_price_per_request IS '视频生成单次请求价格（标准质量）';
COMMENT ON COLUMN groups.video_price_per_request_hd IS '视频生成单次请求价格（高清质量）';

-- +goose Down
ALTER TABLE groups
DROP COLUMN IF EXISTS video_price_per_request,
DROP COLUMN IF EXISTS video_price_per_request_hd;
