package measurement

import (
	"testing"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fltptr(f float64) *float64 { return &f }

func requireValidationFields(t *testing.T, err error) map[string]string {
	t.Helper()

	var verr *ValidationError
	require.ErrorAs(t, err, &verr)

	return verr.Fields
}

func TestBuildMeasurement_Valid(t *testing.T) {
	athleteID := uuid.New()

	m, err := buildMeasurement(athleteID, CreateInput{
		Date:   " 2026-09-24 ",
		Weight: fltptr(72.5),
		Waist:  fltptr(81),
	})
	require.NoError(t, err)

	assert.NotEqual(t, uuid.Nil, m.ID)
	assert.Equal(t, athleteID, m.AthleteID)
	assert.Equal(t, time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC), m.Date)
	require.NotNil(t, m.Weight)
	assert.InDelta(t, 72.5, *m.Weight, 0)
	assert.Nil(t, m.Chest, "an unrecorded metric stays null, not 0")
}

func TestBuildMeasurement_Date(t *testing.T) {
	for name, tc := range map[string]struct {
		date    string
		wantMsg string
	}{
		"missing":              {date: "", wantMsg: httpx.MsgFieldRequired},
		"whitespace":           {date: "   ", wantMsg: httpx.MsgFieldRequired},
		"timestamp not a date": {date: "2026-09-24T09:30:00Z", wantMsg: httpx.MsgDateFormat},
		"day first":            {date: "24-09-2026", wantMsg: httpx.MsgDateFormat},
		"impossible day":       {date: "2026-02-30", wantMsg: httpx.MsgDateFormat},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := buildMeasurement(uuid.New(), CreateInput{Date: tc.date, Weight: fltptr(70)})

			fields := requireValidationFields(t, err)
			assert.Equal(t, tc.wantMsg, fields["date"])
			assert.Len(t, fields, 1)
		})
	}
}

func TestBuildMeasurement_RequiresAtLeastOneValue(t *testing.T) {
	_, err := buildMeasurement(uuid.New(), CreateInput{Date: "2026-09-24"})

	fields := requireValidationFields(t, err)
	for _, name := range []string{"weight", "chest", "waist", "arm", "thigh", "hip"} {
		assert.Equal(t, msgNoValues, fields[name], name)
	}
}

func TestBuildMeasurement_Bounds(t *testing.T) {
	for name, tc := range map[string]struct {
		input     CreateInput
		wantField string
	}{
		"weight in grams":    {input: CreateInput{Weight: fltptr(72500)}, wantField: "weight"},
		"weight too low":     {input: CreateInput{Weight: fltptr(minWeightKG - 1)}, wantField: "weight"},
		"waist in metres":    {input: CreateInput{Waist: fltptr(0.81)}, wantField: "waist"},
		"negative arm":       {input: CreateInput{Arm: fltptr(-34)}, wantField: "arm"},
		"chest too large":    {input: CreateInput{Chest: fltptr(maxCircumferenceCM + 1)}, wantField: "chest"},
		"thigh in metres":    {input: CreateInput{Thigh: fltptr(0.56)}, wantField: "thigh"},
		"hip in millimetres": {input: CreateInput{Hip: fltptr(990)}, wantField: "hip"},
	} {
		t.Run(name, func(t *testing.T) {
			tc.input.Date = "2026-09-24"

			_, err := buildMeasurement(uuid.New(), tc.input)

			fields := requireValidationFields(t, err)
			assert.Contains(t, fields, tc.wantField)
			assert.Len(t, fields, 1)
		})
	}
}

func TestBuildMeasurement_BoundsAreInclusive(t *testing.T) {
	_, err := buildMeasurement(uuid.New(), CreateInput{
		Date:   "2026-09-24",
		Weight: fltptr(minWeightKG),
		Chest:  fltptr(maxCircumferenceCM),
		Arm:    fltptr(minCircumferenceCM),
	})

	require.NoError(t, err)
}

func TestBuildMeasurement_ReportsEveryBadFieldAtOnce(t *testing.T) {
	_, err := buildMeasurement(uuid.New(), CreateInput{
		Date:   "yesterday",
		Weight: fltptr(1),
		Waist:  fltptr(0.8),
	})

	fields := requireValidationFields(t, err)
	assert.Equal(t, httpx.MsgDateFormat, fields["date"])
	assert.Contains(t, fields["weight"], unitKilograms)
	assert.Contains(t, fields["waist"], unitCentimetres)
	assert.Len(t, fields, 3)
}
