
-- Update admin password to bcrypt
UPDATE users SET password = '$2b$12$OlNIY4PilrcuEHNxKDRGje0kI1CNM5O7FB.FIW48OyGnejRq9TLHW' WHERE user_name = 'admin';

-- Update role menu permissions (Menus name fixed, all 14 menus included)
UPDATE roles SET menu_ids = '{1,2,3,4,5,6,7,8,9,10,11,12,13,14}' WHERE role_code = 'R_SUPER';
UPDATE roles SET menu_ids = '{1,2,3}' WHERE role_code = 'R_ADMIN';
UPDATE roles SET menu_ids = '{1}' WHERE role_code = 'R_USER';

-- Insert demo users
INSERT INTO users (user_name, password, nick_name, email, phone, status, role_id, role_name) 
VALUES 
  ('alex', '$2b$12$UnCsG3LRojvPm188SA5YxOs8gnpmycJq1VGfKre/nAvW8VoJvm51K', 'Alex Morgan', 'alex@example.com', '18670001591', 1, 3, '普通用户'),
  ('sophia', '$2b$12$UnCsG3LRojvPm188SA5YxOs8gnpmycJq1VGfKre/nAvW8VoJvm51K', 'Sophia Baker', 'sophia@example.com', '17766664444', 1, 3, '普通用户'),
  ('liam', '$2b$12$UnCsG3LRojvPm188SA5YxOs8gnpmycJq1VGfKre/nAvW8VoJvm51K', 'Liam Park', 'liam@example.com', '18670001597', 1, 3, '普通用户'),
  ('olivia', '$2b$12$UnCsG3LRojvPm188SA5YxOs8gnpmycJq1VGfKre/nAvW8VoJvm51K', 'Olivia Grant', 'olivia@example.com', '18670001596', 1, 3, '普通用户'),
  ('emma', '$2b$12$UnCsG3LRojvPm188SA5YxOs8gnpmycJq1VGfKre/nAvW8VoJvm51K', 'Emma Wilson', 'emma@example.com', '18670001595', 1, 3, '普通用户')
ON CONFLICT (user_name) DO NOTHING;
