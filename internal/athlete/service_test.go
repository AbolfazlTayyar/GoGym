package athlete

import (
	"strings"
	"testing"

	"github.com/AbolfazlTayyar/gogym/internal/httpx"
	"github.com/AbolfazlTayyar/gogym/internal/validate"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strptr(s string) *string   { return &s }
func fltptr(f float64) *float64 { return &f }

func validInput() CreateInput {
	return CreateInput{
		FirstName: "Sara",
		LastName:  "Ahmadi",
		Phone:     "09372144430",
	}
}

func requireValidationFields(t *testing.T, err error) map[string]string {
	t.Helper()

	var verr *ValidationError
	require.ErrorAs(t, err, &verr)

	return verr.Fields
}

func TestBuildAthlete_Valid(t *testing.T) {
	coachID := uuid.New()

	input := validInput()
	input.ExperienceLevel = strptr("Intermediate")
	input.Injuries = strptr("left shoulder")
	input.Goal = strptr("lose 5kg")
	input.Height = fltptr(172)
	input.AthleteType = TypePublic

	a, err := buildAthlete(coachID, input)
	require.NoError(t, err)

	assert.Equal(t, coachID, a.CoachID, "the owning coach comes from the caller, never the input")
	assert.NotEqual(t, uuid.Nil, a.ID)
	assert.Equal(t, "Sara", a.FirstName)
	assert.Equal(t, "09372144430", a.Phone)
	assert.Equal(t, TypePublic, a.AthleteType)
	require.NotNil(t, a.ExperienceLevel)
	assert.Equal(t, ExperienceIntermediate, *a.ExperienceLevel, "experience level is stored lowercased")
	require.NotNil(t, a.Height)
	assert.Equal(t, float64(172), *a.Height)
}

func TestBuildAthlete_ReportsEveryBadFieldAtOnce(t *testing.T) {
	_, err := buildAthlete(uuid.New(), CreateInput{
		FirstName:       "  ",
		Phone:           "+15550001111",
		ExperienceLevel: strptr("expert"),
		Height:          fltptr(1.72),
		AthleteType:     "vip",
	})

	fields := requireValidationFields(t, err)
	assert.Equal(t, httpx.MsgFieldRequired, fields["first_name"], "whitespace-only is not a name")
	assert.Equal(t, httpx.MsgFieldRequired, fields["last_name"])
	assert.Equal(t, validate.MsgIranMobile, fields["phone"])
	assert.Contains(t, fields["experience_level"], ExperienceBeginner)
	assert.Contains(t, fields["height"], "centimetres", "metres instead of centimetres is caught")
	assert.Contains(t, fields["athlete_type"], TypePrivate)
	assert.Len(t, fields, 6)
}

func TestBuildAthlete_MissingPhoneIsRequiredNotMalformed(t *testing.T) {
	input := validInput()
	input.Phone = "   "

	_, err := buildAthlete(uuid.New(), input)

	fields := requireValidationFields(t, err)
	assert.Equal(t, httpx.MsgFieldRequired, fields["phone"])
}

func TestBuildAthlete_DefaultsToPrivate(t *testing.T) {
	a, err := buildAthlete(uuid.New(), validInput())
	require.NoError(t, err)

	assert.Equal(t, TypePrivate, a.AthleteType)
}

func TestBuildAthlete_BlankOptionalsBecomeNull(t *testing.T) {
	input := validInput()
	input.ExperienceLevel = strptr("")
	input.Injuries = strptr("   ")
	input.Goal = strptr("")

	a, err := buildAthlete(uuid.New(), input)
	require.NoError(t, err)

	assert.Nil(t, a.ExperienceLevel)
	assert.Nil(t, a.Injuries)
	assert.Nil(t, a.Goal)
}

func TestBuildAthlete_HeightBounds(t *testing.T) {
	for name, tc := range map[string]struct {
		height  float64
		wantErr bool
	}{
		"metres not centimetres": {height: 1.72, wantErr: true},
		"below minimum":          {height: minHeightCM - 1, wantErr: true},
		"at minimum":             {height: minHeightCM, wantErr: false},
		"typical":                {height: 172, wantErr: false},
		"at maximum":             {height: maxHeightCM, wantErr: false},
		"above maximum":          {height: maxHeightCM + 1, wantErr: true},
	} {
		t.Run(name, func(t *testing.T) {
			input := validInput()
			input.Height = fltptr(tc.height)

			_, err := buildAthlete(uuid.New(), input)
			if !tc.wantErr {
				require.NoError(t, err)
				return
			}

			fields := requireValidationFields(t, err)
			assert.Contains(t, fields, "height")
		})
	}
}

func TestBuildAthlete_NameLengthCountsRunes(t *testing.T) {
	persian := strings.Repeat("ش", maxNameLength)
	require.Greater(t, len(persian), maxNameLength, "the test name is longer in bytes than in runes")

	input := validInput()
	input.FirstName = persian

	a, err := buildAthlete(uuid.New(), input)
	require.NoError(t, err)
	assert.Equal(t, persian, a.FirstName)

	input.FirstName = persian + "ش"
	_, err = buildAthlete(uuid.New(), input)

	fields := requireValidationFields(t, err)
	assert.Contains(t, fields, "first_name")
}

func TestBuildAthlete_NotesLengthBounded(t *testing.T) {
	input := validInput()
	input.Goal = strptr(strings.Repeat("a", maxNotesLength+1))

	_, err := buildAthlete(uuid.New(), input)

	fields := requireValidationFields(t, err)
	assert.Contains(t, fields, "goal")
}

func TestNormalizeListOptions_TypeFilter(t *testing.T) {
	for name, tc := range map[string]struct {
		in   string
		want string
	}{
		"unset defaults to the dashboard's working list": {in: "", want: TypePrivate},
		"explicit private":              {in: TypePrivate, want: TypePrivate},
		"explicit public":               {in: TypePublic, want: TypePublic},
		"all":                           {in: TypeFilterAll, want: TypeFilterAll},
		"case and padding are forgiven": {in: "  PUBLIC ", want: TypePublic},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := normalizeListOptions(ListOptions{Type: tc.in})
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.Type)
		})
	}
}

func TestNormalizeListOptions_RejectsUnknownType(t *testing.T) {
	_, err := normalizeListOptions(ListOptions{Type: "everyone"})

	fields := requireValidationFields(t, err)
	assert.Contains(t, fields, "athlete_type")
}

func TestNormalizeListOptions_PageWindow(t *testing.T) {
	for name, tc := range map[string]struct {
		limit, offset         int
		wantLimit, wantOffset int
	}{
		"unset":           {limit: 0, offset: 0, wantLimit: defaultLimit, wantOffset: 0},
		"negative":        {limit: -5, offset: -5, wantLimit: defaultLimit, wantOffset: 0},
		"within bounds":   {limit: 50, offset: 40, wantLimit: 50, wantOffset: 40},
		"clamped to max":  {limit: maxLimit + 1, offset: 0, wantLimit: maxLimit, wantOffset: 0},
		"exactly the max": {limit: maxLimit, offset: 0, wantLimit: maxLimit, wantOffset: 0},
		"smallest useful": {limit: 1, offset: 1, wantLimit: 1, wantOffset: 1},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := normalizeListOptions(ListOptions{Limit: tc.limit, Offset: tc.offset})
			require.NoError(t, err)
			assert.Equal(t, tc.wantLimit, got.Limit)
			assert.Equal(t, tc.wantOffset, got.Offset)
		})
	}
}

func TestNormalizeListOptions_TrimsQuery(t *testing.T) {
	got, err := normalizeListOptions(ListOptions{Query: "  ada  "})
	require.NoError(t, err)

	assert.Equal(t, "ada", got.Query)
}
