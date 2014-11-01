package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/rdarder/guillotine"
	"math/rand"
	"os"
	"time"
	"log"
	"runtime/pprof"
)

type Solution struct {
	Spec   *guillotine.CutSpec
	Layout *guillotine.LayoutTree
}

func main() {
	var nboards, generations, population, tsize, eliteSize, width, height int
	var seed int64
	var psel, mutateMean float64
	var cx, cpuprofile string
	flag.IntVar(&nboards, "nboards", 10, "Number of boards")
	flag.IntVar(&population, "population", 300, "Population size")
	flag.IntVar(&tsize, "tsize", 5, "Tournament size")
	flag.IntVar(&eliteSize, "eliteSize", 10, "Elite size")
	flag.IntVar(&width, "width", 800, "Target total width")
	flag.IntVar(&height, "height", 800, "Target total height")
	flag.Float64Var(&psel, "psel", 0.8, "Tournament selection probability")
	flag.StringVar(&cx, "crossover", "uniform", "Crossover strategy")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.Float64Var(&mutateMean, "mutateMean", 10,
		"Mean number of genes to be mutated on each individual")
	flag.IntVar(&generations, "generations", 10, "Number of generations")
	flag.Int64Var(&seed, "seed", time.Now().Unix(), "Random seed for repeatable runs")

	flag.Parse()
	var crossover guillotine.Crossover
	switch cx {
	case "uniform":
		crossover = guillotine.UniformCrossover
	case "onepoint":
		crossover = guillotine.OnePointCrossover
	case "twopoint":
		crossover = guillotine.TwoPointCrossover
	default:
		panic("Invalid option for crossover")

	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	r := rand.New(rand.NewSource(seed))
	spec := guillotine.NewRandomSpec(nboards, width, height, r)
	ga := &guillotine.GeneticAlgorithm{
		TotalBoards: uint16(nboards),
		Spec:        spec,
		Evaluator:   (*guillotine.LayoutTree).Area,
		Mutator: guillotine.NormalReplaceMutator{
			Mean:   mutateMean,
			StdDev: mutateMean / 5,
		}.Mutate,
		Breeder: crossover,
		SelectorBuilder: guillotine.NewTournamentSelectorBuilder(
			tsize, float32(psel), r, true),
		R:         r,
		EliteSize: eliteSize}
	//size int, p float32, r *rand.Rand, min bool)
	pop := guillotine.NewRandomPopulation(uint16(nboards), uint(population), r)
	var rankedPop *guillotine.RankedPopulation
	//	fmt.Printf("Spec: %v\n", spec.Boards)
	//	fmt.Printf("Target: %v\n", target)
	for i := 0; i < generations; i++ {
		rankedPop = ga.Evaluate(pop)
		pop = ga.Next(rankedPop)
	}
	bestLayout := guillotine.GetPhenotype(uint16(nboards), rankedPop.Pop[0])

	drawer := guillotine.NewDrawer(bestLayout, spec)
	boxes := drawer.Draw()

	b, err := json.MarshalIndent(boxes, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)
	best := int(rankedPop.Fitnesses[0])
	fmt.Printf("%v%%\n", 100*best/(width*height))

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
