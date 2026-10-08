package movement

import (
	"strings"
	"testing"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strptr(s string) *string { return &s }

func requireValidationFields(t *testing.T, err error) map[string]string {
	t.Helper()

	var verr *ValidationError
	require.ErrorAs(t, err, &verr)

	return verr.Fields
}

func TestApplyInput_Valid(t *testing.T) {
	var m Movement
	err := applyInput(&m, Input{
		Name:        "  Bulgarian split squat ",
		Category:    strptr("strength"),
		Description: strptr("Rear foot on a bench"),
		MuscleGroup: strptr("legs"),
		Equipment:   strptr("dumbbell"),
	})
	require.NoError(t, err)

	assert.Equal(t, "Bulgarian split squat", m.Name)
	assert.Equal(t, strptr("strength"), m.Category)
	assert.Equal(t, strptr("Rear foot on a bench"), m.Description)
	assert.Equal(t, strptr("legs"), m.MuscleGroup)
	assert.Equal(t, strptr("dumbbell"), m.Equipment)
}

func TestApplyInput_ReportsEveryBadFieldAtOnce(t *testing.T) {
	err := applyInput(&Movement{}, Input{
		Name:        " ",
		Category:    strptr(strings.Repeat("a", maxCategoryLength+1)),
		Description: strptr(strings.Repeat("a", maxDescriptionLength+1)),
		MuscleGroup: strptr("quads"),
		Equipment:   strptr("sled"),
	})

	fields := requireValidationFields(t, err)
	assert.Equal(t, httpx.MsgFieldRequired, fields["name"])
	assert.Contains(t, fields, "category")
	assert.Contains(t, fields, "description")
	assert.Equal(t, "must be one of chest, back, shoulders, arms, legs, core, full_body", fields[fieldMuscleGroup])
	assert.Contains(t, fields[fieldEquipment], "must be one of bodyweight, barbell")
}

func TestApplyInput_LeavesMovementUntouchedOnError(t *testing.T) {
	m := Movement{Name: "Back squat", MuscleGroup: strptr("legs")}

	err := applyInput(&m, Input{Name: "Front squat", MuscleGroup: strptr("quads")})
	require.Error(t, err)

	assert.Equal(t, "Back squat", m.Name, "a half-valid update mustn't leak into the movement")
	assert.Equal(t, strptr("legs"), m.MuscleGroup)
}

func TestApplyInput_BlankOptionalsBecomeNull(t *testing.T) {
	m := Movement{Category: strptr("strength"), MuscleGroup: strptr("legs")}

	err := applyInput(&m, Input{
		Name:        "Back squat",
		Category:    strptr("  "),
		MuscleGroup: strptr(""),
	})
	require.NoError(t, err)

	assert.Nil(t, m.Category)
	assert.Nil(t, m.Description, "an omitted field is cleared: an update is a full replacement")
	assert.Nil(t, m.MuscleGroup)
	assert.Nil(t, m.Equipment)
}

func TestApplyInput_ChoicesIgnoreCaseAndPadding(t *testing.T) {
	var m Movement
	err := applyInput(&m, Input{Name: "Leg press", MuscleGroup: strptr(" Legs "), Equipment: strptr("MACHINE")})
	require.NoError(t, err)

	assert.Equal(t, strptr("legs"), m.MuscleGroup)
	assert.Equal(t, strptr("machine"), m.Equipment)
}

func TestApplyInput_NameLengthCountsRunes(t *testing.T) {
	persian := strings.Repeat("س", maxNameLength)

	var m Movement
	require.NoError(t, applyInput(&m, Input{Name: persian}), "%d two-byte runes fit", maxNameLength)

	fields := requireValidationFields(t, applyInput(&m, Input{Name: persian + "س"}))
	assert.Contains(t, fields, "name")
}

func TestApplyInput_NeverSetsOwner(t *testing.T) {
	coachID := uuid.New()
	m := Movement{CoachID: &coachID}

	require.NoError(t, applyInput(&m, Input{Name: "Back squat"}))

	assert.Equal(t, &coachID, m.CoachID)
}

func TestNormalizeListOptions(t *testing.T) {
	got, err := normalizeListOptions(ListOptions{Query: "  squat ", MuscleGroup: "LEGS", Equipment: " barbell"})
	require.NoError(t, err)

	assert.Equal(t, ListOptions{Query: "squat", MuscleGroup: "legs", Equipment: "barbell"}, got)
}

func TestNormalizeListOptions_EmptyIsNoFilter(t *testing.T) {
	got, err := normalizeListOptions(ListOptions{})
	require.NoError(t, err)

	assert.Equal(t, ListOptions{}, got)
}

func TestNormalizeListOptions_RejectsUnknownFilterValues(t *testing.T) {
	_, err := normalizeListOptions(ListOptions{MuscleGroup: "quads", Equipment: "sled"})

	fields := requireValidationFields(t, err)
	assert.Contains(t, fields, fieldMuscleGroup)
	assert.Contains(t, fields, fieldEquipment)
}
