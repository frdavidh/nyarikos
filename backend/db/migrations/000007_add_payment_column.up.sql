ALTER TABLE payments
ADD COLUMN "invoice_number"    varchar UNIQUE,
ADD COLUMN "snap_token"        text,
ADD COLUMN "transaction_id"    varchar UNIQUE,
ADD COLUMN "payment_type"      varchar,
ADD COLUMN "va_number"         varchar,
ADD COLUMN "expiry_time"       timestamptz,
ADD COLUMN "midtrans_response" jsonb;