package duration

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	type args struct {
		d string
	}
	var (
		noError    = func(e error) bool { return e == nil }
		newMatchFn = func(expected error) func(e error) bool {
			return func(e error) bool {
				return errors.Is(e, expected)
			}
		}
	)

	tests := []struct {
		name         string
		args         args
		want         *Duration
		errorMatchFn func(e error) bool
	}{
		{
			name:         "invalid-duration-1",
			args:         args{d: "T0S"},
			want:         nil,
			errorMatchFn: newMatchFn(ErrUnexpectedInput),
		},
		{
			name:         "invalid-duration-2",
			args:         args{d: "P-T0S"},
			want:         nil,
			errorMatchFn: newMatchFn(ErrUnexpectedInput),
		},
		{
			name:         "invalid-duration-3",
			args:         args{d: "PT0SP0D"},
			want:         nil,
			errorMatchFn: newMatchFn(ErrUnexpectedInput),
		},
		{
			name: "period-only",
			args: args{d: "P4Y"},
			want: &Duration{
				Years: 4,
			},
			errorMatchFn: noError,
		},
		{
			name: "time-only-decimal",
			args: args{d: "PT2.5S"},
			want: &Duration{
				Seconds: 2.5,
			},
			errorMatchFn: noError,
		},
		{
			name: "full",
			args: args{d: "P3Y6M4DT12H30M5.5S"},
			want: &Duration{
				Years:   3,
				Months:  6,
				Days:    4,
				Hours:   12,
				Minutes: 30,
				Seconds: 5.5,
			},
			errorMatchFn: noError,
		},
		{
			name: "negative",
			args: args{d: "-PT5M"},
			want: &Duration{
				Minutes:  5,
				Negative: true,
			},
			errorMatchFn: noError,
		},
		{
			name:         "no unit after prefix P",
			args:         args{d: "P6"},
			want:         nil,
			errorMatchFn: newMatchFn(ErrIncompleteExpr),
		},
		{
			name:         "no unit after valid sub-prefix",
			args:         args{d: "P7Y4"},
			want:         nil,
			errorMatchFn: newMatchFn(ErrIncompleteExpr),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.d)
			if !tt.errorMatchFn(err) {
				t.Errorf("error %q doesn't match the expected", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromTimeDuration(t *testing.T) {
	tests := []struct {
		give time.Duration
		want *Duration
	}{
		{
			give: 0,
			want: &Duration{},
		},
		{
			give: time.Minute * 94,
			want: &Duration{
				Hours:   1,
				Minutes: 34,
			},
		},
		{
			give: -time.Second * 10,
			want: &Duration{
				Seconds:  10,
				Negative: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.give.String(), func(t *testing.T) {
			got := FromTimeDuration(tt.give)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Format() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		give time.Duration
		want string
	}{
		{
			give: 0,
			want: "PT0S",
		},
		{
			give: time.Minute * 94,
			want: "PT1H34M",
		},
		{
			give: time.Hour * 72,
			want: "P3D",
		},
		{
			give: time.Hour * 26,
			want: "P1DT2H",
		},
		{
			give: time.Second * 465461651,
			want: "P14Y9M3DT12H54M11S",
		},
		{
			give: -time.Hour * 99544,
			want: "-P11Y4M1W4D",
		},
		{
			give: -time.Second * 10,
			want: "-PT10S",
		},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := Format(tt.give)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Format() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestDuration_ToTimeDuration(t *testing.T) {
	type fields struct {
		Years    float64
		Months   float64
		Weeks    float64
		Days     float64
		Hours    float64
		Minutes  float64
		Seconds  float64
		Negative bool
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Duration
	}{
		{
			name: "seconds",
			fields: fields{
				Seconds: 33.3,
			},
			want: time.Second*33 + time.Millisecond*300,
		},
		{
			name: "hours, minutes, and seconds",
			fields: fields{
				Hours:   2,
				Minutes: 33,
				Seconds: 17,
			},
			want: time.Hour*2 + time.Minute*33 + time.Second*17,
		},
		{
			name: "days",
			fields: fields{
				Days: 2,
			},
			want: time.Hour * 24 * 2,
		},
		{
			name: "weeks",
			fields: fields{
				Weeks: 1,
			},
			want: time.Hour * 24 * 7,
		},
		{
			name: "fractional weeks",
			fields: fields{
				Weeks: 12.5,
			},
			want: time.Hour*24*7*12 + time.Hour*84,
		},
		{
			name: "negative",
			fields: fields{
				Hours:    2,
				Negative: true,
			},
			want: -time.Hour * 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := &Duration{
				Years:    tt.fields.Years,
				Months:   tt.fields.Months,
				Weeks:    tt.fields.Weeks,
				Days:     tt.fields.Days,
				Hours:    tt.fields.Hours,
				Minutes:  tt.fields.Minutes,
				Seconds:  tt.fields.Seconds,
				Negative: tt.fields.Negative,
			}
			if got := duration.ToTimeDuration(); got != tt.want {
				t.Errorf("ToTimeDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDuration_String(t *testing.T) {
	duration, err := Parse("P3Y6M4DT12H30M5.5S")
	if err != nil {
		t.Fatal(err)
	}

	if duration.String() != "P3Y6M4DT12H30M5.5S" {
		t.Errorf("expected: %s, got: %s", "P3Y6M4DT12H30M5.5S", duration.String())
	}

	duration.Seconds = 33.3333

	if duration.String() != "P3Y6M4DT12H30M33.3333S" {
		t.Errorf("expected: %s, got: %s", "P3Y6M4DT12H30M33.3333S", duration.String())
	}

	smallDuration, err := Parse("PT0.0000000000001S")
	if err != nil {
		t.Fatal(err)
	}

	if smallDuration.String() != "PT0.0000000000001S" {
		t.Errorf("expected: %s, got: %s", "PT0.0000000000001S", smallDuration.String())
	}

	negativeDuration, err := Parse("-PT2H5M")
	if err != nil {
		t.Fatal(err)
	}

	if negativeDuration.String() != "-PT2H5M" {
		t.Errorf("expected: %s, got: %s", "-PT2H5M", negativeDuration.String())
	}
}

func TestDuration_MarshalJSON(t *testing.T) {
	td, err := Parse("P3Y6M4DT12H30M5.5S")
	if err != nil {
		t.Fatal(err)
	}

	jsonVal, err := json.Marshal(struct {
		Dur *Duration `json:"d"`
	}{Dur: td})
	if err != nil {
		t.Errorf("did not expect error: %s", err.Error())
	}
	if string(jsonVal) != `{"d":"P3Y6M4DT12H30M5.5S"}` {
		t.Errorf("expected: %s, got: %s", `{"d":"P3Y6M4DT12H30M5.5S"}`, string(jsonVal))
	}

	jsonVal, err = json.Marshal(struct {
		Dur Duration `json:"d"`
	}{Dur: *td})
	if err != nil {
		t.Errorf("did not expect error: %s", err.Error())
	}
	if string(jsonVal) != `{"d":"P3Y6M4DT12H30M5.5S"}` {
		t.Errorf("expected: %s, got: %s", `{"d":"P3Y6M4DT12H30M5.5S"}`, string(jsonVal))
	}
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	jsonStr := `
		{
			"d": "P3Y6M4DT12H30M5.5S"
		}
	`
	expected, err := Parse("P3Y6M4DT12H30M5.5S")
	if err != nil {
		t.Fatal(err)
	}

	var durStructPtr struct {
		Dur *Duration `json:"d"`
	}
	err = json.Unmarshal([]byte(jsonStr), &durStructPtr)
	if err != nil {
		t.Errorf("did not expect error: %s", err.Error())
	}
	if !reflect.DeepEqual(durStructPtr.Dur, expected) {
		t.Errorf("JSON Unmarshal ptr got = %s, want %s", durStructPtr.Dur, expected)
	}

	var durStruct struct {
		Dur Duration `json:"d"`
	}
	err = json.Unmarshal([]byte(jsonStr), &durStruct)
	if err != nil {
		t.Errorf("did not expect error: %s", err.Error())
	}
	if !reflect.DeepEqual(durStruct.Dur, *expected) {
		t.Errorf("JSON Unmarshal ptr got = %s, want %s", &(durStruct.Dur), expected)
	}
}

func TestDuration_MarshalText(t *testing.T) {
	const orig = "P3Y6M4DT12H30M5.5S"
	td, err := Parse(orig)
	if err != nil {
		t.Fatal(err)
	}

	text, err := td.MarshalText()
	if err != nil {
		t.Errorf("did not expect error: %s", err)
	}
	if string(text) != orig {
		t.Errorf("expected: %s, got: %s", orig, text)
	}
}

func TestDuration_UnmarshalText(t *testing.T) {
	const orig = `P3Y6M4DT12H30M5.5S`
	expected, err := Parse(orig)
	if err != nil {
		t.Fatal(err)
	}

	var dur Duration
	err = dur.UnmarshalText([]byte(orig))
	if err != nil {
		t.Errorf("did not expect error: %s", err.Error())
	}
	if !reflect.DeepEqual(dur, *expected) {
		t.Errorf("Text Unmarshal ptr got = %s, want %s", &dur, expected)
	}

	dur = Duration{}
	err = (&dur).UnmarshalText([]byte(orig))
	if err != nil {
		t.Errorf("did not expect error: %s", err)
	}
	if !reflect.DeepEqual(dur, *expected) {
		t.Errorf("Text Unmarshal ptr got = %s, want %s", &dur, expected)
	}
}
