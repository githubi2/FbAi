
INSERT INTO users (id, user_name, password, nick_name, email, phone, status, role_id, role_name, created_at, updated_at) 
VALUES 
  (1, 'admin', '$2b$12$Gvfzs74mwuLbZe2OLP5HT.Y7HfLMneULFMws6n6.NqXEBR/KhuSAS', '超级管理员', 'admin@art-design.com', '', 1, 1, '超级管理员', now(), now()),
  (2, 'alex', '$2b$12$Gvfzs74mwuLbZe2OLP5HT.Y7HfLMneULFMws6n6.NqXEBR/KhuSAS', 'Alex Morgan', 'alex@example.com', '18670001591', 1, 3, '普通用户', now(), now())
ON CONFLICT (id) DO UPDATE SET
  password = EXCLUDED.password,
  nick_name = EXCLUDED.nick_name,
  status = 1;

-- Reset sequence
SELECT setval('users_id_seq', (SELECT COALESCE(MAX(id), 1) FROM users));
