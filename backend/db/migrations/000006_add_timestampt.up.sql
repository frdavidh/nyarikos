ALTER TABLE kost_images
ADD COLUMN updated_at timestamp NOT NULL DEFAULT NOW(),
ADD COLUMN deleted_at timestamp;