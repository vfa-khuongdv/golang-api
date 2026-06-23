CREATE TABLE `refresh_tokens` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `refresh_token` varchar(60) COLLATE utf8mb4_unicode_ci NOT NULL UNIQUE,
  `ip_address` varchar(45) COLLATE utf8mb4_unicode_ci NOT NULL,
  `expired_at` bigint NOT NULL,
  `user_id` bigint UNSIGNED NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_refresh_tokens_deleted_at` (`deleted_at`),
  KEY `fk_refresh_tokens_user` (`user_id`),
  KEY `idx_refresh_token_expired` (`refresh_token`, `expired_at`),
  CONSTRAINT `fk_refresh_tokens_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
