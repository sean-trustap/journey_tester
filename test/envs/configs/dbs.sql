-- This could be used by the test API.
CREATE DATABASE IF NOT EXISTS dev;
-- TODO Address disabling of `RFO5` (do not use special characters in
-- identifiers). We currently use `%` to allow these permissions on all hosts,
-- but this should be updated to the specific hosts defined inside the test
-- environment instead.
CREATE USER dev@'%' IDENTIFIED BY 'dev_pass'; -- noqa: RF05
GRANT ALL ON dev.* TO dev@'%'; -- noqa: RF05
