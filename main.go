package main

import (
	"flag"
	"log"
	"math"
	"math/rand"

	"time"

	linuxproc "github.com/c9s/goprocinfo/linux"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

const sr = beep.SampleRate(44100)
const sampleRate = 500 * time.Millisecond

var threshold int

func init() {
	defaultThreshold := 0
	flag.IntVar(&threshold, "threshold", defaultThreshold, "value baseline")
	flag.IntVar(&threshold, "t", defaultThreshold, "value baseline")
}

func main() {
	flag.Parse()

	var p int
	speaker.Init(sr, sr.N(sampleRate)/5)
	rng := rand.New(rand.NewSource(0x1337))
	sign := 1.0
	speaker.Play(beep.StreamerFunc(func(samples [][2]float64) (int, bool) {
		for i := range samples {
			if rng.Intn(int(sr)) < p {
				samples[i] = [2]float64{sign, sign}
				sign *= -1
			} else {
				samples[i] = [2]float64{0, 0}
			}
		}
		return len(samples), true
	}))

	stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		log.Fatal("stat read fail")
	}
	prevSystem := stat.CPUStatAll.System
	prevUser := stat.CPUStatAll.User

	for range time.Tick(sampleRate) {
		stat, err = linuxproc.ReadStat("/proc/stat")
		if err != nil {
			log.Fatal("stat read fail")
		}
		system := stat.CPUStatAll.System - prevSystem
		user := stat.CPUStatAll.User - prevUser
		prevSystem = stat.CPUStatAll.System
		prevUser = stat.CPUStatAll.User

		rate := float64(user+system)*sampleRate.Seconds() - float64(threshold)

		// If it becomes too high it becomes inaudible.
		// Limit at 3.6 Roentgen.
		rate = math.Min(rate, float64(sr)/3.6)
		speaker.Lock()
		p = int(rate)
		speaker.Unlock()
	}
}
