-- 006_fb_last_error: 为 fb_tokens 添加错误追踪字段
-- 用于记录 FB API 调用失败时的错误信息，前端列表页展示异常状态

ALTER TABLE fb_tokens ADD COLUMN IF NOT EXISTS last_error TEXT DEFAULT '';
ALTER TABLE fb_tokens ADD COLUMN IF NOT EXISTS last_error_at TIMESTAMPTZ DEFAULT NULL;
