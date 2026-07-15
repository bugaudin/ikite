ALTER TABLE `spots`
  ADD COLUMN `collect_interval_min` int NOT NULL DEFAULT 5 AFTER `collect`,
  ADD COLUMN `collect_start_hour` int NOT NULL DEFAULT 8 AFTER `collect_interval_min`,
  ADD COLUMN `collect_end_hour` int NOT NULL DEFAULT 22 AFTER `collect_start_hour`;
