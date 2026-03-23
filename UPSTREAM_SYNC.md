# Upstream Sync Tracking

> **Strategy**: Selective cherry-pick  
> **Upstream**: https://github.com/Wei-Shaw/sub2api  
> **Fork**: https://github.com/jiangmuran/sub2api_pro

## Why "517 commits behind"?

We use **cherry-pick** instead of direct merge, which creates new commit hashes. Git doesn't recognize cherry-picked commits as "same" as upstream, so the counter won't decrease. This is **intentional** and gives us full control over what features to adopt.

---

## ✅ Successfully Merged (Cherry-picked)

### 2026-03-23 - Anthropic Messages API & Claude Code Compatibility

| Upstream Hash | Our Hash | Description | PR/Issue |
|---------------|----------|-------------|----------|
| `ff1f1149` | `76e1ae88` | feat(openai): add /v1/messages endpoint and API compatibility layer | [#809](https://github.com/Wei-Shaw/sub2api/pull/809) |
| `bc194a7d` | `19be4fef` | fix: address PR review - Anthropic error format in panic recovery | [#809](https://github.com/Wei-Shaw/sub2api/pull/809) |
| `92159994` | `1c8c2478` | feat: /v1/messages端点适配codex账号池 | - |
| `7d26b810` | `99629493` | fix: address review - add missing whitespace patterns | - |
| `af96c8ea` | `8f4883e1` | feat: map claude-haiku-4-5 variants to claude-sonnet-4-6 | - |
| `a14babdc` | `7aed7238` | fix: 兼容 Claude Code v2.1.78+ 新 JSON 格式 metadata.user_id | - |

**Files Changed**: 
- New: `backend/internal/pkg/apicompat/` (complete package)
- New: `backend/internal/service/openai_gateway_messages.go`
- New: `backend/internal/service/metadata_userid.go`
- New: `backend/internal/service/metadata_userid_test.go`
- Modified: 15+ files for /v1/messages support

**Conflicts Resolved**:
- `CreateAccountModal.vue`: Merged OAuth model mapping UI (preserved nano-banana)
- `openai_gateway_service.go`: Merged BillingModel logic
- `identity_service.go`: Removed old userIDRegex, adopted ParseMetadataUserID

**Tests**: ✅ All passed (183 new unit tests added)

**Production Deploy**: ✅ 2026-03-23 14:50 CST

---

## 🔍 Analyzed but NOT Merged

### Skipped (Dependency Conflicts)
- `bd9d2671`: go mod tidy - conflicts with current dependency versions

### Recommended for Future (Needs Testing)
- None currently

### Large Features (Not Recommended Yet)
- `bf3d6c0e`: 529 overload cooldown UI config (634 lines)
- Upstream model tracking series (database schema changes required)

---

## 📊 Sync Status

**Last Sync**: 2026-03-23  
**Upstream Position**: `bda7c39e` (latest as of fetch)  
**Our Position**: `7aed7238`  

**Divergence**:
- We are **249 commits ahead** (custom features: nano-banana, voice chat, etc.)
- We are **517 commits behind** (many are duplicates via cherry-pick, others are features we don't need)

---

## 📝 Cherry-pick Checklist

When considering a new upstream commit:

1. ✅ **Check**: Does it fix a critical bug?
2. ✅ **Check**: Does it add a feature we need?
3. ✅ **Check**: Does it conflict with our custom code (nano-banana, etc.)?
4. ✅ **Test**: Can we build backend + frontend after merge?
5. ✅ **Test**: Do existing tests pass?
6. ✅ **Deploy**: Test in production before marking complete

---

## 🎯 Next Candidates to Review

Run this command to see next 50 upstream commits:

```bash
cd /Users/jmr/projects/sub2api_pro
git log --oneline --reverse main..upstream/main | head -50
```

### High Priority (Bug Fixes)
- `72961c58`: fix: Anthropic 平台无限流重置时间的 429 不再误标记账号限流
- `8e1bcf53`: fix: extend RewriteUserID regex to match user_id containing account_uuid
- `078fefed`: fix: 修复账号管理页面容量列显示为0的bug
- `9d70c385`: fix: 修复claude apikey账号请求时未携带beta=true 查询参数的bug

### Medium Priority (Features)
- `ba6de4c4`: feat: /keys页面支持表单筛选
- `80ae592c`: perf(admin): optimize large-dataset loading for dashboard/users/accounts/ops
- `05b1c66a`: perf(admin-usage): avoid expensive count on large usage_logs pagination
- `bd0801a8`: feat(registration): add email domain whitelist policy

### Low Priority (Nice to Have)
- `d4f6ad72`: feat: 新增apikey的usage查询页面
- `3a089242`: feat: 支持基于 crontab 的定时账号测试

---

## 🚀 How to Cherry-pick a Commit

```bash
# 1. Fetch latest upstream
git fetch upstream

# 2. Cherry-pick the commit
git cherry-pick <upstream-commit-hash>

# 3. If conflicts, resolve them
# ... edit files ...
git add .
git cherry-pick --continue

# 4. Test build
cd backend && make build
cd ../frontend && pnpm typecheck

# 5. Push to origin
git push origin main

# 6. Update this file with the merge info
```

---

## 📌 Notes

- **DO NOT** try to merge all 517 commits at once - that would create massive conflicts
- **DO** cherry-pick selectively based on our needs
- **DO** keep this file updated as we merge more commits
- **DO** test thoroughly before deploying to production

---

Last Updated: 2026-03-23 by Assistant
