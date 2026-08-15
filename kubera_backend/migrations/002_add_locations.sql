ALTER TABLE shops
    ADD COLUMN IF NOT EXISTS location_label TEXT,
    ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION;

ALTER TABLE suppliers
    ADD COLUMN IF NOT EXISTS location_label TEXT,
    ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION;

ALTER TABLE shops DROP CONSTRAINT IF EXISTS shops_location_coordinates_check;
ALTER TABLE shops ADD CONSTRAINT shops_location_coordinates_check CHECK (
    (latitude IS NULL AND longitude IS NULL) OR
    (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)
);

ALTER TABLE suppliers DROP CONSTRAINT IF EXISTS suppliers_location_coordinates_check;
ALTER TABLE suppliers ADD CONSTRAINT suppliers_location_coordinates_check CHECK (
    (latitude IS NULL AND longitude IS NULL) OR
    (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)
);
