CREATE TYPE "user_role" AS ENUM (
  'pencari',
  'pemilik',
  'admin'
);

CREATE TYPE "oauth_provider" AS ENUM (
  'google',
  'facebook'
);

CREATE TYPE "booking_status" AS ENUM (
  'pending',
  'paid',
  'cancelled',
  'completed'
);

CREATE TYPE "payment_status" AS ENUM (
  'pending',
  'success',
  'failed',
  'expired'
);

CREATE TYPE "payment_method_type" AS ENUM (
  'bank_transfer',
  'credit_card',
  'ewallet',
  'qris',
  'retail_outlet'
);

CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password" varchar,
  "phone_number" varchar,
  "role" user_role DEFAULT 'pencari',
  "is_active" boolean DEFAULT true,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz DEFAULT now(),
  "deleted_at" timestamptz
);

CREATE TABLE "refresh_tokens" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "token" text UNIQUE NOT NULL,
  "device_info" varchar,
  "ip_address" varchar,
  "expires_at" timestamptz NOT NULL,
  "is_revoked" boolean DEFAULT false,
  "created_at" timestamptz DEFAULT now()
);

CREATE TABLE "social_accounts" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "provider_name" oauth_provider NOT NULL,
  "provider_id" varchar UNIQUE NOT NULL,
  "created_at" timestamptz DEFAULT now()
);

CREATE TABLE "kosts" (
  "id" bigserial PRIMARY KEY,
  "owner_id" bigint NOT NULL,
  "name" varchar NOT NULL,
  "description" text,
  "address" text NOT NULL,
  "city" varchar NOT NULL,
  "is_premium" boolean DEFAULT false,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz DEFAULT now(),
  "deleted_at" timestamptz
);

CREATE TABLE "kost_images" (
  "id" bigserial PRIMARY KEY,
  "kost_id" bigint NOT NULL,
  "image_url" text NOT NULL,
  "is_main" boolean DEFAULT false,
  "created_at" timestamptz DEFAULT now()
);

CREATE TABLE "rooms" (
  "id" bigserial PRIMARY KEY,
  "kost_id" bigint NOT NULL,
  "room_type" varchar NOT NULL,
  "price_per_month" decimal(12,2) NOT NULL,
  "total_rooms" int NOT NULL DEFAULT 1,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz DEFAULT now(),
  "deleted_at" timestamptz
);

CREATE TABLE "room_images" (
  "id" bigserial PRIMARY KEY,
  "room_id" bigint NOT NULL,
  "image_url" text NOT NULL,
  "is_main" boolean DEFAULT false,
  "created_at" timestamptz DEFAULT now()
);

CREATE TABLE "facilities" (
  "id" bigserial PRIMARY KEY,
  "name" varchar UNIQUE NOT NULL,
  "icon_url" text
);

CREATE TABLE "room_facilities" (
  "id" bigserial PRIMARY KEY,
  "room_id" bigint NOT NULL,
  "facility_id" bigint NOT NULL
);

CREATE TABLE "bookings" (
  "id" bigserial PRIMARY KEY,
  "booking_code" varchar UNIQUE NOT NULL,
  "user_id" bigint NOT NULL,
  "room_id" bigint NOT NULL,
  "start_date" date NOT NULL,
  "end_date" date NOT NULL,
  "durations_months" int NOT NULL DEFAULT 1,
  "total_price" decimal(12,2),
  "status" booking_status DEFAULT 'pending',
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz DEFAULT now()
);

CREATE TABLE "payments" (
  "id" bigserial PRIMARY KEY,
  "booking_id" bigint NOT NULL,
  "external_id" varchar UNIQUE NOT NULL,
  "payment_method" payment_method_type,
  "amount" decimal(12,2) NOT NULL,
  "status" payment_status DEFAULT 'pending',
  "checkout_url" text,
  "paid_at" timestamptz,
  "created_at" timestamptz DEFAULT now(),
  "updated_at" timestamptz DEFAULT now()
);

CREATE INDEX "idx_refresh_tokens_token" ON "refresh_tokens" ("token");
CREATE INDEX "idx_kosts_owner_id" ON "kosts" ("owner_id");
CREATE INDEX "idx_kosts_city" ON "kosts" ("city");
CREATE INDEX "idx_kost_images_kost_id" ON "kost_images" ("kost_id");
CREATE INDEX "idx_rooms_kost_id" ON "rooms" ("kost_id");
CREATE INDEX "idx_rooms_price" ON "rooms" ("price_per_month");
CREATE INDEX "idx_room_images_room_id" ON "room_images" ("room_id");
CREATE INDEX "idx_bookings_user_id" ON "bookings" ("user_id");
CREATE INDEX "idx_payments_booking_id" ON "payments" ("booking_id");

ALTER TABLE "refresh_tokens" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "social_accounts" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "kosts" ADD FOREIGN KEY ("owner_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "kost_images" ADD FOREIGN KEY ("kost_id") REFERENCES "kosts" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "rooms" ADD FOREIGN KEY ("kost_id") REFERENCES "kosts" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "room_images" ADD FOREIGN KEY ("room_id") REFERENCES "rooms" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "room_facilities" ADD FOREIGN KEY ("room_id") REFERENCES "rooms" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "room_facilities" ADD FOREIGN KEY ("facility_id") REFERENCES "facilities" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "bookings" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "bookings" ADD FOREIGN KEY ("room_id") REFERENCES "rooms" ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE "payments" ADD FOREIGN KEY ("booking_id") REFERENCES "bookings" ("id") DEFERRABLE INITIALLY IMMEDIATE;
