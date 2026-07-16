ALTER TABLE `wind_data`
  ADD COLUMN `humidity` double DEFAULT NULL AFTER `temp`,
  ADD COLUMN `pressure` double DEFAULT NULL AFTER `humidity`;
