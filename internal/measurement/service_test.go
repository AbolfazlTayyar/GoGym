package measurement

import (
	"testing"
	"time"

	"github.com/AbolfazlTayyar/gogym/internal/platform/httpx"
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

// apply runs applyInput on an empty measurement, for tests that only look at the error.
func apply(input Input) error {
	return applyInput(&Measurement{}, input)
}

func TestApplyInput_Valid(t *testing.T) {
	m := &Measurement{}

	err := applyInput(m, Input{
		Date:   " 2026-09-24 ",
		Weight: fltptr(72.5),
		Waist:  fltptr(81),
	})
	require.NoError(t, err)

	assert.Equal(t, time.Date(2026, 9, 24, 0, 0, 0, 0, time.UTC), m.Date)
	require.NotNil(t, m.Weight)
	assert.InDelta(t, 72.5, *m.Weight, 0)
	assert.Nil(t, m.Chest, "an unrecorded metric stays null, not 0")
}

func TestApplyInput_ReplacesEveryMetric(t *testing.T) {
	m := &Measurement{Weight: fltptr(74), Chest: fltptr(99)}

	require.NoError(t, applyInput(m, Input{Date: "2026-09-24", Weight: fltptr(73)}))

	require.NotNil(t, m.Weight)
	assert.InDelta(t, 73, *m.Weight, 0)
	assert.Nil(t, m.Chest, "a metric left out of an update is cleared")
}

func TestApplyInput_InvalidInputLeavesMeasurementUntouched(t *testing.T) {
	date := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	m := &Measurement{Date: date, Weight: fltptr(74)}

	err := applyInput(m, Input{Date: "2026-09-24", Weight: fltptr(74), Waist: fltptr(0.8)})
	requireValidationFields(t, err)

	assert.Equal(t, date, m.Date)
	require.NotNil(t, m.Weight)
	assert.InDelta(t, 74, *m.Weight, 0)
	assert.Nil(t, m.Waist)
}

func TestApplyInput_Date(t *testing.T) {
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
			err := apply(Input{Date: tc.date, Weight: fltptr(70)})

			fields := requireValidationFields(t, err)
			assert.Equal(t, tc.wantMsg, fields["date"])
			assert.Len(t, fields, 1)
		})
	}
}

func TestApplyInput_RequiresAtLeastOneValue(t *testing.T) {
	err := apply(Input{Date: "2026-09-24"})

	fields := requireValidationFields(t, err)
	for _, name := range []string{"weight", "chest", "waist", "arm", "thigh", "hip"} {
		assert.Equal(t, msgNoValues, fields[name], name)
	}
}

func TestApplyInput_Bounds(t *testing.T) {
	for name, tc := range map[string]struct {
		input     Input
		wantField string
	}{
		"weight in grams":    {input: Input{Weight: fltptr(72500)}, wantField: "weight"},
		"weight too low":     {input: Input{Weight: fltptr(minWeightKG - 1)}, wantField: "weight"},
		"waist in metres":    {input: Input{Waist: fltptr(0.81)}, wantField: "waist"},
		"negative arm":       {input: Input{Arm: fltptr(-34)}, wantField: "arm"},
		"chest too large":    {input: Input{Chest: fltptr(maxCircumferenceCM + 1)}, wantField: "chest"},
		"thigh in metres":    {input: Input{Thigh: fltptr(0.56)}, wantField: "thigh"},
		"hip in millimetres": {input: Input{Hip: fltptr(990)}, wantField: "hip"},
	} {
		t.Run(name, func(t *testing.T) {
			tc.input.Date = "2026-09-24"

			err := apply(tc.input)

			fields := requireValidationFields(t, err)
			assert.Contains(t, fields, tc.wantField)
			assert.Len(t, fields, 1)
		})
	}
}

func TestApplyInput_BoundsAreInclusive(t *testing.T) {
	err := apply(Input{
		Date:   "2026-09-24",
		Weight: fltptr(minWeightKG),
		Chest:  fltptr(maxCircumferenceCM),
		Arm:    fltptr(minCircumferenceCM),
	})

	require.NoError(t, err)
}

func TestApplyInput_ReportsEveryBadFieldAtOnce(t *testing.T) {
	err := apply(Input{
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
