-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE IF NOT EXISTS "timestamps" ("id" integer primary key autoincrement,"t" datetime);
CREATE TABLE IF NOT EXISTS "feeds" ("id" integer primary key autoincrement,"url" varchar(255));
CREATE TABLE IF NOT EXISTS "tags" ("id" integer primary key autoincrement,"name" varchar(255),"feed_id" integer);
CREATE TABLE IF NOT EXISTS "items" ("id" integer primary key autoincrement,"name" varchar(255),"email" varchar(255),"title" varchar(255),"description" varchar(255),"link" varchar(255),"published" datetime,"hide" bool,"feed_id" integer);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE "timestamps";
DROP TABLE "feeds";
DROP TABLE "tags";
DROP TABLE "items";
