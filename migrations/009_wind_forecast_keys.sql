-- Cluster by windguru spot first (main read filter), then day / model / hour.
ALTER TABLE `wind_forecast`
  DROP PRIMARY KEY,
  ADD PRIMARY KEY (`windguru_id`, `forecast_date`, `id_model`, `period`);

ALTER TABLE `wind_forecast`
  DROP INDEX `idx_wind_forecast_loc_date`,
  DROP INDEX `idx_wind_forecast_wgid_date`;

-- Latest forecast day per spot (e.g. "most recent fetch for windguru_id=?").
ALTER TABLE `wind_forecast`
  ADD KEY `idx_wind_forecast_wgid_date_desc` (`windguru_id`, `forecast_date` DESC);

-- One model's hourly series for a spot on a given day.
ALTER TABLE `wind_forecast`
  ADD KEY `idx_wind_forecast_wgid_model_period` (`windguru_id`, `id_model`, `forecast_date`, `period`);
