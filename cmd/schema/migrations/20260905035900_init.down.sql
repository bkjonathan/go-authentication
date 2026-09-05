-- reverse: create index "idx_refresh_tokens_user_id_revoked_at" to table: "refresh_tokens"
DROP INDEX "idx_refresh_tokens_user_id_revoked_at";
-- reverse: create index "idx_refresh_tokens_family_id_revoked_at" to table: "refresh_tokens"
DROP INDEX "idx_refresh_tokens_family_id_revoked_at";
-- reverse: create index "idx_refresh_tokens_expires_at" to table: "refresh_tokens"
DROP INDEX "idx_refresh_tokens_expires_at";
-- reverse: create index "idx_refresh_tokens_deleted_at" to table: "refresh_tokens"
DROP INDEX "idx_refresh_tokens_deleted_at";
-- reverse: create "refresh_tokens" table
DROP TABLE "refresh_tokens";
-- reverse: create index "idx_password_reset_tokens_user_id_used_at" to table: "password_reset_tokens"
DROP INDEX "idx_password_reset_tokens_user_id_used_at";
-- reverse: create index "idx_password_reset_tokens_token_digest" to table: "password_reset_tokens"
DROP INDEX "idx_password_reset_tokens_token_digest";
-- reverse: create index "idx_password_reset_tokens_expires_at" to table: "password_reset_tokens"
DROP INDEX "idx_password_reset_tokens_expires_at";
-- reverse: create index "idx_password_reset_tokens_deleted_at" to table: "password_reset_tokens"
DROP INDEX "idx_password_reset_tokens_deleted_at";
-- reverse: create "password_reset_tokens" table
DROP TABLE "password_reset_tokens";
-- reverse: create index "idx_users_email" to table: "users"
DROP INDEX "idx_users_email";
-- reverse: create index "idx_users_deleted_at" to table: "users"
DROP INDEX "idx_users_deleted_at";
-- reverse: create "users" table
DROP TABLE "users";
