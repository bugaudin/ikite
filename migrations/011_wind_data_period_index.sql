-- idx_period_location duplicates uk_wind_data_period_location on (period, location).
ALTER TABLE wind_data DROP INDEX idx_period_location;

-- Speed up per-location latest/history lookups (ORDER BY period DESC).
ALTER TABLE wind_data ADD INDEX idx_location_period (location, period);
