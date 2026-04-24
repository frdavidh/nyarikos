ALTER TYPE payment_method_type RENAME TO payment_method_type_old;

CREATE TYPE payment_method_type AS ENUM (
    'bank_transfer',
    'credit_card',
    'gopay',
    'shopeepay',
    'qris',
    'cstore'
);

ALTER TABLE payments
  ALTER COLUMN payment_method TYPE payment_method_type
  USING (
    CASE payment_method::text
      WHEN 'ewallet'       THEN 'gopay'::payment_method_type
      WHEN 'retail_outlet' THEN 'cstore'::payment_method_type
      ELSE payment_method::text::payment_method_type
    END
  );

DROP TYPE payment_method_type_old;