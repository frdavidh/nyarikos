ALTER TABLE "payments" DROP CONSTRAINT IF EXISTS payments_booking_id_fkey;
ALTER TABLE "bookings" DROP CONSTRAINT IF EXISTS bookings_room_id_fkey;
ALTER TABLE "bookings" DROP CONSTRAINT IF EXISTS bookings_user_id_fkey;
ALTER TABLE "room_facilities" DROP CONSTRAINT IF EXISTS room_facilities_facility_id_fkey;
ALTER TABLE "room_facilities" DROP CONSTRAINT IF EXISTS room_facilities_room_id_fkey;
ALTER TABLE "room_images" DROP CONSTRAINT IF EXISTS room_images_room_id_fkey;
ALTER TABLE "rooms" DROP CONSTRAINT IF EXISTS rooms_kost_id_fkey;
ALTER TABLE "kost_images" DROP CONSTRAINT IF EXISTS kost_images_kost_id_fkey;
ALTER TABLE "kosts" DROP CONSTRAINT IF EXISTS kosts_owner_id_fkey;
ALTER TABLE "social_accounts" DROP CONSTRAINT IF EXISTS social_accounts_user_id_fkey;
ALTER TABLE "refresh_tokens" DROP CONSTRAINT IF EXISTS refresh_tokens_user_id_fkey;

DROP INDEX IF EXISTS "idx_payments_booking_id";
DROP INDEX IF EXISTS "idx_bookings_user_id";
DROP INDEX IF EXISTS "idx_room_images_room_id";
DROP INDEX IF EXISTS "idx_rooms_price";
DROP INDEX IF EXISTS "idx_rooms_kost_id";
DROP INDEX IF EXISTS "idx_kost_images_kost_id";
DROP INDEX IF EXISTS "idx_kosts_city";
DROP INDEX IF EXISTS "idx_kosts_owner_id";
DROP INDEX IF EXISTS "idx_refresh_tokens_token";

DROP TABLE IF EXISTS "payments";
DROP TABLE IF EXISTS "bookings";
DROP TABLE IF EXISTS "room_facilities";
DROP TABLE IF EXISTS "facilities";
DROP TABLE IF EXISTS "room_images";
DROP TABLE IF EXISTS "rooms";
DROP TABLE IF EXISTS "kost_images";
DROP TABLE IF EXISTS "kosts";
DROP TABLE IF EXISTS "social_accounts";
DROP TABLE IF EXISTS "refresh_tokens";
DROP TABLE IF EXISTS "users";

DROP TYPE IF EXISTS "payment_method_type";
DROP TYPE IF EXISTS "payment_status";
DROP TYPE IF EXISTS "booking_status";
DROP TYPE IF EXISTS "oauth_provider";
DROP TYPE IF EXISTS "user_role";
