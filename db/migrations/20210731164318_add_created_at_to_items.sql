-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE "items" ADD "created_at" datetime;
UPDATE "items" set created_at = CURRENT_TIMESTAMP;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE "items" RENAME TO "items_backup";
CREATE TABLE "items" ("id" integer primary key autoincrement,"name" varchar(255),"email" varchar(255),"title" varchar(255),"description" varchar(255),"link" varchar(255),"published" datetime,"hide" bool,"feed_id" integer);
INSERT INTO "items" SELECT "id","name","email","title","description","link","published","hide","feed_id" from "items_backup";
DROP TABLE "items_backup";
