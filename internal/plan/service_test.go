package plan

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func planStarting(date string) Plan {
	start, err := time.Parse(time.DateOnly, date)
	if err != nil {
		panic(err)
	}
	return Plan{ID: uuid.New(), StartDate: start}
}

func TestCurrentPlanID(t *testing.T) {
	tehran := time.FixedZone("IRST", 3*60*60+30*60)
	now := time.Date(2026, 10, 6, 23, 30, 0, 0, tehran)

	t.Run("latest started plan wins over a future one", func(t *testing.T) {
		future, recent, old := planStarting("2026-11-01"), planStarting("2026-09-15"), planStarting("2026-07-01")
		assert.Equal(t, recent.ID, currentPlanID([]Plan{future, recent, old}, now))
	})

	t.Run("a plan starting today is current", func(t *testing.T) {
		today, old := planStarting("2026-10-06"), planStarting("2026-07-01")
		assert.Equal(t, today.ID, currentPlanID([]Plan{today, old}, now),
			"today is the local date, even though it's already 2026-10-06T20:00Z")
	})

	t.Run("a plan starting tomorrow is not", func(t *testing.T) {
		tomorrow, old := planStarting("2026-10-07"), planStarting("2026-07-01")
		assert.Equal(t, old.ID, currentPlanID([]Plan{tomorrow, old}, now))
	})

	t.Run("same start date goes to the first listed", func(t *testing.T) {
		later, earlier := planStarting("2026-09-15"), planStarting("2026-09-15")
		assert.Equal(t, later.ID, currentPlanID([]Plan{later, earlier}, now))
	})

	t.Run("none when every plan is in the future", func(t *testing.T) {
		assert.Equal(t, uuid.Nil, currentPlanID([]Plan{planStarting("2026-12-01"), planStarting("2026-11-01")}, now))
	})

	t.Run("none when there are no plans", func(t *testing.T) {
		assert.Equal(t, uuid.Nil, currentPlanID(nil, now))
	})
}
