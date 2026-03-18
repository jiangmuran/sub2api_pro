# Grok 媒体功能快速开始指南

> **重要说明**: 系统会自动将上游 Grok API 的响应格式转换为 OpenAI 标准格式 (`{created, data: [{url}]}`），确保完全兼容所有 OpenAI 客户端。

## 🚀 5 分钟上手

### 1. 启动服务

```bash
# 编译并启动
make build
./backend/bin/server

# 或使用 Docker
docker-compose up -d
```

### 2. 运行数据库迁移

```bash
cd backend
go run cmd/migrate/main.go up
```

这将自动添加视频定价字段到 `groups` 表。

### 3. 配置上游 API

在账户管理页面，为你的 Grok API 账户设置 `openai_base_url`:

```
https://api.x.ai/v1
```

### 4. 测试图片编辑

#### 方式一：使用在线测试界面

1. 访问 `http://localhost:8000/user/model-test`
2. 输入或选择 API Key
3. 在 **图片编辑** 模块：
   - 选择模型: `grok-imagine-1.0-edit`
   - 上传图片
   - 输入提示词: "提高清晰度"
   - 点击 **开始编辑**

#### 方式二：使用 cURL

```bash
curl http://localhost:8000/v1/images/edits \
  -H "Authorization: Bearer sk-your-api-key" \
  -F "model=grok-imagine-1.0-edit" \
  -F "prompt=提高清晰度并增强色彩" \
  -F "image=@photo.jpg" \
  -F "size=1024x1024"
```

### 5. 测试视频生成

#### 方式一：使用在线测试界面

1. 访问 `http://localhost:8000/user/model-test`
2. 在 **视频生成** 模块：
   - 选择模型: `grok-imagine-1.0-video`
   - 输入提示词: "霓虹雨夜街头，慢镜头追拍"
   - 设置画面比例: `1280x720`
   - 设置时长: `18` 秒
   - 选择质量: `高清 (720p)`
   - 点击 **生成视频**

#### 方式二：使用 cURL

```bash
curl http://localhost:8000/v1/videos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-your-api-key" \
  -d '{
    "model": "grok-imagine-1.0-video",
    "prompt": "霓虹雨夜街头，慢镜头追拍，赛博朋克风格",
    "size": "1280x720",
    "seconds": 18,
    "quality": "high"
  }'
```

### 6. 配置定价（可选）

#### 数据库直接配置（推荐）

```sql
-- 为特定分组设置视频价格
UPDATE groups 
SET video_price_per_request = 0.08,    -- 标准质量
    video_price_per_request_hd = 0.16  -- 高清质量
WHERE id = 1;
```

#### 通过 API 配置（账户级）

```sql
-- 更新账户的 extra 字段
UPDATE accounts
SET extra = jsonb_set(
  COALESCE(extra, '{}'::jsonb),
  '{openai_manual_model_pricing,grok-imagine-1.0-video}',
  '{"video_price_per_request": 0.10, "video_price_per_request_hd": 0.20}'::jsonb
)
WHERE id = 1;
```

## 📊 验证计费

查看使用记录：

```sql
SELECT 
  id,
  model,
  media_type,
  video_count,
  video_quality,
  image_count,
  image_size,
  cost,
  duration_ms,
  created_at
FROM usage_logs
WHERE model LIKE '%grok%'
ORDER BY created_at DESC
LIMIT 10;
```

## 🎯 常见场景

### 场景 1: 图片去噪和增强

```bash
curl http://localhost:8000/v1/images/edits \
  -H "Authorization: Bearer $API_KEY" \
  -F "model=grok-imagine-1.0-edit" \
  -F "prompt=去除噪点，提高清晰度，增强色彩饱和度" \
  -F "image=@noisy_photo.jpg"
```

### 场景 2: 生成产品宣传视频

```bash
curl http://localhost:8000/v1/videos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "grok-imagine-1.0-video",
    "prompt": "白色背景，产品 360 度旋转展示，专业摄影灯光，高端质感",
    "size": "1280x720",
    "seconds": 12,
    "quality": "high"
  }'
```

### 场景 3: 基于参考图生成视频

```bash
curl http://localhost:8000/v1/videos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "grok-imagine-1.0-video",
    "prompt": "画面缓慢推进，添加电影般的景深效果",
    "size": "1792x1024",
    "seconds": 6,
    "quality": "standard",
    "image_reference": {
      "image_url": "https://example.com/reference.jpg"
    }
  }'
```

## 🔧 故障排除

### 问题 1: 上传图片失败

**症状**: `Failed to parse multipart form`

**解决方案**:
- 检查文件大小是否超过 32MB
- 确认文件格式为 PNG/JPG/WEBP
- 使用 `-F` 参数而不是 `-d`

### 问题 2: 视频生成超时

**症状**: 请求超时，没有响应

**解决方案**:
- 视频生成通常需要 30-60 秒，调整客户端超时设置
- 检查上游 API 是否正常
- 考虑缩短视频时长（6-12 秒）

### 问题 3: 计费金额不对

**症状**: 扣费金额与预期不符

**解决方案**:
1. 检查分组的 `rate_multiplier` 费率倍数
2. 确认 `video_price_per_request` / `video_price_per_request_hd` 配置
3. 查看日志中的实际计费详情

### 问题 4: 模型未出现在下拉列表

**症状**: 在线测试界面看不到 Grok 模型

**解决方案**:
- 确认 API Key 关联的分组有权访问 OpenAI 平台
- 刷新模型列表（点击刷新按钮）
- 手动输入模型名称测试

## 📈 性能优化建议

1. **使用标准质量优先**: 标准质量(480p)生成速度更快，成本更低
2. **合理控制时长**: 6-12 秒的视频生成速度较快
3. **批量处理**: 使用队列系统处理大量请求
4. **缓存结果**: 对于相同的提示词，考虑缓存生成结果

## 🎓 进阶用法

### 集成到你的前端应用

```typescript
import { useVideoAPI } from '@/composables/useVideoAPI'

export function MyComponent() {
  const { generateVideo, loading, progress } = useVideoAPI()
  
  const handleGenerate = async () => {
    try {
      const result = await generateVideo({
        apiKey: 'sk-your-key',
        model: 'grok-imagine-1.0-video',
        prompt: userInput.value,
        size: '1280x720',
        seconds: 18,
        quality: 'high'
      })
      
      videoUrl.value = result.data[0].url
      console.log('成本:', result.data[0].cost)
    } catch (error) {
      console.error('生成失败:', error)
    }
  }
  
  return { handleGenerate, loading, progress }
}
```

### Webhook 回调（异步模式）

对于长时间运行的视频生成，考虑使用 Webhook:

```bash
curl http://localhost:8000/v1/videos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "grok-imagine-1.0-video",
    "prompt": "长视频内容",
    "seconds": 30,
    "callback_url": "https://your-domain.com/webhook/video-complete"
  }'
```

## 💡 最佳实践

1. **提示词优化**: 明确描述场景、风格、运动方式
2. **参数调优**: 根据用途选择合适的尺寸和质量
3. **错误处理**: 实现重试机制和降级策略
4. **成本控制**: 监控使用量，设置预算限制
5. **用户体验**: 显示进度条和预估时间

## 🎉 完成！

现在你已经成功集成了 Grok 图片编辑和视频生成功能！

- 📖 详细文档: `GROK_MEDIA_API.md`
- 🐛 遇到问题？查看故障排除章节
- 💬 需要帮助？查看项目 Issues

祝你使用愉快！🚀
