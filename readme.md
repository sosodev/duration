# duration

[![Go Reference](https://pkg.go.dev/badge/github.com/sosodev/duration.svg)](https://pkg.go.dev/github.com/sosodev/duration)

It's a Go module for parsing [ISO 8601 durations](https://en.wikipedia.org/wiki/ISO_8601#Durations) and converting them to the often much more useful `time.Duration`.

## why?

ISO 8601 is a pretty common standard and sometimes these durations show up in the wild.

## installation

`go get github.com/sosodev/duration`

## [usage](https://go.dev/play/p/Nz5akjy1c6W)

```go
package main

import (
	"fmt"
	"time"
	"github.com/sosodev/duration"
)

func main() {
	d, err := duration.Parse("P3Y6M4DT12H30M5.5S")
	if err != nil {
		panic(err)
	}
	
	fmt.Println(d.Years) // 3
	fmt.Println(d.Months) // 6
	fmt.Println(d.Days) // 4
	fmt.Println(d.Hours) // 12
	fmt.Println(d.Minutes) // 30
	fmt.Println(d.Seconds) // 5.5
	
	d, err = duration.Parse("PT33.3S")
	if err != nil {
		panic(err)
	}
	
	fmt.Println(d.ToTimeDuration() == time.Second*33+time.Millisecond*300) // true

	// For durations with years or months, use Shift or ToTimeDurationFrom
	// with a reference time to get exact calendar arithmetic.
	ref := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	d, err = duration.Parse("P1Y")
	if err != nil {
		panic(err)
	}
	fmt.Println(d.Shift(ref))              // 2025-01-01 00:00:00 +0000 UTC
	fmt.Println(d.ToTimeDurationFrom(ref)) // 8784h0m0s (366 days, 2024 is a leap year)
}
```

## correctness

This module aims to implement the ISO 8601 duration specification correctly. It properly supports fractional units and has unit tests that assert the correctness of its parsing and conversions.

### `FromTimeDuration` and `Format`

`FromTimeDuration` decomposes a `time.Duration` into exact fixed-length units only (Weeks, Days, Hours, Minutes, Seconds). It never produces Years or Months, since those are calendar-dependent and cannot be derived from a plain `time.Duration` without a reference point. `Format` delegates to `FromTimeDuration`.

### `ToTimeDuration`

`ToTimeDuration` is exact for all units except Years and Months, which are approximated using fixed-length values (365-day year, 365/12-day month). It is **deprecated** for use with durations that contain Years or Months — use `ToTimeDurationFrom` instead.

### `Shift` and `ToTimeDurationFrom`

For exact calendar arithmetic, use `Shift(ref time.Time) time.Time` or `ToTimeDurationFrom(ref time.Time) time.Duration`. Both implement the ISO 8601 day-preservation rule: if adding months would produce an invalid date, the day is clamped to the last valid day of the target month — for example, January 31 + 1 month = **February 28**, not March 3.

