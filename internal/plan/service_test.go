package plan

import (
	"strings"
	"testing"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func intptr(i int) *int       { return &i }
func strptr(s string) *string { return &s }

func requireFields(t *testing.T, err error) map[string]string {
	t.Helper()

	var verr *ValidationError
	require.ErrorAs(t, err, &verr)

	return verr.Fields
}

func TestBuildPlan(t *testing.T) {
	athleteID := uuid.New()

	t.Run("valid input is trimmed and a blank note is stored as NULL", func(t *testing.T) {
		p, err := buildPlan(CreateInput{
			AthleteID: " " + athleteID.String() + " ",
			StartDate: "2026-09-15",
			Title:     "  Cut phase 1 ",
			Note:      strptr("   "),
		})
		require.NoError(t, err)

		assert.NotEqual(t, uuid.Nil, p.ID)
		assert.Equal(t, athleteID, p.AthleteID)
		assert.Equal(t, time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC), p.StartDate)
		assert.Equal(t, "Cut phase 1", p.Title)
		assert.Nil(t, p.Note)
	})

	t.Run("every missing field is reported at once", func(t *testing.T) {
		_, err := buildPlan(CreateInput{})

		assert.Equal(t, map[string]string{
			"athlete_id": httpx.MsgFieldRequired,
			"start_date": httpx.MsgFieldRequired,
			"title":      httpx.MsgFieldRequired,
		}, requireFields(t, err))
	})

	t.Run("lengths count runes, so Persian gets the same allowance", func(t *testing.T) {
		_, err := buildPlan(CreateInput{
			AthleteID: athleteID.String(),
			StartDate: "2026-09-15",
			Title:     strings.Repeat("ب", maxTitleLength),
			Note:      strptr(strings.Repeat("ن", maxNotesLength+1)),
		})

		assert.Equal(t, map[string]string{"note": "must be at most 2000 characters"}, requireFields(t, err))
	})

	t.Run("impossible date", func(t *testing.T) {
		_, err := buildPlan(CreateInput{AthleteID: athleteID.String(), StartDate: "2026-02-30", Title: "x"})
		assert.Equal(t, map[string]string{"start_date": httpx.MsgDateFormat}, requireFields(t, err))
	})
}

func TestBuildDay_Labels(t *testing.T) {
	for _, label := range []string{"A", "G", "day1", "day7", " B "} {
		t.Run("accepts "+label, func(t *testing.T) {
			d, err := buildDay(uuid.New(), AddDayInput{Label: label, OrderIndex: intptr(0)})
			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(label), d.Label)
		})
	}

	for _, label := range []string{"H", "a", "Day1", "day0", "day8", "day 1", "Monday"} {
		t.Run("rejects "+label, func(t *testing.T) {
			_, err := buildDay(uuid.New(), AddDayInput{Label: label, OrderIndex: intptr(0)})
			assert.Equal(t, map[string]string{"label": msgDayLabel}, requireFields(t, err))
		})
	}
}

func TestBuildDay_OrderIndex(t *testing.T) {
	t.Run("0 is a slot, not a missing value", func(t *testing.T) {
		d, err := buildDay(uuid.New(), AddDayInput{Label: "A", OrderIndex: intptr(0)})
		require.NoError(t, err)
		assert.Equal(t, 0, d.OrderIndex)
	})

	t.Run("the last slot is MaxDaysPerPlan-1", func(t *testing.T) {
		_, err := buildDay(uuid.New(), AddDayInput{Label: "G", OrderIndex: intptr(MaxDaysPerPlan - 1)})
		require.NoError(t, err)

		_, err = buildDay(uuid.New(), AddDayInput{Label: "G", OrderIndex: intptr(MaxDaysPerPlan)})
		assert.Equal(t, map[string]string{"order_index": "must be between 0 and 6"}, requireFields(t, err))
	})

	t.Run("missing, reported alongside a bad label", func(t *testing.T) {
		_, err := buildDay(uuid.New(), AddDayInput{Label: "Z"})
		assert.Equal(t, map[string]string{"label": msgDayLabel, "order_index": httpx.MsgFieldRequired}, requireFields(t, err))
	})
}

func TestDaySlotError_SpeaksLikeBuildDay(t *testing.T) {
	_, buildErr := buildDay(uuid.New(), AddDayInput{Label: "A", OrderIndex: intptr(MaxDaysPerPlan)})
	assert.Equal(t, requireFields(t, buildErr), requireFields(t, daySlotError(ErrDaySlotOutOfRange)),
		"the schema's range check reads exactly like the service's")

	assert.Equal(t, map[string]string{"order_index": msgDaySlotTaken}, requireFields(t, daySlotError(ErrDaySlotTaken)))

	other := assert.AnError
	assert.Same(t, other, daySlotError(other))
}

func TestBuildBlock(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		dayID := uuid.New()

		b, err := buildBlock(dayID, AddBlockInput{OrderIndex: intptr(0), Sets: intptr(3), Notes: strptr(" ")})
		require.NoError(t, err)

		assert.Equal(t, dayID, b.DayID)
		assert.Equal(t, 3, b.Sets)
		assert.Nil(t, b.RestSeconds, "rest is optional")
		assert.Nil(t, b.Notes)
	})

	t.Run("required and out-of-range fields", func(t *testing.T) {
		_, err := buildBlock(uuid.New(), AddBlockInput{Sets: intptr(0), RestSeconds: intptr(-1)})

		assert.Equal(t, map[string]string{
			"order_index":  httpx.MsgFieldRequired,
			"sets":         "must be between 1 and 50",
			"rest_seconds": "must be between 0 and 3600",
		}, requireFields(t, err))
	})

	t.Run("milliseconds sent as seconds are caught", func(t *testing.T) {
		_, err := buildBlock(uuid.New(), AddBlockInput{OrderIndex: intptr(0), Sets: intptr(3), RestSeconds: intptr(90000)})
		assert.Contains(t, requireFields(t, err), "rest_seconds")
	})
}

func TestBuildBlockMovement(t *testing.T) {
	blockID, movementID := uuid.New(), uuid.New()

	t.Run("valid", func(t *testing.T) {
		fields := map[string]string{}

		bm := buildBlockMovement(fields, 0, blockID, AddMovementInput{
			MovementID:   movementID.String(),
			Reps:         intptr(8),
			Load:         strptr(" 75% 1RM "),
			OrderInBlock: intptr(0),
		})

		assert.Empty(t, fields)
		assert.Equal(t, blockID, bm.BlockID)
		assert.Equal(t, movementID, bm.MovementID)
		require.NotNil(t, bm.Load)
		assert.Equal(t, "75% 1RM", *bm.Load, "load is free text, kept as written apart from trimming")
		assert.Nil(t, bm.DurationSeconds)
	})

	t.Run("errors are keyed by array position", func(t *testing.T) {
		fields := map[string]string{}

		buildBlockMovement(fields, 2, blockID, AddMovementInput{
			MovementID:      "not-a-uuid",
			Reps:            intptr(0),
			DurationSeconds: intptr(maxDurationSeconds + 1),
			Load:            strptr(strings.Repeat("x", maxLoadLength+1)),
		})

		assert.Equal(t, map[string]string{
			"[2].movement_id":      msgUUID,
			"[2].reps":             "must be between 1 and 1000",
			"[2].duration_seconds": "must be between 1 and 14400",
			"[2].load":             "must be at most 100 characters",
			"[2].order_in_block":   httpx.MsgFieldRequired,
		}, fields)
	})

	t.Run("a missing movement_id is required, not malformed", func(t *testing.T) {
		fields := map[string]string{}
		buildBlockMovement(fields, 0, blockID, AddMovementInput{OrderInBlock: intptr(0)})
		assert.Equal(t, map[string]string{"[0].movement_id": httpx.MsgFieldRequired}, fields)
	})
}
