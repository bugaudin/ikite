ALTER TABLE `spots`
  ADD COLUMN `windguru_id` int(11) DEFAULT NULL AFTER `windguru_station_id`;

UPDATE `spots` SET `windguru_id` = 373090 WHERE `id` = 'ky';

CREATE TABLE IF NOT EXISTS `wind_forecast` (
  `forecast_date` date NOT NULL,
  `location` varchar(20) NOT NULL,
  `id_model` int NOT NULL,
  `model` varchar(32) NOT NULL,
  `period` datetime NOT NULL,
  `wind` double DEFAULT NULL,
  `gust` double DEFAULT NULL,
  `wind_dir` double DEFAULT NULL,
  `temp` double DEFAULT NULL,
  `fetched_at` datetime NOT NULL,
  PRIMARY KEY (`forecast_date`, `location`, `id_model`, `period`),
  KEY `idx_wind_forecast_loc_date` (`location`, `forecast_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
