-- create index "idx_refresh_tokens_token_digest" to table: "refresh_tokens"
CREATE UNIQUE INDEX "idx_refresh_tokens_token_digest" ON "refresh_tokens" ("token_digest");
