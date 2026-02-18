package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type securityChatRepository struct {
	db *sql.DB
}

func NewSecurityChatRepository(db *sql.DB) service.SecurityChatRepository {
	return &securityChatRepository{db: db}
}

func (r *securityChatRepository) UpsertSession(ctx context.Context, input *service.SecurityChatSessionUpsertInput) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil security chat repository")
	}
	if input == nil {
		return fmt.Errorf("nil security chat session input")
	}
	if strings.TrimSpace(input.SessionID) == "" {
		return fmt.Errorf("empty session_id")
	}
	firstAt := input.FirstAt
	if firstAt.IsZero() {
		firstAt = time.Now().UTC()
	}
	lastAt := input.LastAt
	if lastAt.IsZero() {
		lastAt = firstAt
	}

	query := `
INSERT INTO security_chat_sessions (
    session_id,
    user_id,
    api_key_id,
    account_id,
    group_id,
    platform,
    model,
    message_preview,
    first_at,
    last_at,
    expires_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (session_id) DO UPDATE SET
    user_id = COALESCE(EXCLUDED.user_id, security_chat_sessions.user_id),
    api_key_id = COALESCE(EXCLUDED.api_key_id, security_chat_sessions.api_key_id),
    account_id = COALESCE(EXCLUDED.account_id, security_chat_sessions.account_id),
    group_id = COALESCE(EXCLUDED.group_id, security_chat_sessions.group_id),
    platform = COALESCE(NULLIF(EXCLUDED.platform, ''), security_chat_sessions.platform),
    model = COALESCE(NULLIF(EXCLUDED.model, ''), security_chat_sessions.model),
    message_preview = COALESCE(NULLIF(EXCLUDED.message_preview, ''), security_chat_sessions.message_preview),
    last_at = GREATEST(security_chat_sessions.last_at, EXCLUDED.last_at),
    expires_at = GREATEST(security_chat_sessions.expires_at, EXCLUDED.expires_at);
`

	_, err := r.db.ExecContext(
		ctx,
		query,
		strings.TrimSpace(input.SessionID),
		opsNullInt64(input.UserID),
		opsNullInt64(input.APIKeyID),
		opsNullInt64(input.AccountID),
		opsNullInt64(input.GroupID),
		opsNullString(input.Platform),
		opsNullString(input.Model),
		opsNullString(input.MessagePreview),
		firstAt,
		lastAt,
		input.ExpiresAt,
	)
	return err
}

func (r *securityChatRepository) InsertChatLog(ctx context.Context, input *service.SecurityChatLogInput) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("nil security chat repository")
	}
	if input == nil {
		return 0, fmt.Errorf("nil security chat input")
	}
	if strings.TrimSpace(input.SessionID) == "" {
		return 0, fmt.Errorf("empty session_id")
	}

	messagesJSON, err := json.Marshal(input.Messages)
	if err != nil {
		return 0, err
	}

	query := `
INSERT INTO security_chat_logs (
    session_id,
    request_id,
    client_request_id,
    user_id,
    api_key_id,
    account_id,
    group_id,
    platform,
    model,
    request_path,
    stream,
    status_code,
    messages,
    message_preview,
    created_at,
    expires_at
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
RETURNING id;
`

	var id int64
	err = r.db.QueryRowContext(
		ctx,
		query,
		strings.TrimSpace(input.SessionID),
		opsNullString(input.RequestID),
		opsNullString(input.ClientRequestID),
		opsNullInt64(input.UserID),
		opsNullInt64(input.APIKeyID),
		opsNullInt64(input.AccountID),
		opsNullInt64(input.GroupID),
		opsNullString(input.Platform),
		opsNullString(input.Model),
		opsNullString(input.RequestPath),
		input.Stream,
		opsNullInt(input.StatusCode),
		string(messagesJSON),
		opsNullString(input.MessagePreview),
		input.CreatedAt,
		input.ExpiresAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

type securityChatSessionRow struct {
	SessionID      string
	UserID         sql.NullInt64
	UserEmail      sql.NullString
	APIKeyID       sql.NullInt64
	AccountID      sql.NullInt64
	GroupID        sql.NullInt64
	Platform       sql.NullString
	Model          sql.NullString
	MessagePreview sql.NullString
	LastAt         time.Time
	RequestCount   int64
}

func (r *securityChatRepository) ListSessions(ctx context.Context, filter *service.SecurityChatSessionFilter) ([]*service.SecurityChatSession, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("nil security chat repository")
	}

	page, pageSize, startTime, endTime := filter.Normalize()
	offset := (page - 1) * pageSize

	conditions := make([]string, 0, 8)
	args := make([]any, 0, 12)
	args = append(args, startTime.UTC(), endTime.UTC())

	addCondition := func(condition string, values ...any) {
		conditions = append(conditions, condition)
		args = append(args, values...)
	}

	if filter != nil {
		if filter.UserID != nil && *filter.UserID > 0 {
			addCondition(fmt.Sprintf("user_id = $%d", len(args)+1), *filter.UserID)
		}
		if filter.APIKeyID != nil && *filter.APIKeyID > 0 {
			addCondition(fmt.Sprintf("api_key_id = $%d", len(args)+1), *filter.APIKeyID)
		}
		if filter.AccountID != nil && *filter.AccountID > 0 {
			addCondition(fmt.Sprintf("account_id = $%d", len(args)+1), *filter.AccountID)
		}
		if filter.GroupID != nil && *filter.GroupID > 0 {
			addCondition(fmt.Sprintf("group_id = $%d", len(args)+1), *filter.GroupID)
		}
		if s := strings.TrimSpace(filter.SessionID); s != "" {
			addCondition(fmt.Sprintf("session_id = $%d", len(args)+1), s)
		}
		if p := strings.TrimSpace(filter.Platform); p != "" {
			addCondition(fmt.Sprintf("platform = $%d", len(args)+1), strings.ToLower(p))
		}
		if m := strings.TrimSpace(filter.Model); m != "" {
			addCondition(fmt.Sprintf("model = $%d", len(args)+1), m)
		}
		if q := strings.TrimSpace(filter.Query); q != "" {
			like := "%" + strings.ToLower(q) + "%"
			startIdx := len(args) + 1
			addCondition(
				fmt.Sprintf("(LOWER(COALESCE(session_id,'')) LIKE $%d OR LOWER(COALESCE(message_preview,'')) LIKE $%d)", startIdx, startIdx+1),
				like, like,
			)
		}
	}

	where := ""
	if len(conditions) > 0 {
		where = "AND " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`
SELECT COUNT(1)
FROM security_chat_sessions
WHERE last_at >= $1 AND last_at < $2 %s;
`, where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			total = 0
		} else {
			return nil, 0, err
		}
	}

	listQuery := fmt.Sprintf(`
SELECT
  s.session_id,
  s.user_id,
  u.email,
  s.api_key_id,
  s.account_id,
  s.group_id,
  s.platform,
  s.model,
  s.message_preview,
  s.last_at,
  (SELECT COUNT(1)
     FROM security_chat_logs l
    WHERE l.session_id = s.session_id
      AND l.user_id IS NOT DISTINCT FROM s.user_id
      AND l.api_key_id IS NOT DISTINCT FROM s.api_key_id) AS request_count
FROM security_chat_sessions s
LEFT JOIN users u ON u.id = s.user_id
WHERE s.last_at >= $1 AND s.last_at < $2 %s
ORDER BY s.last_at DESC
LIMIT $%d OFFSET $%d;
`, where, len(args)+1, len(args)+2)

	listArgs := append(append([]any{}, args...), pageSize, offset)
	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]*service.SecurityChatSession, 0)
	for rows.Next() {
		var row securityChatSessionRow
		if err := rows.Scan(
			&row.SessionID,
			&row.UserID,
			&row.UserEmail,
			&row.APIKeyID,
			&row.AccountID,
			&row.GroupID,
			&row.Platform,
			&row.Model,
			&row.MessagePreview,
			&row.LastAt,
			&row.RequestCount,
		); err != nil {
			return nil, 0, err
		}

		items = append(items, service.SecurityChatSessionFromRow(
			row.SessionID,
			row.UserID,
			row.APIKeyID,
			row.AccountID,
			row.GroupID,
			row.UserEmail,
			row.Platform,
			row.Model,
			row.MessagePreview,
			row.LastAt,
			row.RequestCount,
		))
	}

	return items, total, nil
}

func buildSecurityChatLogConditions(filter *service.SecurityChatMessageFilter, startTime, endTime time.Time, tableAlias string) (string, []any) {
	if filter == nil {
		filter = &service.SecurityChatMessageFilter{}
	}

	prefix := ""
	if strings.TrimSpace(tableAlias) != "" {
		prefix = strings.TrimSpace(tableAlias) + "."
	}

	conditions := make([]string, 0, 12)
	args := make([]any, 0, 12)

	addCondition := func(condition string, values ...any) {
		conditions = append(conditions, condition)
		args = append(args, values...)
	}

	if !filter.IgnoreTimeRange {
		args = append(args, startTime.UTC(), endTime.UTC())
		conditions = append(conditions, prefix+"created_at >= $1", prefix+"created_at < $2")
	}

	if s := strings.TrimSpace(filter.SessionID); s != "" {
		addCondition(fmt.Sprintf("%ssession_id = $%d", prefix, len(args)+1), s)
	}
	if filter.UserID != nil && *filter.UserID > 0 {
		addCondition(fmt.Sprintf("%suser_id = $%d", prefix, len(args)+1), *filter.UserID)
	}
	if filter.APIKeyID != nil && *filter.APIKeyID > 0 {
		addCondition(fmt.Sprintf("%sapi_key_id = $%d", prefix, len(args)+1), *filter.APIKeyID)
	}
	if filter.AccountID != nil && *filter.AccountID > 0 {
		addCondition(fmt.Sprintf("%saccount_id = $%d", prefix, len(args)+1), *filter.AccountID)
	}
	if filter.GroupID != nil && *filter.GroupID > 0 {
		addCondition(fmt.Sprintf("%sgroup_id = $%d", prefix, len(args)+1), *filter.GroupID)
	}
	if p := strings.TrimSpace(filter.Platform); p != "" {
		addCondition(fmt.Sprintf("LOWER(COALESCE(%splatform,'')) = $%d", prefix, len(args)+1), strings.ToLower(p))
	}
	if m := strings.TrimSpace(filter.Model); m != "" {
		addCondition(fmt.Sprintf("%smodel = $%d", prefix, len(args)+1), m)
	}
	if rp := strings.TrimSpace(filter.RequestPath); rp != "" {
		addCondition(fmt.Sprintf("%srequest_path = $%d", prefix, len(args)+1), rp)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}
	return where, args
}

func (r *securityChatRepository) ListMessages(ctx context.Context, filter *service.SecurityChatMessageFilter) ([]*service.SecurityChatLog, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("nil security chat repository")
	}
	if filter == nil || (strings.TrimSpace(filter.SessionID) == "" && !filter.AllowEmptySession) {
		return []*service.SecurityChatLog{}, 0, nil
	}

	page, pageSize, startTime, endTime := filter.Normalize()
	offset := (page - 1) * pageSize

	where, args := buildSecurityChatLogConditions(filter, startTime, endTime, "l")

	countQuery := fmt.Sprintf(`SELECT COUNT(1) FROM security_chat_logs l %s`, where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			total = 0
		} else {
			return nil, 0, err
		}
	}

	listQuery := fmt.Sprintf(`
 SELECT
  l.id,
  l.session_id,
  l.request_id,
  l.client_request_id,
  l.user_id,
  u.email,
  l.api_key_id,
  l.account_id,
  l.group_id,
  l.platform,
  l.model,
  l.request_path,
  l.stream,
  l.status_code,
  l.messages::TEXT,
  l.created_at
FROM security_chat_logs l
LEFT JOIN users u ON u.id = l.user_id
%s
ORDER BY l.created_at ASC
LIMIT $%d OFFSET $%d;
`, where, len(args)+1, len(args)+2)

	listArgs := append(append([]any{}, args...), pageSize, offset)
	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]*service.SecurityChatLog, 0)
	for rows.Next() {
		var (
			row             service.SecurityChatLog
			messagesRaw     string
			userID          sql.NullInt64
			userEmail       sql.NullString
			apiKeyID        sql.NullInt64
			accountID       sql.NullInt64
			groupID         sql.NullInt64
			platform        sql.NullString
			model           sql.NullString
			requestID       sql.NullString
			clientRequestID sql.NullString
			requestPath     sql.NullString
			statusCode      sql.NullInt64
		)
		if err := rows.Scan(
			&row.ID,
			&row.SessionID,
			&requestID,
			&clientRequestID,
			&userID,
			&userEmail,
			&apiKeyID,
			&accountID,
			&groupID,
			&platform,
			&model,
			&requestPath,
			&row.Stream,
			&statusCode,
			&messagesRaw,
			&row.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		if requestID.Valid {
			row.RequestID = &requestID.String
		}
		if clientRequestID.Valid {
			row.ClientRequestID = &clientRequestID.String
		}
		if userID.Valid {
			v := userID.Int64
			row.UserID = &v
		}
		if userEmail.Valid {
			row.UserEmail = &userEmail.String
		}
		if apiKeyID.Valid {
			v := apiKeyID.Int64
			row.APIKeyID = &v
		}
		if accountID.Valid {
			v := accountID.Int64
			row.AccountID = &v
		}
		if groupID.Valid {
			v := groupID.Int64
			row.GroupID = &v
		}
		if platform.Valid {
			row.Platform = &platform.String
		}
		if model.Valid {
			row.Model = &model.String
		}
		if requestPath.Valid {
			row.RequestPath = &requestPath.String
		}
		if statusCode.Valid {
			v := int(statusCode.Int64)
			row.StatusCode = &v
		}

		var messages []service.SecurityChatMessage
		if err := json.Unmarshal([]byte(messagesRaw), &messages); err == nil {
			row.Messages = messages
		}

		items = append(items, &row)
	}

	return items, total, nil
}

func (r *securityChatRepository) GetStats(ctx context.Context, filter *service.SecurityChatMessageFilter) (*service.SecurityChatStats, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil security chat repository")
	}
	if filter == nil {
		filter = &service.SecurityChatMessageFilter{}
	}

	_, _, startTime, endTime := filter.Normalize()
	where, args := buildSecurityChatLogConditions(filter, startTime, endTime, "l")

	var requestCount int64
	var sessionCount int64
	var estimatedBytes int64
	countQuery := fmt.Sprintf(`
SELECT
  COUNT(1),
  COUNT(DISTINCT session_id),
  COALESCE(SUM(pg_column_size(messages)), 0)
FROM security_chat_logs l
%s;
`, where)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&requestCount, &sessionCount, &estimatedBytes); err != nil {
		if err == sql.ErrNoRows {
			requestCount = 0
			sessionCount = 0
			estimatedBytes = 0
		} else {
			return nil, err
		}
	}

	var tableBytes int64
	if err := r.db.QueryRowContext(ctx, `SELECT pg_total_relation_size('security_chat_logs')`).Scan(&tableBytes); err != nil {
		return nil, err
	}

	platformBuckets := make([]service.SecurityChatPlatformBucket, 0, 3)
	platformQuery := fmt.Sprintf(`
SELECT
  CASE
    WHEN LOWER(COALESCE(l.platform,'')) LIKE '%%opencode%%' THEN 'opencode'
    WHEN LOWER(COALESCE(l.platform,'')) LIKE '%%codex%%' THEN 'codex'
    ELSE 'other'
  END AS bucket,
  COUNT(1) AS count
FROM security_chat_logs l
%s
GROUP BY bucket;
`, where)
	rows, err := r.db.QueryContext(ctx, platformQuery, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var bucket string
		var count int64
		if err := rows.Scan(&bucket, &count); err != nil {
			return nil, err
		}
		platformBuckets = append(platformBuckets, service.SecurityChatPlatformBucket{Key: bucket, Count: count})
	}

	return &service.SecurityChatStats{
		RequestCount:    requestCount,
		SessionCount:    sessionCount,
		EstimatedBytes:  estimatedBytes,
		TableBytes:      tableBytes,
		PlatformBuckets: platformBuckets,
	}, nil
}

func (r *securityChatRepository) DeleteExpired(ctx context.Context, cutoff time.Time) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("nil security chat repository")
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM security_chat_logs WHERE expires_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *securityChatRepository) DeleteExpiredSessions(ctx context.Context, cutoff time.Time) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("nil security chat repository")
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM security_chat_sessions WHERE expires_at < $1`, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *securityChatRepository) DeleteSession(ctx context.Context, sessionID string, userID *int64, apiKeyID *int64) (int64, int64, error) {
	if r == nil || r.db == nil {
		return 0, 0, fmt.Errorf("nil security chat repository")
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return 0, 0, fmt.Errorf("session_id required")
	}

	conditions := []string{"session_id = $1"}
	args := []any{sessionID}

	if userID != nil && *userID > 0 {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)+1))
		args = append(args, *userID)
	}
	if apiKeyID != nil && *apiKeyID > 0 {
		conditions = append(conditions, fmt.Sprintf("api_key_id = $%d", len(args)+1))
		args = append(args, *apiKeyID)
	}
	where := strings.Join(conditions, " AND ")

	deleteLogsQuery := fmt.Sprintf("DELETE FROM security_chat_logs WHERE %s", where)
	resLogs, err := r.db.ExecContext(ctx, deleteLogsQuery, args...)
	if err != nil {
		return 0, 0, err
	}
	logsDeleted, _ := resLogs.RowsAffected()

	deleteSessionsQuery := fmt.Sprintf("DELETE FROM security_chat_sessions WHERE %s", where)
	resSessions, err := r.db.ExecContext(ctx, deleteSessionsQuery, args...)
	if err != nil {
		return logsDeleted, 0, err
	}
	sessionsDeleted, _ := resSessions.RowsAffected()

	return logsDeleted, sessionsDeleted, nil
}

func (r *securityChatRepository) DeleteSessions(ctx context.Context, sessionIDs []string, userID *int64, apiKeyID *int64) (int64, int64, error) {
	if r == nil || r.db == nil {
		return 0, 0, fmt.Errorf("nil security chat repository")
	}
	if len(sessionIDs) == 0 {
		return 0, 0, fmt.Errorf("session_ids required")
	}

	conditions := []string{"session_id = ANY($1)"}
	args := []any{pq.Array(sessionIDs)}

	if userID != nil && *userID > 0 {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)+1))
		args = append(args, *userID)
	}
	if apiKeyID != nil && *apiKeyID > 0 {
		conditions = append(conditions, fmt.Sprintf("api_key_id = $%d", len(args)+1))
		args = append(args, *apiKeyID)
	}
	where := strings.Join(conditions, " AND ")

	deleteLogsQuery := fmt.Sprintf("DELETE FROM security_chat_logs WHERE %s", where)
	resLogs, err := r.db.ExecContext(ctx, deleteLogsQuery, args...)
	if err != nil {
		return 0, 0, err
	}
	logsDeleted, _ := resLogs.RowsAffected()

	deleteSessionsQuery := fmt.Sprintf("DELETE FROM security_chat_sessions WHERE %s", where)
	resSessions, err := r.db.ExecContext(ctx, deleteSessionsQuery, args...)
	if err != nil {
		return logsDeleted, 0, err
	}
	sessionsDeleted, _ := resSessions.RowsAffected()

	return logsDeleted, sessionsDeleted, nil
}

func (r *securityChatRepository) DeleteSessionsByFilter(ctx context.Context, filter *service.SecurityChatSessionFilter) (int64, int64, error) {
	if r == nil || r.db == nil {
		return 0, 0, fmt.Errorf("nil security chat repository")
	}
	if filter == nil {
		return 0, 0, fmt.Errorf("filter required")
	}

	_, _, startTime, endTime := filter.Normalize()
	conditions := make([]string, 0, 8)
	args := make([]any, 0, 10)
	args = append(args, startTime.UTC(), endTime.UTC())

	addCondition := func(condition string, values ...any) {
		conditions = append(conditions, condition)
		args = append(args, values...)
	}

	conditions = append(conditions, "last_at >= $1", "last_at < $2")
	if filter.UserID != nil && *filter.UserID > 0 {
		addCondition(fmt.Sprintf("user_id = $%d", len(args)+1), *filter.UserID)
	}
	if filter.APIKeyID != nil && *filter.APIKeyID > 0 {
		addCondition(fmt.Sprintf("api_key_id = $%d", len(args)+1), *filter.APIKeyID)
	}
	if filter.AccountID != nil && *filter.AccountID > 0 {
		addCondition(fmt.Sprintf("account_id = $%d", len(args)+1), *filter.AccountID)
	}
	if filter.GroupID != nil && *filter.GroupID > 0 {
		addCondition(fmt.Sprintf("group_id = $%d", len(args)+1), *filter.GroupID)
	}
	if s := strings.TrimSpace(filter.SessionID); s != "" {
		addCondition(fmt.Sprintf("session_id = $%d", len(args)+1), s)
	}
	if p := strings.TrimSpace(filter.Platform); p != "" {
		addCondition(fmt.Sprintf("platform = $%d", len(args)+1), strings.ToLower(p))
	}
	if m := strings.TrimSpace(filter.Model); m != "" {
		addCondition(fmt.Sprintf("model = $%d", len(args)+1), m)
	}
	if q := strings.TrimSpace(filter.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		startIdx := len(args) + 1
		addCondition(
			fmt.Sprintf("(LOWER(COALESCE(session_id,'')) LIKE $%d OR LOWER(COALESCE(message_preview,'')) LIKE $%d)", startIdx, startIdx+1),
			like, like,
		)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	deleteLogsQuery := fmt.Sprintf("DELETE FROM security_chat_logs WHERE session_id IN (SELECT session_id FROM security_chat_sessions %s)", where)
	resLogs, err := r.db.ExecContext(ctx, deleteLogsQuery, args...)
	if err != nil {
		return 0, 0, err
	}
	logsDeleted, _ := resLogs.RowsAffected()

	deleteSessionsQuery := fmt.Sprintf("DELETE FROM security_chat_sessions %s", where)
	resSessions, err := r.db.ExecContext(ctx, deleteSessionsQuery, args...)
	if err != nil {
		return logsDeleted, 0, err
	}
	sessionsDeleted, _ := resSessions.RowsAffected()

	return logsDeleted, sessionsDeleted, nil
}

func (r *securityChatRepository) DeleteLogsByFilter(ctx context.Context, filter *service.SecurityChatMessageFilter) (int64, int64, error) {
	if r == nil || r.db == nil {
		return 0, 0, fmt.Errorf("nil security chat repository")
	}
	if filter == nil {
		return 0, 0, fmt.Errorf("filter required")
	}

	_, _, startTime, endTime := filter.Normalize()
	where, args := buildSecurityChatLogConditions(filter, startTime, endTime, "")

	deleteLogsQuery := fmt.Sprintf("DELETE FROM security_chat_logs %s", where)
	resLogs, err := r.db.ExecContext(ctx, deleteLogsQuery, args...)
	if err != nil {
		return 0, 0, err
	}
	logsDeleted, _ := resLogs.RowsAffected()

	if logsDeleted == 0 {
		return 0, 0, nil
	}

	resSessions, err := r.db.ExecContext(ctx, `
DELETE FROM security_chat_sessions s
WHERE NOT EXISTS (
  SELECT 1 FROM security_chat_logs l WHERE l.session_id = s.session_id
);
`)
	if err != nil {
		return logsDeleted, 0, err
	}
	sessionsDeleted, _ := resSessions.RowsAffected()

	return logsDeleted, sessionsDeleted, nil
}
