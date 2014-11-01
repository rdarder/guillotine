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
	var nboards, tries, width, height int
	var seed int64
	flag.IntVar(&nboards, "nboards", 10, "Number of boards")
	flag.IntVar(&width, "width", 400, "Width")
	flag.IntVar(&height, "height", 400, "Height")
	flag.IntVar(&tries, "tries", 10000, "Number of random tries")
	flag.Int64Var(&seed, "seed", time.Now().Unix(), "Random seed for repeatable runs")
	flag.Parse()

	target := width * height
	r := rand.New(rand.NewSource(seed))
	spec := guillotine.NewRandomSpec(nboards, width, height, r)
	results := make([]uint, 0, tries)
	scores := &Scores{bestArea: math.MaxUint32}
	for i := 0; i < tries; i++ {
		g := guillotine.NewRandomGenotype(uint16(nboards), r)
		lt := guillotine.GetPhenotype(uint16(nboards), g)
		area := lt.Area(spec)
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
	if len(values) == 0 {
		return 0
	}
	var sum uint = 0
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}

func stddev(values []uint, avg float64) float64 {
	if len(values) == 0 {
		return 0
	}
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
