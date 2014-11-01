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
	var width = flag.Int("width", 800, "Target total width")
	var height = flag.Int("height", 800, "Target total height")
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

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	r := rand.New(rand.NewSource(*seed))
	spec := guillotine.NewRandomSpec(*nboards, *width, *height, r)
	target := *width * (*height)
	ga := &guillotine.GeneticAlgorithm{
		TotalBoards: uint16(*nboards),
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
		Breeder: crossover,
		SelectorBuilder: guillotine.NewTournamentSelectorBuilder(
			*tsize, float32(*psel), r, true),
		R:         r,
		EliteSize: *eliteSize}
	//size int, p float32, r *rand.Rand, min bool)
	pop := guillotine.NewRandomPopulation(uint16(*nboards), uint(*population), r)
	var rankedPop *guillotine.RankedPopulation
	//	fmt.Printf("Spec: %v\n", spec.Boards)
	//	fmt.Printf("Target: %v\n", target)
	for i := 0; i < *generations; i++ {
		rankedPop = ga.Evaluate(pop)
		pop = ga.Next(rankedPop)
	}
	bestLayout := guillotine.GetPhenotype(uint16(*nboards), rankedPop.Pop[0])

	drawer := guillotine.NewDrawer(bestLayout, spec)
	boxes := drawer.Draw()

	b, err := json.Marshal(boxes)
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)
	best := int(rankedPop.Fitnesses[0])
	fmt.Printf("%v%%\n", 100*best/target)

	//fmt.Println("Layout:\n%v\n%v", bestLayout.Picks, bestLayout.Stacks)

	//	bestLayout := guillotine.GetPhenotype(uint16(nboards), pop[0])
	//	fmt.Printf("Layout:\n%v\n%v\n", bestLayout.Picks, bestLayout.Stacks)
	//	fmt.Printf("Genotype:\n%v\n", pop[0])

	/*
			nboards         uint16
		spec            *CutSpec
		evaluator       Fitness
		mutator         Mutator
		breeder         Crossover
		selectorBuilder SelectorBuilder
		r               *rand.Rand
		minFitness      bool

	*/

}
