# Sub2API Chat Interface

## 概览

全新的移动优先聊天界面，无需登录即可使用，专为手机用户优化。访问地址：`https://api.ai.org.kg/chat`

## 核心功能

### ✅ 已实现

#### 1. 聊天功能
- ✅ **多轮对话** - 完整的上下文对话
- ✅ **流式输出** - 打字机效果实时显示
- ✅ **模型切换** - 从 `/v1/models` 动态加载可用模型
- ✅ **Markdown 渲染** - 支持代码高亮、表格、列表等
- ✅ **对话管理** - 新建、删除、重命名对话
- ✅ **本地存储** - 使用 IndexedDB 存储，支持离线访问
- ✅ **消费显示** - 实时显示当前对话和累计消费

#### 2. 生图功能
- ✅ **基础生图** - 支持 Grok 图片模型
- ✅ **尺寸选择** - 1024x1024, 1024x1792, 1792x1024
- ✅ **数量设置** - 一次生成 1-4 张
- ✅ **历史记录** - 查看所有生成的图片
- ✅ **图片预览** - 点击放大查看
- ✅ **下载保存** - 一键下载到本地
- ✅ **离线缓存** - 图片转为 Data URL 存储

#### 3. 移动端优化
- ✅ **响应式设计** - 完美适配手机、平板、桌面
- ✅ **触摸手势** - 滑动操作侧边栏
- ✅ **折叠侧边栏** - 抽屉式对话列表
- ✅ **虚拟键盘适配** - 输入框自动上推
- ✅ **底部导航** - Tab 切换聊天/生图

## 技术架构

### 前端
- **框架**: Vue 3 + TypeScript + Vite
- **样式**: Tailwind CSS
- **存储**: IndexedDB (本地持久化)
- **Markdown**: marked + DOMPurify
- **代码高亮**: highlight.js

### API
- **聊天**: `POST /v1/chat/completions` (OpenAI 格式)
- **生图**: `POST /v1/images/generations` (OpenAI 格式)
- **模型**: `GET /v1/models`
- **价格**: 前端硬编码常用模型价格

### 文件结构
```
frontend/src/
├── types/chat.ts                          # 类型定义
├── composables/
│   ├── useChatAPI.ts                     # 聊天 API
│   ├── useImageAPI.ts                    # 生图 API
│   └── useChatStorage.ts                 # IndexedDB 存储
├── components/chat/
│   └── ChatMessage.vue                   # 消息组件
└── views/public/
    └── ChatView.vue                      # 主界面
```

## 使用说明

### 首次访问

1. 访问 `https://api.ai.org.kg/chat`
2. 输入你的 Sub2API API Key (sk-...)
3. 系统会自动验证并加载可用模型

### 聊天

1. 点击 "New Chat" 创建新对话
2. 在底部输入框输入消息
3. 按 Enter 发送，Shift+Enter 换行
4. 实时查看流式响应和消费金额
5. 支持重新生成回复

### 生图

1. 点击底部 "🎨 Images" Tab
2. 输入图片描述
3. 选择尺寸和数量
4. 点击 "Generate" 生成
5. 点击图片预览，点击下载按钮保存

### 移动端操作

- **打开侧边栏**: 点击左上角菜单图标
- **关闭侧边栏**: 点击遮罩层或向左滑动
- **切换对话**: 在侧边栏点击对话卡片
- **删除对话**: 点击对话卡片上的删除按钮
- **切换 Tab**: 点击底部 "Chat" 或 "Images"

## 数据存储

所有数据存储在浏览器 IndexedDB 中：

- **API Key**: localStorage (`sub2api_chat_key`)
- **对话记录**: IndexedDB (`conversations` 表)
- **消息内容**: IndexedDB (`messages` 表)
- **生成图片**: IndexedDB (`images` 表，含 Data URL 缓存)
- **累计消费**: localStorage (`total_cost`)

### 清除数据

1. 浏览器开发者工具 → Application → Storage
2. 清除 IndexedDB `sub2api-chat` 数据库
3. 清除 localStorage `sub2api_chat_key` 和 `total_cost`

或直接在聊天界面点击 "🔑 Change Key" 更换 API Key（数据保留）

## 消费计算

### 聊天模型价格 (每百万 token)

| 模型 | Input | Output |
|------|-------|--------|
| gpt-4o | $2.50 | $10.00 |
| gpt-4o-mini | $0.15 | $0.60 |
| claude-sonnet-4 | $3.00 | $15.00 |
| claude-opus-4 | $15.00 | $75.00 |
| grok-2-1212 | $2.00 | $10.00 |

### 图片价格

| 尺寸 | Grok | DALL-E 3 |
|------|------|----------|
| 1024×1024 | $0.04 | $0.04 |
| 1024×1792 | $0.06 | $0.08 |
| 1792×1024 | $0.06 | $0.08 |

价格会从后端 API 响应的 usage 字段实时计算。

## 性能优化

- **代码分割**: 路由级别懒加载
- **图片缓存**: 转换为 Data URL 避免过期
- **虚拟滚动**: 处理大量消息（未来优化）
- **请求中断**: 支持取消正在进行的生成
- **自动重试**: API 失败自动切换账户（后端）

## 浏览器兼容性

- ✅ Chrome 90+
- ✅ Safari 14+
- ✅ Firefox 88+
- ✅ Edge 90+
- ⚠️ 需要支持 IndexedDB API

## 已知限制

1. **无云端同步** - 数据仅存本地，换设备需重新设置
2. **无账户系统** - 完全独立于 Sub2API 用户系统
3. **图片 URL 过期** - 虽有 Data URL 缓存，但大量图片会占用空间
4. **无导出功能** - 暂不支持导出聊天记录（未来计划）

## 未来计划

- [ ] 对话导出（Markdown/JSON/PDF）
- [ ] 语音输入（Web Speech API）
- [ ] PWA 支持（可安装到桌面）
- [ ] 主题切换（亮色/暗色）
- [ ] 快捷指令（预设 Prompt）
- [ ] 图片编辑（基于生成结果编辑）
- [ ] 使用统计图表
- [ ] 分享对话链接

## 故障排除

### API Key 验证失败
- 检查 Key 是否正确（以 `sk-` 开头）
- 确认 Key 未过期且有余额
- 检查浏览器控制台错误信息

### 消息无法发送
- 检查网络连接
- 查看浏览器控制台 Network 面板
- 确认后端服务正常运行

### 图片无法生成
- 检查 prompt 是否符合内容政策
- 确认 API Key 有图片生成权限
- 查看控制台错误详情

### 数据丢失
- IndexedDB 可能被浏览器清理（隐私模式）
- 检查浏览器存储设置
- 定期导出重要对话（未来功能）

## 开发调试

```bash
# 前端开发
cd frontend
pnpm install
pnpm dev

# 构建部署
pnpm build

# 后端重启（应用前端更新）
ssh root@api.ai.org.kg "systemctl restart sub2api.service"
```

## 贡献

如有问题或建议，请提交 Issue 或 Pull Request。

## 许可证

遵循 Sub2API Pro 主项目许可证。
