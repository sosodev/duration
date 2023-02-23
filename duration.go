package duration

import (
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
	Years   float64
	Months  float64
	Weeks   float64
	Days    float64
	Hours   float64
	Minutes float64
	Seconds float64
}

const (
	parsingPeriod = iota
	parsingTime
)

var (
	// ErrUnexpectedInput is returned when an input in the duration string does not match expectations
	ErrUnexpectedInput = errors.New("unexpected input")
)

// Parse attempts to parse the given duration string into a *Duration
// if parsing fails an error is returned instead
func Parse(d string) (*Duration, error) {
	state := parsingPeriod
	duration := &Duration{}
	num := ""
	var err error

	for _, char := range d {
		switch char {
		case 'P':
			state = parsingPeriod
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

	return duration, nil
}

// Format formats the given duration into a ISO 8601 duration string (e.g. P1DT6H5M)
func Format(d time.Duration) string {
	neg := false
	if d < 0 {
		neg = true
		d = -d
	}

	if d == 0 {
		return "PT0S"
	}

	duration := &Duration{}
	if d.Hours() >= 8760 {
		duration.Years = math.Floor(d.Hours() / 8760)
		d -= time.Duration(duration.Years) * time.Hour * 8760
	}
	if d.Hours() >= 730 {
		duration.Months = math.Floor(d.Hours() / 730)
		d -= time.Duration(duration.Months) * time.Hour * 730
	}
	if d.Hours() >= 168 {
		duration.Weeks = math.Floor(d.Hours() / 168)
		d -= time.Duration(duration.Weeks) * time.Hour * 168
	}
	if d.Hours() >= 24 {
		duration.Days = math.Floor(d.Hours() / 24)
		d -= time.Duration(duration.Days) * time.Hour * 24
	}
	if d.Hours() >= 1 {
		duration.Hours = math.Floor(d.Hours())
		d -= time.Duration(duration.Hours) * time.Hour
	}
	if d.Minutes() >= 1 {
		duration.Minutes = math.Floor(d.Minutes())
		d -= time.Duration(duration.Minutes) * time.Minute
	}
	duration.Seconds = d.Seconds()

	if neg {
		return "-" + duration.String()
	}
	return duration.String()
}

// ToTimeDuration converts the *Duration to the standard library's time.Duration
// note that for *Duration's with period values of a month or year that the duration becomes a bit fuzzy
// since obviously those things vary month to month and year to year
// I used the values that Google's search provided me with as I couldn't find anything concrete on what they should be
func (duration *Duration) ToTimeDuration() time.Duration {
	var timeDuration time.Duration

	if duration.Years != 0 {
		timeDuration += time.Duration(math.Round(duration.Years * 3.154e+16))
	}
	if duration.Months != 0 {
		timeDuration += time.Duration(math.Round(duration.Months * 2.628e+15))
	}
	if duration.Weeks != 0 {
		timeDuration += time.Duration(math.Round(duration.Weeks * 6.048e+14))
	}
	if duration.Days != 0 {
		timeDuration += time.Duration(math.Round(duration.Days * 8.64e+13))
	}
	if duration.Hours != 0 {
		timeDuration += time.Duration(math.Round(duration.Hours * 3.6e+12))
	}
	if duration.Minutes != 0 {
		timeDuration += time.Duration(math.Round(duration.Minutes * 6e+10))
	}
	if duration.Seconds != 0 {
		timeDuration += time.Duration(math.Round(duration.Seconds * 1e+9))
	}

	return timeDuration
}

// String returns the ISO8601 duration string for the *Duration
func (duration *Duration) String() string {
	d := "P"

	appendD := func(designator string, value float64, isTime bool) {
		if !strings.Contains(d, "T") && isTime {
			d += "T"
		}

		d += fmt.Sprintf("%s%s", strconv.FormatFloat(value, 'f', -1, 64), designator)
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

	return d
}
