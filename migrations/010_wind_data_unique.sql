DELETE w1 FROM wind_data w1
INNER JOIN wind_data w2
  ON w1.period = w2.period AND w1.location = w2.location AND w1.id < w2.id;

ALTER TABLE wind_data
  ADD UNIQUE KEY uk_wind_data_period_location (period, location);
