ALTER TABLE `wind_forecast`
  ADD COLUMN `windguru_id` int(11) NOT NULL DEFAULT 0 AFTER `location`;

UPDATE `wind_forecast` wf
  INNER JOIN `spots` s ON s.id = wf.location
  SET wf.windguru_id = s.windguru_id
  WHERE wf.windguru_id = 0 AND s.windguru_id IS NOT NULL;

ALTER TABLE `wind_forecast`
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (`forecast_date`, `windguru_id`, `id_model`, `period`),
  ADD KEY `idx_wind_forecast_wgid_date` (`windguru_id`, `forecast_date`);
