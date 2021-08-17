-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE "timestamps" ADD "name" varchar(255);
UPDATE "timestamps" set name = "feeds_fetched_at" where id = (SELECT id from "timestamps" limit 1);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE "timestamps" RENAME TO "timestamps_backup";
CREATE TABLE "timestamps" ("id" integer primary key autoincrement,"t" datetime);
INSERT INTO "timestamps" SELECT "id","t" from "timestamps_backup";
DROP TABLE "timestamps_backup";
