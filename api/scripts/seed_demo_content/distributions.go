// Distribution helpers for realistic engagement simulation.
//
// All randomness flows through an explicit *rand.Rand so the caller
// controls determinism. Seed once with rand.New(rand.NewSource(42))
// in main and re-runs produce identical activity data — same as the
// Faker.seed(42) convention in the Phase 3a Python distributions.
package main

import (
	"math"
	"math/rand"
	"time"
)

// Academic-calendar month weights (sum = 12.00). Same numbers as the
// Phase 3a Python `DEFAULT_MONTH_WEIGHTS` so re-seeding either layer
// produces the same time-distribution shape.
//
// Pattern: low in Jun-Jul (summer), steady during semester weeks,
// peak in May + December (exam weeks).
var defaultMonthWeights = [12]float64{
	1.10, // Jan — post-break ramp-up
	1.00, // Feb
	1.00, // Mar
	1.00, // Apr
	1.50, // May — spring exam peak
	0.50, // Jun — summer low
	0.45, // Jul — summer low
	1.20, // Aug — fall ramp-up
	1.00, // Sep
	1.00, // Oct
	1.10, // Nov
	1.65, // Dec — fall exam peak
}

// zipfDistribution returns n integers summing to ~total following Zipf's law.
//
// Exponent s controls skew: 1.0 classic, 1.2 moderate (top 5 of 82
// hold ~55% with floor=5), 2.0 severe. Each value >= floor.
//
// Pure function — deterministic given inputs. Caller decides whether
// to shuffle before assigning to entities (typical: shuffle so the
// "popular" buckets don't correlate with insert order).
func zipfDistribution(n int, s float64, total int, floor int) []int {
	if n <= 0 {
		return nil
	}

	raw := make([]float64, n)
	var norm float64
	for i := range n {
		raw[i] = 1.0 / math.Pow(float64(i+1), s)
		norm += raw[i]
	}

	counts := make([]int, n)
	rest := total - floor*n
	for i := range n {
		c := int(math.Round(float64(rest)*raw[i]/norm)) + floor
		if c < floor {
			c = floor
		}
		counts[i] = c
	}

	// Adjust the head so sum exactly matches total — rounding errors
	// accumulate on n items; fix on the biggest bucket where ±1 is invisible.
	var sum int
	for _, c := range counts {
		sum += c
	}
	if drift := total - sum; drift != 0 {
		counts[0] += drift
	}
	return counts
}

// longTailCount samples a positive integer from an exponential distribution
// with mean ~mean. Use for per-entity activity counts where most items get
// roughly the mean and a handful go viral.
//
// floor guarantees a minimum (prevents "guide has 0 votes" empty states).
// cap=0 means uncapped.
func longTailCount(rng *rand.Rand, mean float64, floor int, cap int) int {
	if mean <= 0 {
		return floor
	}
	sample := int(rng.ExpFloat64() * mean)
	if sample < floor {
		sample = floor
	}
	if cap > 0 && sample > cap {
		sample = cap
	}
	return sample
}

// backdatedTimestamp samples a UTC timestamp in [start, end] weighted
// by month-of-year per defaultMonthWeights.
//
// Two-step: pick a month-window weighted by its weight, then pick a
// uniform offset within that month's intersection with [start, end].
func backdatedTimestamp(rng *rand.Rand, start, end time.Time) time.Time {
	type window struct {
		lo, hi time.Time
		weight float64
	}
	var windows []window
	cur := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	for !cur.After(end) {
		next := cur.AddDate(0, 1, 0)
		lo := cur
		if start.After(lo) {
			lo = start
		}
		hi := next
		if end.Before(hi) {
			hi = end
		}
		if lo.Before(hi) {
			windows = append(windows, window{lo: lo, hi: hi, weight: defaultMonthWeights[cur.Month()-1]})
		}
		cur = next
	}
	if len(windows) == 0 {
		return uniformInWindow(rng, start, end)
	}

	var total float64
	for _, w := range windows {
		total += w.weight
	}
	if total == 0 {
		return uniformInWindow(rng, start, end)
	}
	pick := rng.Float64() * total
	var acc float64
	for _, w := range windows {
		acc += w.weight
		if pick <= acc {
			return uniformInWindow(rng, w.lo, w.hi)
		}
	}
	last := windows[len(windows)-1]
	return uniformInWindow(rng, last.lo, last.hi)
}

func uniformInWindow(rng *rand.Rand, lo, hi time.Time) time.Time {
	span := hi.Sub(lo).Seconds()
	offset := rng.Float64() * span
	return lo.Add(time.Duration(offset * float64(time.Second)))
}

// voteSplit allocates `total` votes into (up, down) with the rest dropped
// (cancel votes). Defaults: 80% up, 15% down, 5% cancel.
func voteSplit(rng *rand.Rand, total int, upShare, downShare float64) (up, down int) {
	if total <= 0 {
		return 0, 0
	}
	up = int(float64(total) * upShare)
	down = int(float64(total) * downShare)
	for range total - up - down {
		pick := rng.Float64()
		switch {
		case pick < 0.05:
			// cancel
		case pick < 0.05+upShare:
			up++
		default:
			down++
		}
	}
	return up, down
}
