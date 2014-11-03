package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/rdarder/guillotine"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

type Solution struct {
	Spec   *guillotine.CutSpec
	Layout *guillotine.LayoutTree
}

func main() {
	var nboards = flag.Int("nboards", 10, "Number of boards")
	var population = flag.Int("population", 300, "Population size")
	var tsize = flag.Int("tsize", 5, "Tournament size")
	var eliteSize = flag.Int("eliteSize", 10, "Elite size")
	var area = flag.Int("area", 2000, "Target total area")
	var maxWidth = flag.Int("maxWidth", 0, "sheet max width")
	var psel = flag.Float64("psel", 0.8, "Tournament selection probability")
	var cx = flag.String("crossover", "uniform", "Crossover strategy")
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	var weightMutateMean = flag.Float64("weightMutateMean", 10,
		"Mean number of gene weights to be mutated on each individual")
	var configMutateMean = flag.Float64("configMutateMean", 10,
		"Mean number of pick configs to be mutated on each individual")
	var generations = flag.Int("generations", 10, "Number of generations")
	var seed = flag.Int64("seed", time.Now().Unix(), "Random seed for repeatable runs")

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var crossover guillotine.Crossover
	switch *cx {
	case "uniform":
		crossover = guillotine.UniformCrossover
	case "onepoint":
		crossover = guillotine.OnePointCrossover
	case "twopoint":
		crossover = guillotine.TwoPointCrossover
	default:
		panic("Invalid option for crossover")
	}

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

	ga := &guillotine.GeneticAlgorithm{
		Spec:        spec,
		Evaluator:   (*guillotine.LayoutTree).Area,
		Mutator: guillotine.CompoundWeightConfigMutator{
			Weight: guillotine.NormalWeightMutator{
				Mean:   *weightMutateMean,
				StdDev: *weightMutateMean / 5,
			},
			Config: guillotine.NormalConfigMutator{
				Mean:   *configMutateMean,
				StdDev: *configMutateMean / 5,
			},
		}.Mutate,
		Breeder:         crossover,
		SelectorBuilder: guillotine.NewTournamentSelectorBuilder(*tsize, float32(*psel), r, true),
		R:               r,
		EliteSize:       uint(*eliteSize),
	}
	pop := guillotine.NewRandomPopulation(uint16(*nboards), uint(*population), r)
	rankedPop := ga.Evaluate(pop)
	for i := 1; i < *generations; i++ {
		pop = ga.Next(rankedPop)
		rankedPop = ga.Evaluate(pop)
	}

	bestLayout := guillotine.GetPhenotype(spec, rankedPop.Pop[0])
	drawer := guillotine.NewDrawer(bestLayout)
	b, err := json.Marshal(drawer.Draw())
	if err != nil {
		log.Fatal("error:", err)
	}
	os.Stdout.Write(b)
	best := rankedPop.Fitnesses[0]
	fmt.Printf("\nWaste: %.2f%%\n", 100*(float32(best)/float32(target)-1))
}
