package main

import (
	"github.com/rdarder/guillotine"
)
import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"
)

/* Random search of cutting layouts */

type Scores struct {
	avg      float64
	stddev   float64
	best     *guillotine.LayoutTree
	bestArea uint
	waste    float64
}

type RunSpec struct {
	cutSpec *guillotine.CutSpec
	area    uint
}

func main() {
	var nboards = flag.Int("nboards", 10, "Number of boards")
	var area = flag.Int("area", 2000, "Target total area")
	var maxWidth = flag.Int("maxWidth", 0, "sheet max width")
	var seed = flag.Int64("seed", time.Now().Unix(), "Random seed for repeatable runs")
	var tries = flag.Int("tries", 100000, "Number of tries")
	flag.Parse()

	r := rand.New(rand.NewSource(*seed))
	var width, height uint
	var limitWidth bool
	if *maxWidth == 0 {
		width, height = guillotine.AreaDimensions(float64(*area), r)
		limitWidth = false
	} else {
		width, height = guillotine.MaxWidthDimensions(*maxWidth, r)
		limitWidth = true
	}
	spec := guillotine.NewRandomSpec(*nboards, width, height, r, limitWidth)
	target := width * height

	results := make([]uint, 0, *tries)
	scores := &Scores{bestArea: math.MaxUint32}
	for i := 0; i < *tries; i++ {
		g := guillotine.NewRandomGenotype(uint16(*nboards), r)
		lt := guillotine.GetPhenotype(spec, g)
		area := lt.Area()
		results = append(results, area)
		if area < scores.bestArea {
			scores.bestArea = area
		}
	}
	scores.waste = (float64(scores.bestArea)/float64(target) - 1) * 100
	scores.avg = avg(results)
	scores.stddev = stddev(results, scores.avg)
	fmt.Printf("%+v\n", scores)
	fmt.Printf("area: %v, spec: %+v\n", target, spec)
}

func avg(values []uint) float64 {
	var sum uint = 0
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}

func stddev(values []uint, avg float64) float64 {
	var ds float64 = 0
	for v := range values {
		if d := float64(v) - avg; d > 0 {
			ds += d
		} else {
			ds += -d
		}
	}
	variance := ds / float64(len(values))
	return math.Sqrt(variance)
}
