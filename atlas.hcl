# src - the shadow schema, what the GORM models describe. `make db-diff`
#       rebuilds it with `go run ./cmd/schema` first, so it is never stale.
# url - the real database, where migrations are applied.
# dev - a throwaway container Atlas replays the migration files in to work out
#       what they add up to. Wiped on every run.
#
# The Makefile builds all three from .env.

env "local" {
  src = getenv("SHADOW_URL")
  url = getenv("DB_URL")
  dev = getenv("DEV_URL")

  # The format lives in the dir URL, not in a `format` attribute: `migrate
  # apply` has no --dir-format flag and only reads it from here.
  migration {
    dir = "file://cmd/schema/migrations?format=golang-migrate"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
