package duration

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Duration holds all the smaller units that make up the duration
type Duration struct {
	Years    float64
	Months   float64
	Weeks    float64
	Days     float64
	Hours    float64
	Minutes  float64
	Seconds  float64
	Negative bool
}

const (
	parsingPeriod = iota
	parsingTime

	hoursPerDay   = 24
	hoursPerWeek  = hoursPerDay * 7
	hoursPerMonth = hoursPerYear / 12
	hoursPerYear  = hoursPerDay * 365

	nsPerSecond = 1000000000
	nsPerMinute = nsPerSecond * 60
	nsPerHour   = nsPerMinute * 60
	nsPerDay    = nsPerHour * hoursPerDay
	nsPerWeek   = nsPerHour * hoursPerWeek
	nsPerMonth  = nsPerHour * hoursPerMonth
	nsPerYear   = nsPerHour * hoursPerYear
)

var (
	// ErrUnexpectedInput is returned when an input in the duration string does not match expectations
	ErrUnexpectedInput = errors.New("unexpected input")
	// ErrIncompleteExpr is returned when the expression is incomplete, (e.g. missing unit).
	ErrIncompleteExpr = errors.New("incomplete expression")
)

// Parse attempts to parse the given duration string into a *Duration,
// if parsing fails an error is returned instead.
func Parse(d string) (*Duration, error) {
	state := parsingPeriod
	duration := &Duration{}
	num := ""
	var err error

	switch {
	case strings.HasPrefix(d, "P"): // standard duration
	case strings.HasPrefix(d, "-P"): // negative duration
		duration.Negative = true
		d = strings.TrimPrefix(d, "-") // remove the negative sign
	default:
		return nil, ErrUnexpectedInput
	}

	for _, char := range d {
		switch char {
		case 'P':
			if state != parsingPeriod {
				return nil, ErrUnexpectedInput
			}
		case 'T':
			state = parsingTime
		case 'Y':
			if state != parsingPeriod {
				return nil, ErrUnexpectedInput
			}

			duration.Years, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return nil, err
			}
			num = ""
		case 'M':
			if state == parsingPeriod {
				duration.Months, err = strconv.ParseFloat(num, 64)
				if err != nil {
					return nil, err
				}
				num = ""
			} else if state == parsingTime {
				duration.Minutes, err = strconv.ParseFloat(num, 64)
				if err != nil {
					return nil, err
				}
				num = ""
			}
		case 'W':
			if state != parsingPeriod {
				return nil, ErrUnexpectedInput
			}

			duration.Weeks, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return nil, err
			}
			num = ""
		case 'D':
			if state != parsingPeriod {
				return nil, ErrUnexpectedInput
			}

			duration.Days, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return nil, err
			}
			num = ""
		case 'H':
			if state != parsingTime {
				return nil, ErrUnexpectedInput
			}

			duration.Hours, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return nil, err
			}
			num = ""
		case 'S':
			if state != parsingTime {
				return nil, ErrUnexpectedInput
			}

			duration.Seconds, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return nil, err
			}
			num = ""
		default:
			if unicode.IsNumber(char) || char == '.' {
				num += string(char)
				continue
			}

			return nil, ErrUnexpectedInput
		}
	}
	if num != "" {
		return nil, ErrIncompleteExpr
	}

	return duration, nil
}

// FromTimeDuration converts a time.Duration into a *Duration using exact fixed-length
// units (Weeks, Days, Hours, Minutes, Seconds). Never produces Years or Months.
func FromTimeDuration(d time.Duration) *Duration {
	duration := &Duration{}
	if d == 0 {
		return duration
	}

	if d < 0 {
		d = -d
		duration.Negative = true
	}

	if d >= nsPerWeek {
		duration.Weeks = float64(d / nsPerWeek)
		d -= time.Duration(duration.Weeks) * nsPerWeek
	}
	if d >= nsPerDay {
		duration.Days = float64(d / nsPerDay)
		d -= time.Duration(duration.Days) * nsPerDay
	}
	if d >= nsPerHour {
		duration.Hours = float64(d / nsPerHour)
		d -= time.Duration(duration.Hours) * nsPerHour
	}
	if d >= nsPerMinute {
		duration.Minutes = float64(d / nsPerMinute)
		d -= time.Duration(duration.Minutes) * nsPerMinute
	}
	duration.Seconds = d.Seconds()

	return duration
}

// Format formats a time.Duration into an ISO 8601 string (e.g. P1DT6H5M).
// Negative durations are prefixed with "-"; zero returns "PT0S".
// Output never contains Years or Months; see FromTimeDuration.
func Format(d time.Duration) string {
	return FromTimeDuration(d).String()
}

// ToTimeDuration converts the *Duration to time.Duration.
// Durations with Years or Months use fixed-length approximations and may be inexact.
//
// Deprecated: Use ToTimeDurationFrom for exact results when Years or Months are set.
func (duration *Duration) ToTimeDuration() time.Duration {
	var timeDuration time.Duration

	// zero checks are here to avoid unnecessary math operations, on a duration such as `PT5M`
	if duration.Years != 0 {
		timeDuration += time.Duration(math.Round(duration.Years * nsPerYear))
	}
	if duration.Months != 0 {
		timeDuration += time.Duration(math.Round(duration.Months * nsPerMonth))
	}
	if duration.Weeks != 0 {
		timeDuration += time.Duration(math.Round(duration.Weeks * nsPerWeek))
	}
	if duration.Days != 0 {
		timeDuration += time.Duration(math.Round(duration.Days * nsPerDay))
	}
	if duration.Hours != 0 {
		timeDuration += time.Duration(math.Round(duration.Hours * nsPerHour))
	}
	if duration.Minutes != 0 {
		timeDuration += time.Duration(math.Round(duration.Minutes * nsPerMinute))
	}
	if duration.Seconds != 0 {
		timeDuration += time.Duration(math.Round(duration.Seconds * nsPerSecond))
	}
	if duration.Negative {
		timeDuration = -timeDuration
	}

	return timeDuration
}

// String returns the ISO8601 duration string for the *Duration
func (duration *Duration) String() string {
	d := "P"
	hasTime := false

	appendD := func(designator string, value float64, isTime bool) {
		if !hasTime && isTime {
			d += "T"
			hasTime = true
		}

		d += strconv.FormatFloat(value, 'f', -1, 64) + designator
	}

	if duration.Years != 0 {
		appendD("Y", duration.Years, false)
	}

	if duration.Months != 0 {
		appendD("M", duration.Months, false)
	}

	if duration.Weeks != 0 {
		appendD("W", duration.Weeks, false)
	}

	if duration.Days != 0 {
		appendD("D", duration.Days, false)
	}

	if duration.Hours != 0 {
		appendD("H", duration.Hours, true)
	}

	if duration.Minutes != 0 {
		appendD("M", duration.Minutes, true)
	}

	if duration.Seconds != 0 {
		appendD("S", duration.Seconds, true)
	}

	// if the duration is zero, return "PT0S"
	if d == "P" {
		d += "T0S"
	}

	if duration.Negative {
		return "-" + d
	}

	return d
}

// MarshalJSON satisfies the Marshaler interface by return a valid JSON string representation of the duration
func (duration Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(duration.String())
}

// UnmarshalJSON satisfies the Unmarshaler interface by return a valid JSON string representation of the duration
func (duration *Duration) UnmarshalJSON(source []byte) error {
	durationString := ""
	err := json.Unmarshal(source, &durationString)
	if err != nil {
		return err
	}

	parsed, err := Parse(durationString)
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}

	*duration = *parsed
	return nil
}

// MarshalText implements [encoding.TextMarshaler].
func (duration Duration) MarshalText() ([]byte, error) {
	return []byte(duration.String()), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler].
func (duration *Duration) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}

	*duration = *parsed
	return nil
}
