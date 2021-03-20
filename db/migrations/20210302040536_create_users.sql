-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE IF NOT EXISTS "users" ("id" integer primary key autoincrement,"username" varchar(255) unique,"password_hash" varchar(255));

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE "users";
