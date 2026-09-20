// Package tenant is the single place every tenant-owned query is scoped
// through. v1 is single-coach (per docs/spec.md), so "tenant" is a coach id
// today, but docs/product-direction.md section 7 flags that a gym account
// will eventually need an org layer above coach. Routing every scoping
// decision through this package means that day is one file to change
// (Scope's signature and body) instead of a rewrite of every repository's
// WHERE clause.
//
// Row-level security: not implemented. Postgres RLS would be reasonable
// defense in depth (a query that forgets to call Scope would still be
// blocked at the DB), but v1 has exactly one enforcement path — this
// package — and every tenant-owned repository is required to go through it,
// so a second, DB-level copy of the same rule is redundant for now. Revisit
// once org_id lands and more than one code path can construct a query
// against tenant-owned tables (e.g. a background job, an admin console).
package tenant

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// contextKey is the Gin context key the authenticated coach's id is stored
// under. It is unexported so callers can only read/write it through
// SetCoachID/CoachIDFromContext, keeping this package the single choke point
// for tenant scoping.
const contextKey = "tenant_coach_id"

// SetCoachID records the authenticated coach's id on the Gin context. Called
// once, by the auth middleware, after a request's JWT has been validated.
func SetCoachID(c *gin.Context, coachID uuid.UUID) {
	c.Set(contextKey, coachID)
}

// CoachIDFromContext returns the authenticated coach's id set by the auth
// middleware, and false if the request reached this point unauthenticated.
func CoachIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(contextKey)
	if !ok {
		return uuid.UUID{}, false
	}

	coachID, ok := v.(uuid.UUID)
	return coachID, ok
}

// Scope returns db scoped to rows owned by coachID. Every repository query
// against a tenant-owned table (one with a coach_id column) must be built
// from this, never from a bare db.Where("coach_id = ?", ...) — that
// duplication is exactly what let a forgotten WHERE clause leak another
// coach's data in the first place.
func Scope(db *gorm.DB, coachID uuid.UUID) *gorm.DB {
	return db.Where("coach_id = ?", coachID)
}
