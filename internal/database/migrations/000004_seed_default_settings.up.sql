INSERT INTO `settings` (`key`, `value`, `created_at`, `updated_at`) VALUES
  ('mail_host', 'smtp.gmail.com', NOW(3), NOW(3)),
  ('mail_port', '587', NOW(3), NOW(3)),
  ('mail_username', '', NOW(3), NOW(3)),
  ('mail_password', '', NOW(3), NOW(3)),
  ('mail_from', '', NOW(3), NOW(3)),
  ('frontend_url', 'http://localhost:5173', NOW(3), NOW(3));
