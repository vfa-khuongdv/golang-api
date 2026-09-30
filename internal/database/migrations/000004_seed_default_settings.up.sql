INSERT INTO `settings` (`key`, `value`, `created_at`, `updated_at`) VALUES
  ('mail.host', '127.0.0.1', NOW(3), NOW(3)),
  ('mail.port', '1026', NOW(3), NOW(3)),
  ('mail.username', '', NOW(3), NOW(3)),
  ('mail.password', '', NOW(3), NOW(3)),
  ('mail.from', 'noreply@example.com', NOW(3), NOW(3)),
  ('app.frontend_url', 'http://localhost:5173', NOW(3), NOW(3));
