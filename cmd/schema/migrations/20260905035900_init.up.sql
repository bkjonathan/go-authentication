-- create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "version" integer NOT NULL DEFAULT 1,
  "email" character varying(320) NOT NULL,
  "password_hash" character varying(255) NOT NULL,
  "first_name" character varying(100) NOT NULL,
  "last_name" character varying(100) NOT NULL,
  "photo_key" character varying(255) NULL,
  "role" character varying(32) NOT NULL DEFAULT 'staff',
  "status" character varying(32) NOT NULL DEFAULT 'active',
  "failed_login_attempts" integer NOT NULL DEFAULT 0,
  "locked_until" timestamptz NULL,
  "last_login_at" timestamptz NULL,
  "password_changed_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "chk_users_role" CHECK ((role)::text = ANY ((ARRAY['owner'::character varying, 'admin'::character varying, 'manager'::character varying, 'staff'::character varying])::text[])),
  CONSTRAINT "chk_users_status" CHECK ((status)::text = ANY ((ARRAY['active'::character varying, 'pending_verification'::character varying, 'suspended'::character varying, 'deactivated'::character varying])::text[]))
);
-- create index "idx_users_deleted_at" to table: "users"
CREATE INDEX "idx_users_deleted_at" ON "users" ("deleted_at");
-- create index "idx_users_email" to table: "users"
CREATE UNIQUE INDEX "idx_users_email" ON "users" ("email") WHERE (deleted_at IS NULL);
-- create "password_reset_tokens" table
CREATE TABLE "password_reset_tokens" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "version" integer NOT NULL DEFAULT 1,
  "user_id" uuid NOT NULL,
  "token_digest" character(64) NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "used_at" timestamptz NULL,
  "requested_from_ip" character varying(45) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_password_reset_tokens_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_password_reset_tokens_deleted_at" to table: "password_reset_tokens"
CREATE INDEX "idx_password_reset_tokens_deleted_at" ON "password_reset_tokens" ("deleted_at");
-- create index "idx_password_reset_tokens_expires_at" to table: "password_reset_tokens"
CREATE INDEX "idx_password_reset_tokens_expires_at" ON "password_reset_tokens" ("expires_at");
-- create index "idx_password_reset_tokens_token_digest" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "idx_password_reset_tokens_token_digest" ON "password_reset_tokens" ("token_digest");
-- create index "idx_password_reset_tokens_user_id_used_at" to table: "password_reset_tokens"
CREATE INDEX "idx_password_reset_tokens_user_id_used_at" ON "password_reset_tokens" ("user_id", "used_at");
-- create "refresh_tokens" table
CREATE TABLE "refresh_tokens" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "version" integer NOT NULL DEFAULT 1,
  "user_id" uuid NOT NULL,
  "family_id" uuid NOT NULL,
  "token_digest" character(64) NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "revoked_at" timestamptz NULL,
  "revoked_reason" character varying(32) NULL,
  "replaced_by_token_id" uuid NULL,
  "user_agent" character varying(255) NULL,
  "ip_address" character varying(45) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_refresh_tokens_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "chk_refresh_tokens_revoked_reason" CHECK ((revoked_reason IS NULL) OR ((revoked_reason)::text = ANY ((ARRAY['rotated'::character varying, 'signed_out'::character varying, 'password_changed'::character varying, 'reuse_detected'::character varying, 'admin_revoked'::character varying])::text[])))
);
-- create index "idx_refresh_tokens_deleted_at" to table: "refresh_tokens"
CREATE INDEX "idx_refresh_tokens_deleted_at" ON "refresh_tokens" ("deleted_at");
-- create index "idx_refresh_tokens_expires_at" to table: "refresh_tokens"
CREATE INDEX "idx_refresh_tokens_expires_at" ON "refresh_tokens" ("expires_at");
-- create index "idx_refresh_tokens_family_id_revoked_at" to table: "refresh_tokens"
CREATE INDEX "idx_refresh_tokens_family_id_revoked_at" ON "refresh_tokens" ("family_id", "revoked_at");
-- create index "idx_refresh_tokens_user_id_revoked_at" to table: "refresh_tokens"
CREATE INDEX "idx_refresh_tokens_user_id_revoked_at" ON "refresh_tokens" ("user_id", "revoked_at");
