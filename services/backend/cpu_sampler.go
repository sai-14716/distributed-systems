package main

import (
	"math"
	"runtime"
	"sync/atomic"
	"time"
)

type cpuSampler struct {
	alpha   float64
	pctBits uint64 // float64 bits
}

func newCPUSampler(every time.Duration, alpha float64) *cpuSampler {
	if every <= 0 {
		every = 250 * time.Millisecond
	}
	if alpha <= 0 || alpha > 1 {
		alpha = 0.2
	}

	s := &cpuSampler{alpha: alpha}

	go func() {
		var (
			lastCPU  = processCPUSeconds()
			lastAt   = time.Now()
			lastEWMA = 0.0
		)

		t := time.NewTicker(every)
		defer t.Stop()

		for now := range t.C {
			curCPU := processCPUSeconds()
			dCPU := curCPU - lastCPU
			dt := now.Sub(lastAt).Seconds()

			pct := 0.0
			if dt > 0 && dCPU > 0 {
				denom := float64(runtime.NumCPU()) * dt
				if denom > 0 {
					pct = (dCPU / denom) * 100.0
				}
			}

			if pct < 0 {
				pct = 0
			}
			if pct > 100 {
				pct = 100
			}

			lastEWMA = s.alpha*pct + (1.0-s.alpha)*lastEWMA
			if math.IsNaN(lastEWMA) || math.IsInf(lastEWMA, 0) {
				lastEWMA = 0
			}

			atomic.StoreUint64(&s.pctBits, math.Float64bits(lastEWMA))

			lastCPU = curCPU
			lastAt = now
		}
	}()

	return s
}

func (s *cpuSampler) CPUPct() float64 {
	if s == nil {
		return 0
	}
	return math.Float64frombits(atomic.LoadUint64(&s.pctBits))
}

// cpuBucket5 maps 0..100 CPU% to 0..20 (5% buckets).
func cpuBucket5(cpuPct float64) int {
	if cpuPct <= 0 {
		return 0
	}
	if cpuPct >= 100 {
		return 20
	}
	b := int(cpuPct / 5.0)
	if b < 0 {
		b = 0
	}
	if b > 20 {
		b = 20
	}
	return b
}
