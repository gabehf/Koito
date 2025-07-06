-- +goose Up
UPDATE users
SET username = LOWER(username);