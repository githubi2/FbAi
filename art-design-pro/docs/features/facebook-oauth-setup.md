# Facebook OAuth 前置条件 — 详细设置步骤

> 最后更新: 2026-06-17

---

## 概览

在后台系统使用 Facebook OAuth 之前，需要先在 [Meta for Developers](https://developers.facebook.com) 完成以下配置。整个过程约需 **15-30 分钟**（含应用审核等待，Basic 权限即时可用）。

---

## 第一步：注册 Meta 开发者账号

1. 打开 https://developers.facebook.com
2. 点击右上角 **「我的应用」→「创建应用」**
3. 如果未注册开发者账号，按提示验证手机号或信用卡（Meta 要求）
4. 完成注册后会进入应用管理面板

> ⚠️ **必须是个人 Facebook 账号**（不能是 Business 账号），且账号需要已绑定手机号。

---

## 第二步：创建 Facebook 应用

1. 点击 **「创建应用」** 绿色按钮
2. 选择应用类型：**「其他」→「商务」**（或「无类型」均可）
3. 填写：
   | 字段 | 值 | 说明 |
   |------|----|------|
   | 应用名称 | `Your App Name` | 例如 `art-design-ads` |
   | 联系邮箱 | 你的邮箱 | 用于接收通知 |
   | 商务管理平台 | 可选 | 如果没有 BM 可以跳过 |
4. 点击 **「创建应用」**
5. 输入密码确认

> 📌 记录下 **「应用编号 (App ID)」**，顶部导航栏可见。

---

## 第三步：添加「Facebook 登录」产品

1. 进入应用面板后，左侧菜单 → **「添加产品」**
2. 找到 **「Facebook 登录」**（不是 "Instagram 登录" 或 "受限登录"）
3. 点击 **「设置」→ 选择「Web」**
4. 填入站点 URL：`http://localhost:3006`（开发环境）

![dashboard] 应用面板 → 产品 → Facebook 登录 → 设置 → Web

---

## 第四步：配置 OAuth 回调 URL（最关键）

1. 左侧菜单 → **「Facebook 登录」→「设置」**
2. 找到 **「有效的 OAuth 跳转 URI」** 输入框
3. 填入以下 URL（每行一个）：

   ```
   http://localhost:9090/api/v1/fb/callback
   ```

   > **生产环境需改为 HTTPS 域名**，例如：
   > `https://your-domain.com/api/v1/fb/callback`

4. 点击 **「保存更改」**（页面底部）

![setting] 左侧 Facebook 登录 → 设置 → 有效的 OAuth 跳转 URI

---

## 第五步：配置应用域名（可选，本地开发可跳过）

如果生产环境使用自有域名：

1. **「设置」→「基本」**
2. 找到 **「应用域名」** 字段
3. 填入域名（不带协议，如 `your-domain.com`）
4. 填入 **「隐私权政策网址」** 和 **「用户数据删除」** URL（必填才能上线）
5. 保存

> 📌 开发环境下可以暂时填 `http://localhost:3006` 作为域名。

---

## 第六步：获取 App Secret

1. **「设置」→「基本」**
2. 找到 **「应用密钥 (App Secret)」**
3. 点击 **「显示」**，输入密码确认
4. 复制密钥

> 🔐 **App Secret 绝对不能提交到 Git！它已通过 .env 文件管理且 .env 在 .gitignore 中。**

---

## 第七步：申请 Marketing API 权限

这是操作广告账户必需的权限。

### 7a. 获取 Basic 权限（即时可用，无需审核）

1. 左侧菜单 → **「应用审核」→「权限和功能」**
2. 搜索以下权限并点击 **「获取访问口令」** / **「添加到应用」**：
   - `ads_read` — 读取广告数据
   - `ads_management` — 管理广告
   - `business_management` — 管理商务管理平台（BM）
3. 在开发模式下，这些权限对 **应用管理员/开发者/测试用户** 即时生效

### 7b. 发布上线（需要 Meta 审核）

如果非开发者用户也需要使用（生产环境）：

1. 上述三项权限 → 点击 **「申请高级访问」**
2. 填写使用说明：
   - **ads_read**: "读取用户绑定的广告账户数据，在管理后台展示广告报表"
   - **ads_management**: "允许用户在管理后台创建/编辑/暂停广告系列"
   - **business_management**: "获取用户可访问的 BM 列表，用于关联广告账户"
3. 上传录屏展示功能（Meta 审核要求）
4. 提交后等待审核（通常 1-3 个工作日）

> 📌 **开发测试阶段不需要审核**。只要你是应用的管理员/开发者，这些权限可以直接使用。

---

## 第八步：添加测试用户（可选）

如果你的 Facebook 账号不是应用管理员，可以添加为测试用户：

1. **「应用面板」→「角色」→「测试用户」**
2. 点击 **「添加」**
3. 生成或输入测试用户的 Facebook 账号
4. 该测试用户可以授权应用并测试广告 API

---

## 第九步：配置 .env 文件

将获取到的凭据填入后台配置：

```bash
# art-design-server/.env

# ==================== Facebook OAuth 配置 ====================
FB_APP_ID=123456789012345          # 替换为你的 App ID
FB_APP_SECRET=your_app_secret_here    # 替换为 App Secret
FB_REDIRECT_URI=http://localhost:9090/api/v1/fb/callback
FB_GRAPH_VERSION=v22.0             # 默认 v22.0，参考 Meta 最新稳定版
```

可使用以下命令快速追加：

```bash
cd E:/FbAi/art-design-server
# 如果已有旧的 FB_ 配置，先清除
grep -v "FB_" .env > .env.tmp && mv .env.tmp .env

# 追加新配置
cat >> .env << 'EOF'

# ==================== Facebook OAuth 配置 ====================
FB_APP_ID=你的APP_ID
FB_APP_SECRET=你的APP_SECRET
FB_REDIRECT_URI=http://localhost:9090/api/v1/fb/callback
FB_GRAPH_VERSION=v22.0
EOF
```

---

## 第十步：重启后端服务

修改 .env 后需要重启：

```bash
cd E:/FbAi/art-design-server
taskkill //F //IM server.exe 2>/dev/null
GOPROXY=https://goproxy.cn,direct go build -o server.exe ./main.go
./server.exe &
```

验证配置生效：

```bash
# 登录获取 token
TOKEN=$(curl -s -X POST http://localhost:9090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"userName":"admin","password":"admin123"}' | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['token'])")

# 测试授权链接生成
curl -s http://localhost:9090/api/v1/fb/auth-url \
  -H "Authorization: Bearer $TOKEN"

# 正确输出示例:
# {"code":200,"msg":"success","data":{"authUrl":"https://graph.facebook.com/v22.0/oauth/authorize?..."}}
```

---

## 第十一步：完整链路测试

### 11a. 前端页面测试

1. 启动前端 `cd art-design-pro && pnpm dev`
2. 浏览器打开 `http://localhost:3006/#/ad-account/list`
3. 应该看到：
   - 连接状态面板：**「未连接」**标签
   - **「连接 Facebook」**蓝色按钮
4. 点击按钮 → 新窗口打开 Facebook 授权页面
5. 授权后自动跳转回 `/#/ad-account/list?fb_connected=success`
6. 页面自动刷新，显示连接状态和广告账户列表

### 11b. 状态查询

```bash
curl -s http://localhost:9090/api/v1/fb/status \
  -H "Authorization: Bearer $TOKEN"
# {"code":200,"data":{"connected":true,"fbUserName":"你的FB名字",...}}
```

### 11c. 广告账户列表

```bash
curl -s http://localhost:9090/api/v1/fb/ad-accounts \
  -H "Authorization: Bearer $TOKEN"
# {"code":200,"data":{"adAccounts":[...],"businesses":[...]}}
```

---

## 常见问题排查

| 问题 | 原因 | 解决 |
|------|------|------|
| 授权链接返回 500 "Facebook 应用未配置" | `.env` 中 FB_APP_ID 为空 | 检查 `.env`，确保填入了真实值 |
| Facebook 授权页显示 "应用未设置" | OAuth 跳转 URI 未配置 | 检查第四步，确认 URI 完全匹配 |
| 回调后 500 "无效的 state 参数" | state 过期（>5分钟）或 CSRF 不匹配 | 重新点击"连接 Facebook"获取新链接 |
| 广告账户列表为空 | 你的 Facebook 账号没有广告账户 | 在 BM 中创建广告账户，或确认有 `ads_read` 权限 |
| 授权页报错 "功能不可用" | 应用处于开发模式且非管理员访问 | 参考第八步添加测试用户 |
| `ads_management` 权限报错 | 权限未申请 | 参考第七步申请 |

---

## Meta 开发者平台关键页面速查

| 页面 | 路径 | 用途 |
|------|------|------|
| 应用列表 | https://developers.facebook.com/apps | 创建/切换应用 |
| 基本设置 | 应用面板 → 设置 → 基本 | App ID, App Secret, 域名 |
| Facebook 登录设置 | 应用面板 → Facebook 登录 → 设置 | OAuth 回调 URI |
| 权限和功能 | 应用面板 → 应用审核 → 权限和功能 | 申请 ads_read/ads_management 等 |
| Graph API Explorer | https://developers.facebook.com/tools/explorer | 调试 API 调用 |
| Access Token 调试器 | https://developers.facebook.com/tools/debug/accesstoken | 检查 token 有效性 |
