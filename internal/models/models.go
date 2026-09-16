// Package models holds GORM model structs for all entities in
// docs/er-diagram.md. This is a temporary home: as each module (coach,
// athlete, plan, movement, ...) is built, its model(s) should move into
// that module's own model.go per the module-first layout, and this package
// should shrink until it can be deleted.
//
// Column names, types, and nullability must match migrations/*.sql exactly.
// GORM's AutoMigrate is never used — migrations are the single source of
// truth for the schema.
package models
