package main

import (
	"fmt"
	"strings"
	"time"
)

type Ord interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

type Histogram[T Ord] struct {
	buckets  []uint64
	limiters []T
}

func NewHistogram[T Ord](limiters []T) *Histogram[T] {
	for i, v := range limiters {
		if i > 0 {
			pv := limiters[i-1]
			if pv >= v {
				panic("unsorted limiters")
			}
		}
	}
	return &Histogram[T]{
		buckets:  make([]uint64, len(limiters)+1),
		limiters: limiters,
	}
}

func (h *Histogram[T]) Add(t T) {
	for i, v := range h.limiters {
		if t < v {
			h.buckets[i]++
		}
	}
	h.buckets[len(h.buckets)-1]++
}

func (h *Histogram[T]) Describe() string {
	var sb strings.Builder

	var mx uint64
	for i, v := range h.buckets {
		if i > 0 {
			b := v - h.buckets[i-1]
			if b > mx {
				mx = b
			}
		} else {
			mx = v
		}
	}

	const w = 16
	var prev T
	var pb uint64

	for i, v := range h.limiters {
		fmt.Fprintf(&sb, "[%06v, %06v): ", prev, v)
		prev = v
		buckVal := h.buckets[i] - pb
		nx := float32(w) / float32(mx) * float32(buckVal)
		for range int(nx) {
			fmt.Fprint(&sb, "#")
		}
		fmt.Fprintln(&sb, "\t", buckVal)
		pb = h.buckets[i]
	}

	fmt.Fprintf(&sb, "[%06v,   +Inf): ", prev)
	buckVal := h.buckets[len(h.buckets)-1] - pb
	nx := float32(w) / float32(mx) * float32(buckVal)
	for range int(nx) {
		fmt.Fprint(&sb, "#")
	}
	fmt.Fprintln(&sb, "\t", buckVal)
	return sb.String()
}

type statItem struct {
	samples int
	delayed int
	max     time.Duration
	min     time.Duration
	hist    *Histogram[time.Duration]
}

type stats struct {
	global statItem
	azure  statItem
	pa     statItem

	curTime time.Time
}

func NewStats() *stats {
	return &stats{
		global:  newStatItem(),
		azure:   newStatItem(),
		pa:      newStatItem(),
		curTime: time.Time{},
	}
}

func newStatItem() statItem {
	return statItem{
		samples: 0,
		delayed: 0,
		max:     0,
		min:     0,
		hist: NewHistogram[time.Duration]([]time.Duration{
			time.Second,
			5 * time.Second,
			20 * time.Second,
			60 * time.Second,
			5 * time.Minute,
			10 * time.Minute,
			15 * time.Minute,
			30 * time.Minute,
			1 * time.Hour,
			2 * time.Hour,
		}),
	}
}

func (s *stats) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%v: %v\n", "Global", s.global.String())
	fmt.Fprintf(&sb, "%v: %v\n", "Azure", s.azure.String())
	fmt.Fprintf(&sb, "%v: %v\n", "   PA", s.pa.String())
	return sb.String()
}

func (s *stats) process(t time.Time, pa bool) {

	if s.curTime.IsZero() {
		s.curTime = t
	}

	delay := s.curTime.Sub(t)
	s.global.process(delay)
	if pa {
		s.pa.process(delay)
	} else {
		s.azure.process(delay)
	}
	if delay <= 0 {
		s.curTime = t
	}
}

func (s *statItem) String() string {
	return fmt.Sprintf("total: %09d, delayed: %09d (%g%%), min: %v, max: %v\n%v",
		s.samples, s.delayed, 100.0*float32(s.delayed)/float32(s.samples), s.min, s.max, s.hist.Describe())
}

func (s *statItem) process(delay time.Duration) {
	s.samples++
	if delay <= 0 {
		return
	}
	s.delayed++
	if delay > s.max {
		s.max = delay
	}
	if s.min == 0 || s.min > delay {
		s.min = delay
	}
	s.hist.Add(delay)
}
