-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE "feeds" RENAME TO "feeds_backup";
CREATE TABLE "feeds" ("id" integer primary key autoincrement,"url" varchar(255), "fetch_limit" integer DEFAULT 0);
INSERT INTO "feeds" SELECT "id","url","fetch_limit" from "feeds_backup";
DROP TABLE "feeds_backup";

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE "feeds" ADD "auto_dismissed_at" datetime;
UPDATE "feeds" set auto_dismissed_at = CURRENT_TIMESTAMP;
