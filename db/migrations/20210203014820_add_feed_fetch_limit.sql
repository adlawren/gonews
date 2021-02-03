-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE "feeds" ADD "fetch_limit" integer DEFAULT 0;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE "feeds" RENAME TO "feeds_backup";
CREATE TABLE "feeds" ("id" integer primary key autoincrement,"url" varchar(255));
INSERT INTO "feeds" SELECT "id","url" from "feeds_backup";
DROP TABLE "feeds_backup";
