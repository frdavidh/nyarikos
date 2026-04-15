ALTER TABLE "room_facilities"
ADD CONSTRAINT "uq_room_facilities" UNIQUE ("room_id", "facility_id");
