"""Distribution helpers for producing realistic engagement data.

Every function that uses randomness takes an explicit `random.Random`
instance (typically `Faker.random`). This keeps determinism at the call
site — the caller controls the seed, the module just does the math.

Four families of helpers:

1. `zipf_distribution`        — engagement counts (views, vote totals)
2. `long_tail_count`          — per-entity activity counts
3. `backdated_timestamp`      — academic-calendar-weighted timestamps
4. `vote_split`               — allocate a vote budget into up/down/cancel

The academic-calendar weights in `DEFAULT_MONTH_WEIGHTS` are opinionated:
December + May peak (exam weeks), August-November + February-April are
steady, June-July are low (summer). Override via the `weights` kwarg if
a non-academic product uses these helpers later.
"""

from __future__ import annotations

import random
from dataclasses import dataclass
from datetime import UTC, datetime, timedelta

# ---------------------------------------------------------------------------
# Academic-calendar month weights (sum = 12.00)
# ---------------------------------------------------------------------------

# Relative weights per calendar month. Pattern: low in June-July (summer),
# steady during semester weeks, peak in May + December (exam weeks).
DEFAULT_MONTH_WEIGHTS: tuple[float, ...] = (
    1.10,  # Jan — post-break ramp-up
    1.00,  # Feb
    1.00,  # Mar
    1.00,  # Apr
    1.50,  # May — spring exam peak
    0.50,  # Jun — summer low
    0.45,  # Jul — summer low
    1.20,  # Aug — fall ramp-up
    1.00,  # Sep
    1.00,  # Oct
    1.10,  # Nov
    1.65,  # Dec — fall exam peak
)


# ---------------------------------------------------------------------------
# Zipf (engagement counts)
# ---------------------------------------------------------------------------


def zipf_distribution(n: int, s: float, total: int, floor: int = 0) -> list[int]:
    """Return n integers summing to ~total, following Zipf's law.

    Exponent `s` controls skew:
      - 1.0 → classic Zipf (long tail)
      - 1.2 → moderately skewed (top 5 of 82 hold ~55% of mass with floor=5)
      - 2.0 → severely skewed (top item dominates)

    Each returned value is >= floor.

    Pure function — deterministic given inputs. No RNG; the caller
    decides whether to `rng.shuffle()` the result to decorrelate from
    insert order.
    """
    if n <= 0:
        return []
    if total < floor * n:
        raise ValueError(f"total ({total}) must be >= floor ({floor}) * n ({n}) = {floor * n}")

    raw = [1.0 / ((i + 1) ** s) for i in range(n)]
    norm = sum(raw)
    # Proportional allocation with floor enforced.
    counts = [max(floor, round((total - floor * n) * r / norm) + floor) for r in raw]

    # Adjust the head to match `total` exactly — rounding errors accumulate;
    # fix them on the biggest bucket where a ±1 is invisible.
    drift = total - sum(counts)
    if drift != 0:
        counts[0] += drift

    return counts


# ---------------------------------------------------------------------------
# Long-tail per-entity counts (votes, sessions)
# ---------------------------------------------------------------------------


def long_tail_count(
    rng: random.Random, *, mean: float, floor: int = 0, cap: int | None = None
) -> int:
    """Sample a positive integer from an exponential long-tail distribution.

    Produces a mean of ~`mean` (with an exponential right tail). Use for
    things like "how many votes does this guide have" where most items
    get the mean-ish and a handful go viral.

    `floor` guarantees a minimum (prevents "this guide has 0 of X" empty
    states). `cap` truncates runaway samples if set.
    """
    if mean <= 0:
        raise ValueError(f"mean must be positive, got {mean}")
    # expovariate(lambda) has mean = 1/lambda
    sample = int(rng.expovariate(1.0 / mean))
    sample = max(floor, sample)
    if cap is not None:
        sample = min(cap, sample)
    return sample


# ---------------------------------------------------------------------------
# Backdated timestamps
# ---------------------------------------------------------------------------


@dataclass(frozen=True)
class TimeWindow:
    """A time window for backdating. Both endpoints are inclusive."""

    start: datetime
    end: datetime

    def __post_init__(self) -> None:
        if self.start >= self.end:
            raise ValueError(f"start ({self.start}) must be before end ({self.end})")


def window_ending_now(*, months_back: int = 12, now: datetime | None = None) -> TimeWindow:
    """Convenience: a `months_back`-wide window ending at `now` (default: UTC now)."""
    end = now if now is not None else datetime.now(tz=UTC)
    # Approximate month arithmetic — good enough for demo data (30.44 days avg).
    start = end - timedelta(days=int(months_back * 30.44))
    return TimeWindow(start=start, end=end)


def backdated_timestamp(
    rng: random.Random,
    *,
    window: TimeWindow,
    month_weights: tuple[float, ...] = DEFAULT_MONTH_WEIGHTS,
) -> datetime:
    """Sample a timestamp from `window`, weighted by month-of-year.

    Two-step sampling: pick a month (weighted by month_weights), then
    pick a uniform offset within that month's intersection with `window`.

    Falls back to uniform-over-window if no month intersects (e.g.
    window is narrower than a month, or weights are all zero).
    """
    if len(month_weights) != 12:
        raise ValueError(f"month_weights must have 12 entries, got {len(month_weights)}")

    # Enumerate every (month, year) pair that overlaps the window + compute
    # the overlap interval. Weight each pair by month_weights[month-1].
    overlaps: list[tuple[datetime, datetime, float]] = []
    cursor = datetime(window.start.year, window.start.month, 1, tzinfo=window.start.tzinfo)
    while cursor <= window.end:
        if cursor.month == 12:
            next_cursor = cursor.replace(year=cursor.year + 1, month=1)
        else:
            next_cursor = cursor.replace(month=cursor.month + 1)
        lo = max(cursor, window.start)
        hi = min(next_cursor, window.end)
        if lo < hi:
            weight = month_weights[cursor.month - 1]
            overlaps.append((lo, hi, weight))
        cursor = next_cursor

    if not overlaps or sum(w for _, _, w in overlaps) == 0:
        return _uniform_in(rng, window.start, window.end)

    weights = [w for _, _, w in overlaps]
    lo, hi, _ = rng.choices(overlaps, weights=weights, k=1)[0]
    return _uniform_in(rng, lo, hi)


def _uniform_in(rng: random.Random, lo: datetime, hi: datetime) -> datetime:
    span = (hi - lo).total_seconds()
    offset = rng.uniform(0, span)
    return lo + timedelta(seconds=offset)


# ---------------------------------------------------------------------------
# Vote splitting (for study_guide_votes allocation)
# ---------------------------------------------------------------------------


def vote_split(
    rng: random.Random,
    total: int,
    *,
    up_share: float = 0.80,
    down_share: float = 0.15,
) -> tuple[int, int]:
    """Split a total vote budget into (up, down) counts.

    Remaining share (~5% by default) is dropped — simulates votes that
    the user cast then cancelled. Caller discards that portion.

    Rounding rule: `up` takes the floor of its share; `down` takes the
    rounded share; the leftover (if any) is allocated to up/down with
    additional rng draws so the sum exactly matches up+down allocated.
    """
    if total < 0:
        raise ValueError(f"total must be non-negative, got {total}")
    if up_share + down_share > 1.0 + 1e-9:
        raise ValueError(f"up_share + down_share must be <= 1.0, got {up_share + down_share}")

    up_count = int(total * up_share)
    down_count = int(total * down_share)
    leftover = total - up_count - down_count
    # Small stochastic correction to avoid everything rounding the same way.
    for _ in range(leftover):
        pick = rng.random()
        if pick < 0.05:
            continue  # cancel
        elif pick < 0.05 + up_share:
            up_count += 1
        else:
            down_count += 1
    return up_count, down_count
