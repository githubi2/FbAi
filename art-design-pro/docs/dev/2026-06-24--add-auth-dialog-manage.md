# Add Authorization Dialog for Ad Account Manage Page — 2026-06-24

## Modified

| File | Change | Reason |
|------|--------|--------|
| `src/views/ad-account/manage/index.vue` | Added `addAuth` dialog template + logic; imports `fetchFbAuthUrl`, `fetchFbAccountList`, `ElAlert`, `ElIcon`, `ElInput`, `Loading` icon; `handleBatchAction` now routes `'addAuth'` to `handleOpenAddAuth` instead of requiring row selection | Implement "增加授权" dialog per user requirements |
| `src/locales/langs/zh.json` | Added `addAuthTitle`, `addAuthTip`, `addAuthSuccess` keys | i18n for new dialog UI text |
| `src/locales/langs/en.json` | Added English equivalents for above keys | i18n parity |

## Added

| File | Purpose |
|------|---------|
| Dialog template section | `ElDialog` with OAuth URL display (short link + full link), copy buttons, polling alert, success alert, cancel + confirm footer |
| `addAuthDialogVisible`, `addAuthShortUrl`, `addAuthFullUrl`, `addAuthSuccess`, `addAuthCopySuccess`, `isAddAuthPolling` | Reactive state for dialog lifecycle |
| `startAddAuthPolling()`, `stopAuthPolling()`, `stopAuthPollingAndClose()` | Polling management (3s interval, auto-stop on new account detected) |
| `handleOpenAddAuth()`, `handleAddAuthConfirm()` | Dialog open (generate URL + start poll) and confirm (close + refresh table) |
| `copyAddAuthShortUrl()`, `copyAddAuthFullUrl()`, `fallbackAddAuthCopy()` | Clipboard copy with fallback |
| `openAddAuthUrl()` | `window.open` the full auth URL |
| `.add-auth-dialog-body` SCSS | Layout styles for link boxes, tip text, action area, success bar |

## Why

The "增加授权" (Add Authorization) button on the ad account management page was a TODO placeholder. User provided a screenshot showing the dialog design. The implementation follows the existing FB OAuth authorization flow pattern from `views/ad-account/index.vue`:
1. Generate OAuth URL via `fetchFbAuthUrl()` (GET /api/v1/fb/auth-url)
2. Show short URL (recommended) + full URL with copy buttons
3. Poll `fetchFbAccountList()` every 3s to detect new authorization
4. Show green success alert when new account detected
5. Confirm button closes dialog and refreshes table

Unlike the existing auth dialog in `index.vue` which only has a cancel button, this dialog adds a confirm button (disabled until success) for explicit user acknowledgment.

## Dialog Layout

```
┌──────────────────────────────────────────┐
│  增加授权                                 │
├──────────────────────────────────────────┤
│  复制以下链接，在已登录 Facebook 的        │
│  浏览器中打开并完成授权：                  │
│                                          │
│  短链接（推荐）                           │
│  [URL input readonly] [复制链接]          │
│                                          │
│  完整链接                                 │
│  [URL input readonly] [复制链接]          │
│                                          │
│  [在当前浏览器打开]  或复制链接到其他浏览器  │
│                                          │
│  ✅ 授权成功！新授权的 FB 账号已添加       │
│  (green ElAlert — only after success)    │
│                                          │
│  ℹ️ 等待授权完成... (blue — during poll)  │
│                                          │
│          [取消]    [确定]                 │
└──────────────────────────────────────────┘
```
