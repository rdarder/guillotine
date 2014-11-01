package guillotine

import (
	"fmt"
	"math/rand"
	"sort"
)

var _ = fmt.Println

func GetPhenotype(nboards uint16, genotype Genotype) *LayoutTree {
	var  gcopy Genotype = make([]WeightedJoin, len(genotype))
	copy(gcopy[:], genotype[:])
	sort.Sort(gcopy)
	genotype = gcopy
//	sort.Sort(genotype)
	lt := emptyLayoutTree(nboards)
	for i := 0; nboards > 1; i++ {
		wj := &genotype[i]
		//		fmt.Printf("%v -> %v: (%v,%v)\n", i, nboards, wj.i, wj.j)
		if lt.take(wj.i, wj.j, wj.config) {
			nboards -= 1
		}
	}
	return lt
}

type Population []Genotype

func NewRandomPopulation(nboards uint16, size uint, r *rand.Rand) Population {
	pop := make([]Genotype, size)
	for i := range pop {
		pop[i] = NewRandomGenotype(nboards, r)
	}
	return pop
}

/*
func (id Individual) evaluate(spec *CutSpec) {
	nboards := uint16(len(spec.Boards))
	id.phenotype = GetPhenotype(nboards, id.genotype)
	id.fitness = id.phenotype.Area(spec)
}

func (pop Population) evaluate(spec *CutSpec) {
	for _, id := range pop {
		id.evaluate(spec)
	}
}
*/
type RankedPopulation struct {
	Pop       Population
	Fitnesses []uint
}

func (rp *RankedPopulation) Less(i, j int) bool {
	return rp.Fitnesses[i] < rp.Fitnesses[j]
}
func (rp *RankedPopulation) Swap(i, j int) {
	rp.Pop[i], rp.Pop[j] = rp.Pop[j], rp.Pop[i]
	rp.Fitnesses[i], rp.Fitnesses[j] = rp.Fitnesses[j], rp.Fitnesses[i]
}
func (rp *RankedPopulation) Len() int { return len(rp.Pop) }

type Selector interface {
	next() Genotype
}
type SelectorBuilder func(rp *RankedPopulation) Selector

type TournamentSelector struct {
	size int
	buf  FitnessPositions
	p    float32
	rp   *RankedPopulation
	r    *rand.Rand
	min  bool
}

func NewTournamentSelectorBuilder(size int, p float32, r *rand.Rand, min bool) SelectorBuilder {
	return func(rp *RankedPopulation) Selector {
		return &TournamentSelector{
			size: size,
			buf:  make([]FitnessPosition, size),
			p:    p,
			r:    r,
			rp:   rp,
			min:  min,
		}
	}
}

type FitnessPosition struct {
	i       int
	fitness uint
}

type FitnessPositions []FitnessPosition

func (fps FitnessPositions) Len() int           { return len(fps) }
func (fps FitnessPositions) Less(i, j int) bool { return fps[i].fitness < fps[j].fitness }
func (fps FitnessPositions) Swap(i, j int)      { fps[i], fps[j] = fps[j], fps[i] }

func (ts *TournamentSelector) winnerRank() int {
	//This could be faster if modelled with a
	//negative binomial distribution generator
	for {
		for i := 0; i < ts.size; i++ {
			if ts.r.Float32() > ts.p {
				return i
			}
		}
	}
}

func (ts *TournamentSelector) next() Genotype {
	fps := ts.buf
	for i := 0; i < ts.size; i++ {
		ri := ts.r.Intn(len(ts.rp.Fitnesses))
		fps[i].i = ri
		fps[i].fitness = ts.rp.Fitnesses[ri]
	}
	var winnerIndex int
	winnerRank := ts.winnerRank()
	if ts.min {
		winnerIndex = fps.getKminIndex(winnerRank)
	} else {
		winnerIndex = fps.getKmaxIndex(winnerRank)
	}
	//	fmt.Printf("candidates: %v. winnerRank: %v, winnerIndex: %v\n", fps, winnerRank, winnerIndex)
	return ts.rp.Pop[winnerIndex]
}

func (fps FitnessPositions) getKminIndex(k int) int {
	//This would be faster by implementing quickSelect,
	//and/or lower k specific implementations.
	//another alternative is just to keep an inverted K sized heap
	//replacing the root when the new entry is less than it
	sort.Sort(fps)
	return fps[k].i
}
func (fps FitnessPositions) getKmaxIndex(k int) int {
	sort.Reverse(fps)
	return fps[k].i
}

func (pop Population) checkEvenSize() {
	if len(pop)%2 != 0 {
		panic("Population size must be even")
	}
}

type GeneticAlgorithm struct {
	TotalBoards     uint16
	Spec            *CutSpec
	Evaluator       Fitness
	Mutator         Mutator
	Breeder         Crossover
	SelectorBuilder SelectorBuilder
	R               *rand.Rand
	EliteSize       int
}

func (ga GeneticAlgorithm) breed(p1, p2 Genotype) (c1, c2 Genotype) {
	c1, c2 = ga.Breeder(p1, p2, ga.R)
	ga.Mutator(c1, ga.R)
	ga.Mutator(c2, ga.R)
	return c1, c2
}

func (ga GeneticAlgorithm) Evaluate(pop Population) (rp *RankedPopulation) {
	fitness := make([]uint, len(pop))
	for i, genotype := range pop {
		phenotype := GetPhenotype(ga.TotalBoards, genotype)
		fitness[i] = ga.Evaluator(phenotype, ga.Spec)
	}
	rp = &RankedPopulation{pop, fitness}
//	fmt.Println(rp.Fitnesses)
	sort.Sort(rp)
//	fmt.Println(rp.Fitnesses)
	return rp
}

func (ga GeneticAlgorithm) Next(rp *RankedPopulation) Population {
	selector := ga.SelectorBuilder(rp)
	psize := len(rp.Pop)
	pepsi := make([]Genotype, psize)
	copy(pepsi[:ga.EliteSize], rp.Pop[:ga.EliteSize])
	for i := ga.EliteSize; i < psize; i++ {
		p1, p2 := selector.next(), selector.next()
		c1, c2 := ga.breed(p1, p2)
		pepsi[i] = c1
		if i < psize - 1 {
			i++
			pepsi[i] = c2
		}
	}
	return pepsi
}
