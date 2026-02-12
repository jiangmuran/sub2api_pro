-- Add disk usage metrics to ops_system_metrics

ALTER TABLE ops_system_metrics
  ADD COLUMN IF NOT EXISTS disk_used_mb BIGINT,
  ADD COLUMN IF NOT EXISTS disk_total_mb BIGINT,
  ADD COLUMN IF NOT EXISTS disk_usage_percent DOUBLE PRECISION;

COMMENT ON COLUMN ops_system_metrics.disk_used_mb IS 'Disk used MB for primary volume.';
COMMENT ON COLUMN ops_system_metrics.disk_total_mb IS 'Disk total MB for primary volume.';
COMMENT ON COLUMN ops_system_metrics.disk_usage_percent IS 'Disk usage percent for primary volume.';
