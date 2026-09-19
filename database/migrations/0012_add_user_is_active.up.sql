-- 0012_add_user_is_active.up.sql
-- Adds the is_active column to the user table to support account deactivation and soft-delete semantics.
ALTER TABLE `user`
  ADD COLUMN is_active TINYINT(1) NOT NULL DEFAULT 1 AFTER full_name;
