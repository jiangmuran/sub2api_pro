# Grok 图片编辑和视频生成 API 文档

本项目已完整支持 Grok 系列模型的图片编辑和视频生成功能，完全兼容 OpenAI API 格式。

## 🎯 核心特性

- ✅ **图片编辑 API** (`/v1/images/edits`) - 支持上传图片并使用 AI 进行编辑
- ✅ **视频生成 API** (`/v1/videos`) - 支持文本生成视频，6-30 秒自动扩展
- ✅ **单次计费** - 支持按质量（标准/高清）分别定价
- ✅ **透传架构** - 请求直接透传到支持相同接口的上游 API
- ✅ **定价管理** - 支持分组级和账户级定价配置
- ✅ **在线测试** - ModelTestView 已集成图片编辑和视频生成测试界面
- ✅ **前端 API** - 提供完整的 TypeScript composables

## 🎯 多轮视频生成支持

### 架构说明

sub2api 采用**透传架构**，将视频生成请求转发到上游 [grok2api](https://github.com/jiangmuran/grok2api) 服务。grok2api 已完美实现多轮视频生成逻辑：

1. **自动多轮规划**：根据目标时长（6-30秒）自动规划生成轮次
2. **链式扩展**：每轮6秒，自动传递 `post_id` 和 `extendPostId` 参数
3. **流式处理**：实时处理进度更新、错误恢复等
4. **格式转换**：sub2api 自动将 Grok 响应格式转换为 OpenAI 标准格式

### 配置说明

在 sub2api 账户管理中配置：

```json
{
  "base_url": "https://your-grok2api-domain.com",
  "api_key": "your-grok2api-token",
  "model_mapping": {
    "grok-imagine-1.0-video": "grok-imagine-1.0-video",
    "grok-imagine-1.0-edit": "grok-imagine-1.0-edit"
  }
}
```

### 视频时长

- ✅ 支持 6-30 秒
- ✅ 自动多轮扩展（每轮6秒）
- ✅ 12秒 = 2轮，18秒 = 3轮，以此类推

## 📡 API 端点

### 1. 图片编辑 API

**端点**: `POST /v1/images/edits`

**请求格式**: `multipart/form-data`

**参数**:

| 字段 | 类型 | 必填 | 说明 | 可选值 |
|------|------|------|------|--------|
| `model` | string | ✅ | 图像编辑模型 | `grok-imagine-1.0-edit` |
| `prompt` | string | ✅ | 编辑描述 | 任意文本 |
| `image` | file | ✅ | 待编辑图片 | PNG, JPG, WEBP |
| `n` | integer | ❌ | 生成数量 | 1-10（流式模式限 1-2） |
| `stream` | boolean | ❌ | 流式输出 | `true`, `false` |
| `size` | string | ❌ | 图片尺寸 | `1280x720`, `720x1280`, `1792x1024`, `1024x1792`, `1024x1024` |
| `response_format` | string | ❌ | 响应格式 | `url`, `b64_json`, `base64` |

**示例**:

```bash
curl http://localhost:8000/v1/images/edits \
  -H "Authorization: Bearer $API_KEY" \
  -F "model=grok-imagine-1.0-edit" \
  -F "prompt=提高清晰度并增强色彩" \
  -F "image=@/path/to/image.png" \
  -F "n=1" \
  -F "size=1024x1024" \
  -F "response_format=url"
```

**响应示例**:

```json
{
  "created": 1710756789,
  "data": [
    {
      "url": "https://example.com/edited-image.png"
    }
  ]
}
```

### 2. 视频生成 API

**端点**: `POST /v1/videos`

**请求格式**: `application/json`

**参数**:

| 字段 | 类型 | 必填 | 说明 | 可选值 |
|------|------|------|------|--------|
| `model` | string | ✅ | 视频模型 | `grok-imagine-1.0-video` |
| `prompt` | string | ✅ | 视频提示词 | 任意文本 |
| `size` | string | ❌ | 画面比例 | `1280x720`, `720x1280`, `1792x1024`, `1024x1792`, `1024x1024` |
| `seconds` | integer | ❌ | 视频时长（秒） | 6-30，上游自动处理多轮扩展 |
| `quality` | string | ❌ | 视频质量 | `standard`(480p), `high`(720p) |
| `image_reference` | object | ❌ | 参考图 | `{"image_url": "https://..."}` 或 Data URI |

**示例**:

```bash
curl http://localhost:8000/v1/videos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "grok-imagine-1.0-video",
    "prompt": "霓虹雨夜街头，慢镜头追拍，赛博朋克风格",
    "size": "1792x1024",
    "seconds": 18,
    "quality": "high"
  }'
```

**响应示例** (OpenAI 标准格式):

```json
{
  "created": 1710756789,
  "data": [
    {
      "url": "https://example.com/generated-video.mp4"
    }
  ]
}
```

> **注意**: 上游 Grok API 返回的原始格式如下，系统会自动转换为 OpenAI 标准格式：
> ```json
> {
>   "id": "video_3b4325581c8d447e91ecda61",
>   "object": "video",
>   "created_at": 1773759941,
>   "status": "completed",
>   "model": "grok-imagine-1.0-video",
>   "url": "https://..."
> }
> ```

**带参考图示例**:

```bash
curl http://localhost:8000/v1/videos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "grok-imagine-1.0-video",
    "prompt": "基于参考图生成动态效果",
    "size": "1280x720",
    "seconds": 6,
    "quality": "standard",
    "image_reference": {
      "image_url": "https://example.com/reference.jpg"
    }
  }'
```

## 💰 计费配置

### 1. 数据库字段

已在 `groups` 表添加以下字段：

```sql
ALTER TABLE groups
ADD COLUMN video_price_per_request DECIMAL(20,8),     -- 标准质量单次价格
ADD COLUMN video_price_per_request_hd DECIMAL(20,8);  -- 高清质量单次价格
```

### 2. 默认价格

如果未配置分组价格，系统使用以下默认值：

- **Grok 视频模型**:
  - 标准质量: `$0.10` / 次
  - 高清质量: `$0.20` / 次

- **其他视频模型**:
  - 标准质量: `$0.08` / 次
  - 高清质量: `$0.15` / 次

### 3. 定价优先级

1. **分组级定价** - `Group.VideoPricePerRequest` / `VideoPricePerRequestHD`
2. **账户级手动定价** - `Account.Extra["openai_manual_model_pricing"][model]["video_price_per_request"]`
3. **默认价格** - 代码中的硬编码价格

### 4. 配置示例

**分组级配置**（通过管理后台）:

```json
{
  "id": 1,
  "name": "Premium Group",
  "video_price_per_request": 0.08,
  "video_price_per_request_hd": 0.16
}
```

**账户级手动配置**（写入 `accounts.extra` 字段）:

```json
{
  "openai_manual_model_pricing": {
    "grok-imagine-1.0-video": {
      "video_price_per_request": 0.12,
      "video_price_per_request_hd": 0.24
    }
  }
}
```

## 🖥️ 前端集成

### 1. 在线测试界面

访问 **模型测试** 页面 (`/user/model-test`)，已包含三个测试模块：

1. **图片生成** - 使用 `/v1/images/generations` 接口
2. **图片编辑** - 上传图片并编辑（新增）
3. **视频生成** - 生成 6-30 秒视频（新增）

### 2. 前端 API Composables

#### 图片编辑

```typescript
import { useImageEditAPI } from '@/composables/useImageEditAPI'

const { editImage, loading, error } = useImageEditAPI()

const result = await editImage({
  apiKey: 'your-api-key',
  model: 'grok-imagine-1.0-edit',
  prompt: '提高清晰度',
  image: fileObject,
  n: 1,
  size: '1024x1024',
  response_format: 'url'
})

console.log(result.data) // [{ url: '...', cost: 0.04 }]
```

#### 视频生成

```typescript
import { useVideoAPI } from '@/composables/useVideoAPI'

const { generateVideo, loading, progress, error } = useVideoAPI()

const result = await generateVideo({
  apiKey: 'your-api-key',
  model: 'grok-imagine-1.0-video',
  prompt: '霓虹雨夜街头',
  size: '1280x720',
  seconds: 18,
  quality: 'high'
})

console.log(result.data) // [{ url: '...', cost: 0.20 }]
console.log(progress.value) // '生成中...' | 'Processing video...' | 'Completed'
```

## 🏗️ 架构说明

### 请求流程

```
Client Request
    ↓
API Key 认证
    ↓
分组权限检查
    ↓
Handler (ImageEdits / VideoGenerations)
    ↓
Service (ForwardImageEdits / ForwardVideoGeneration)
    ↓
透传到上游 OpenAI 兼容 API
    ↓
解析响应
    ↓
异步计费 (RecordUsage)
    ↓
返回结果给客户端
```

### 计费流程

```typescript
// 视频计费示例
if (result.VideoCount > 0) {
  videoConfig := &VideoPriceConfig{
    PricePerRequest:   group.VideoPricePerRequest,
    PricePerRequestHD: group.VideoPricePerRequestHD,
  }
  
  // 账户级手动定价覆盖
  if manual := lookupOpenAIAccountStoredVideoPricing(account, model, quality); manual > 0 {
    if quality == "high" {
      videoConfig.PricePerRequestHD = &manual
    } else {
      videoConfig.PricePerRequest = &manual
    }
  }
  
  cost = billingService.CalculateVideoCost(model, quality, videoCount, videoConfig, rateMultiplier)
}
```

## 🔧 开发指南

### 1. 添加新的上游 API

修改 `backend/internal/service/openai_gateway_service.go`:

```go
// 在 ForwardVideoGeneration 中配置上游 URL
targetURL := buildOpenAIEndpointURL("https://api.openai.com", "videos")
if baseURL := strings.TrimSpace(account.GetOpenAIBaseURL()); baseURL != "" {
    validatedURL, validateErr := s.validateUpstreamBaseURL(baseURL)
    if validateErr != nil {
        return nil, validateErr
    }
    targetURL = buildOpenAIEndpointURL(validatedURL, "videos")
}
```

### 2. 修改默认价格

修改 `backend/internal/service/billing_service.go`:

```go
func (s *BillingService) CalculateVideoCost(...) *CostBreakdown {
    // ...
    if unitPrice <= 0 {
        modelLower := strings.ToLower(model)
        if strings.Contains(modelLower, "grok") && strings.Contains(modelLower, "video") {
            if quality == "high" {
                unitPrice = 0.20 // 修改此处
            } else {
                unitPrice = 0.10 // 修改此处
            }
        }
    }
    // ...
}
```

### 3. 数据库迁移

运行迁移添加视频定价字段：

```bash
cd backend
go run cmd/migrate/main.go up
```

或手动执行 SQL:

```sql
-- migrations/066_add_video_pricing_fields.sql
ALTER TABLE groups
ADD COLUMN IF NOT EXISTS video_price_per_request DECIMAL(20,8),
ADD COLUMN IF NOT EXISTS video_price_per_request_hd DECIMAL(20,8);
```

## 📊 监控和日志

所有请求都会记录到 `usage_logs` 表，包含以下信息：

- `model` - 使用的模型名称
- `media_type` - `"video"` 或 `"image"`
- `video_count` - 生成的视频数量（视频请求）
- `video_quality` - 视频质量 (`"standard"` / `"high"`)
- `image_count` - 生成的图片数量（图片编辑请求）
- `image_size` - 图片尺寸
- `cost` - 实际扣费金额
- `duration_ms` - 请求耗时（毫秒）

## 🧪 测试

### 后端测试

```bash
cd backend
go test ./internal/service -run TestCalculateVideoCost -v
go test ./internal/handler -run TestVideoGenerations -v
```

### 前端测试

```bash
cd frontend
pnpm typecheck  # 类型检查
pnpm lint       # 代码检查
pnpm test       # 运行测试
```

### 手动测试

1. **启动服务**:
   ```bash
   make build
   ./backend/bin/server
   ```

2. **测试图片编辑**:
   ```bash
   curl http://localhost:8000/v1/images/edits \
     -H "Authorization: Bearer sk-xxx" \
     -F "model=grok-imagine-1.0-edit" \
     -F "prompt=提高清晰度" \
     -F "image=@test.png"
   ```

3. **测试视频生成**:
   ```bash
   curl http://localhost:8000/v1/videos \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer sk-xxx" \
     -d '{"model":"grok-imagine-1.0-video","prompt":"测试视频","seconds":6}'
   ```

## 🔐 安全注意事项

1. **上游 URL 验证** - 所有上游 URL 都经过严格验证，防止 SSRF 攻击
2. **文件类型检查** - 图片编辑仅接受 PNG/JPG/WEBP 格式
3. **大小限制** - 默认 32MB 上传限制，可通过 `MaxBodySize` 配置
4. **API Key 认证** - 所有请求必须提供有效的 API Key
5. **分组权限** - 检查 API Key 是否属于有效分组

## 📝 常见问题

### Q: 如何添加新的视频模型？

A: 新模型会自动被识别（通过模型名称包含 "video"），无需修改代码。只需在定价配置中设置价格即可。

### Q: 视频生成超时怎么办？

A: 视频生成可能需要较长时间，确保：
1. 上游 API 支持长时间请求
2. 调整 `http.Client` 的超时设置
3. 考虑使用异步模式（Webhook 回调）

### Q: 如何支持基础号池的 720p 超分？

A: 已内置支持，服务端会先生成 480p 再按 `video.upscale_timing` 执行超分，对客户端透明。

### Q: 能否支持更长的视频（超过 30 秒）？

A: 服务端已支持 6-30 秒自动链式扩展，无需额外配置。如需支持更长视频，需修改 `ForwardVideoGeneration` 中的参数验证。

## 🎉 总结

本项目现已完整支持 Grok 系列模型的图片编辑和视频生成功能：

✅ 完全兼容 OpenAI API 格式  
✅ 透传架构，易于对接多个上游  
✅ 灵活的单次计费系统  
✅ 完整的前端集成和在线测试  
✅ 全面的类型安全和错误处理  

立即开始使用这些强大的 AI 媒体生成功能！
