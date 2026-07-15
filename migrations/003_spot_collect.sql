ALTER TABLE `spots`
  ADD COLUMN `collect` tinyint(1) NOT NULL DEFAULT 1 AFTER `visible`;
