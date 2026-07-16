CREATE TABLE IF NOT EXISTS `prediction_history` (
  `target_date` date NOT NULL,
  `created_at` datetime NOT NULL,
  `peak_start_hr` tinyint unsigned NOT NULL,
  `peak_end_hr` tinyint unsigned NOT NULL,
  `peak_wind` double NOT NULL,
  `peak_wind_lo` double NOT NULL,
  `peak_wind_hi` double NOT NULL,
  `peak_gust` double NOT NULL,
  `peak_gust_max` double NOT NULL,
  `peak_dir` double NOT NULL,
  `good_start_hr` tinyint unsigned NOT NULL,
  `good_end_hr` tinyint unsigned NOT NULL,
  `wind_down_hr` tinyint unsigned NOT NULL,
  `similar_days` int NOT NULL DEFAULT 0,
  PRIMARY KEY (`target_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
