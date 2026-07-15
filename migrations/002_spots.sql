CREATE TABLE IF NOT EXISTS `spots` (
  `id` varchar(20) NOT NULL,
  `name` varchar(100) NOT NULL,
  `windguru_station_id` int DEFAULT NULL,
  `sort_order` int NOT NULL DEFAULT 0,
  `visible` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_windguru_station` (`windguru_station_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO `spots` (`id`, `name`, `windguru_station_id`, `sort_order`, `visible`) VALUES
  ('ky', 'Kiryat Yam', NULL, 10, 1),
  ('kh', 'Kiryat Haim', NULL, 20, 0),
  ('15233', 'Betzet', 15233, 30, 1),
  ('st', 'Shavei Tzion', 2763, 40, 1),
  ('bg', 'Bat Galim', 2049, 50, 1),
  ('2256', 'Atlit', 2256, 60, 1),
  ('hp', 'Hadera', 3377, 70, 1),
  ('2752', 'Sea of G', 2752, 80, 1),
  ('5730', 'Merom', 5730, 90, 0),
  ('5731', 'Zemach', 5731, 100, 0),
  ('5732', 'Avney Eitan', 5732, 110, 0),
  ('1909', 'Diamond', 1909, 120, 0),
  ('3379', 'Kineret', 3379, 130, 0),
  ('1091', 'Paros', 1091, 140, 0),
  ('5500', 'Mykonos', 5500, 150, 0),
  ('14905', 'Squamish', 14905, 160, 0),
  ('2667', 'Tarifa', 2667, 170, 0),
  ('3708', 'Los Alcazares', 3708, 180, 0);
