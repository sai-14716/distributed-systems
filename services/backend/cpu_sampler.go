package main

import (
	"math"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
)

// cpuSampler tracks per-process CPU utilization using Getrusage, and exposes a smoothed percent.
// This measures CPU actually consumed by request work (as opposed to sleeps).
type cpuSampler struct {
	pctBits uint64 // float64 bits
}

func newCPUSampler(sampleEvery time.Duration, ewmaAlpha float64) *cpuSampler {
	if sampleEvery <= 0 {
		sampleEvery = 500 * time.Millisecond
	}
	if ewmaAlpha <= 0 || ewmaAlpha > 1 {
		ewmaAlpha = 0.2
	}

	s := &cpuSampler{}
	atomic.StoreUint64(&s.pctBits, math.Float64bits(0))

	go func() {
		var lastCPU float64
		var lastWall time.Time
		var smoothed float64
		t := time.NewTicker(sampleEvery)
		defer t.Stop()

		for range t.C {
			cpu := processCPUSeconds()
			now := time.Now()
			if lastWall.IsZero() {
				lastCPU = cpu
				lastWall = now
				continue
			}
			dCPU := cpu - lastCPU
			dWall := now.Sub(lastWall).Seconds()
			lastCPU = cpu
			lastWall = now
			if dWall <= 0 {
				continue
			}

			// Normalize by CPU count so we stay in the familiar 0..100% range.
			pct := (dCPU / dWall) / float64(runtime.NumCPU()) * 100.0
			if pct < 0 {
				pct = 0
			}
			if pct > 100 {
				pct = 100
			}

			if smoothed == 0 {
				smoothed = pct
			} else {
				smoothed = ewmaAlpha*pct + (1-ewmaAlpha)*smoothed
			}
			atomic.StoreUint64(&s.pctBits, math.Float64bits(smoothed))
		}
	}()

	return s
}

func (s *cpuSampler) CPUPct() float64 {
	return math.Float64frombits(atomic.LoadUint64(&s.pctBits))
}

func processCPUSeconds() float64 {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return 0
	}
	ut := float64(ru.Utime.Sec) + float64(ru.Utime.Usec)/1e6
	st := float64(ru.Stime.Sec) + float64(ru.Stime.Usec)/1e6
	return ut + st
}

func cpuBucket5(cpuPct float64) int {
	if cpuPct < 0 {
		cpuPct = 0
	}
	if cpuPct > 100 {
		cpuPct = 100
	}
	b := int(cpuPct) / 5
	if b < 0 {
		return 0
	}
	if b > 20 {
		return 20
	}
	return b
}
