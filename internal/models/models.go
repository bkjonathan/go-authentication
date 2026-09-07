package models

import "gorm.io/gorm"

// All is every model in migration order - a table is listed after the tables it
// references, so a fresh database builds cleanly.
func All() []any {
	return []any{
		&User{},
		&RefreshToken{},
		&PasswordResetToken{},
	}
}

// Migrate brings the schema up to what the models declare. Useful in
// development; production schema changes should go through versioned
// migrations, which AutoMigrate deliberately is not (it never drops a column).
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(All()...)
}
