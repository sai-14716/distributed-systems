//go:build windows

package main

// Windows doesn't provide Getrusage; keep CPU at 0% rather than failing to build.
func processCPUSeconds() float64 { return 0 }
