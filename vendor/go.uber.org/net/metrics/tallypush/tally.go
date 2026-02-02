// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package tallypush integrates go.uber.org/net/metrics with push-based StatsD
// and M3 systems.
package tallypush // import "go.uber.org/net/metrics/tallypush"

import (
	"math"
	"sort"

	"github.com/uber-go/tally"
	"go.uber.org/net/metrics/push"
)

// New creates a push.Target that integrates with the Tally metrics package.
// Tally supports pushing to StatsD-based systems, M3, or both. See the Tally
// documentation for details: https://godoc.org/github.com/uber-go/tally.
func New(scope tally.Scope) push.Target {
	return &target{scope}
}

type target struct {
	tally.Scope
}

func (tp *target) NewCounter(spec push.Spec) push.Counter {
	return &counter{
		Counter: tp.Tagged(spec.Tags).Counter(spec.Name),
	}
}

func (tp *target) NewGauge(spec push.Spec) push.Gauge {
	return &gauge{tp.Tagged(spec.Tags).Gauge(spec.Name)}
}

func (tp *target) NewHistogram(spec push.HistogramSpec) push.Histogram {
	buckets := make([]float64, len(spec.Buckets))
	for i := range spec.Buckets {
		if spec.Buckets[i] == math.MaxInt64 {
			buckets[i] = math.MaxFloat64
		} else {
			buckets[i] = float64(spec.Buckets[i])
		}
	}
	return &histogram{
		Histogram: tp.Tagged(spec.Tags).Histogram(
			spec.Name,
			tally.ValueBuckets(buckets),
		),
		lasts:       make([]int64, len(buckets)),
		bucketValue: spec.Buckets,
	}
}

type counter struct {
	tally.Counter

	last int64
}

func (c *counter) Set(total int64) {
	delta := total - c.last
	c.last = total
	c.Inc(delta)
}

type gauge struct {
	tally.Gauge
}

func (tg *gauge) Set(value int64) {
	tg.Update(float64(value))
}

type histogram struct {
	tally.Histogram

	// lasts keeps the last value pushed to tally
	lasts []int64
	// bucketValue keeps the static bucket value to be able to report correctly
	bucketValue []int64
}

// Set is log(n) because it performs binary search to find the index that bucket belongs to. Although, if the user
// didn't populate the HistogramSpec correctly, the first time it would incur additional cost of O(n) to insert the
// new bucket elements
func (th *histogram) Set(bucket int64, total int64) {
	index := sort.Search(len(th.lasts), func(i int) bool {
		return th.bucketValue[i] >= bucket
	})

	th.ensureBucket(index, bucket, false)
	th.recordValue(index, bucket, total)
}

// ensureBucket makes sure that the bucket at index is the same as the bucket from the user parameters.
// For the new API (SetIndex) we disable insertions because there is always going to be a mismatch.
func (th *histogram) ensureBucket(index int, bucket int64, panicOnError bool) {
	switch {
	case index >= len(th.bucketValue):
		th.bucketValue = append(th.bucketValue, bucket)
		th.lasts = append(th.lasts, 0)
	case bucket != th.bucketValue[index]:
		// Only the new API allows panics, we do this to ensure they are using it correctly
		if panicOnError {
			panic("insertion is not supported")
		}
		th.lasts = append(th.lasts[:index], append([]int64{0}, th.lasts[index:]...)...)
		th.bucketValue = append(th.bucketValue[:index], append([]int64{bucket}, th.bucketValue[index:]...)...)
	}
}

func (th *histogram) recordValue(index int, bucket int64, total int64) {
	delta := total - th.lasts[index]
	th.lasts[index] = total

	for i := int64(0); i < delta; i++ {
		th.RecordValue(float64(bucket))
	}
}

// SetIndex is O(1) because it can access the index directly. It allows to add missing buckets at the end, but in the
// middle will panic.
func (th *histogram) SetIndex(bucketIndex int, bucket int64, total int64) {
	th.ensureBucket(bucketIndex, bucket, true)
	th.recordValue(bucketIndex, bucket, total)
}
