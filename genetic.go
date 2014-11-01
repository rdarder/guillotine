package guillotine

import "math/rand"

type WeightedJoin struct {
	weight float32
	i, j   uint16
	config Join
}

type Genotype []WeightedJoin

func (g Genotype) copy() Genotype {
	c := make([]WeightedJoin, len(g))
	copy(c, g)
	return c
}

func NewGenotype(n uint16) Genotype {
	length := n * (n - 1) / 2
	return make([]WeightedJoin, length)
}

func NewRandomGenotype(nboards uint16, r *rand.Rand) Genotype {
	c := NewGenotype(nboards)
	k := 0
	for i := uint16(0); i < nboards; i++ {
		for j := i + 1; j < nboards; j++ {
			config := r.Intn(8)
			wj := &c[k]
			wj.i = i
			wj.j = j
			wj.config = Join(config)
			wj.weight = r.Float32()
			k++
		}
	}
	return c
}

func (c Genotype) Len() int           { return len(c) }
func (c Genotype) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Genotype) Less(i, j int) bool { return c[i].weight < c[j].weight }

//Create a fresh pair of Genotypes, utility function for Crossovers
func freshPair(p1, p2 Genotype) (n int, c1, c2 Genotype) {
	if n := len(p1); n != len(p2) {
		panic("weighted joins must have the same length")
	} else {
		return n, make([]WeightedJoin, n), make([]WeightedJoin, n)
	}
}

type Crossover func(p1, p2 Genotype, r *rand.Rand) (c1, c2 Genotype)

func UniformCrossover(p1, p2 Genotype, r *rand.Rand) (c1, c2 Genotype) {
	n, c1, c2 := freshPair(p1, p2)
	for i := 0; i < n; {
		rs := r.Int63()
		for j := uint(0); i < n && j < 63; j++ {
			if (rs & (1 << j)) == 0 {
				c1[i] = p1[i]
				c2[i] = p2[i]
			} else {
				c1[i] = p2[i]
				c2[i] = p1[i]
			}
			i++
		}
	}
	return
}

var _ Crossover = UniformCrossover

func OnePointCrossover(p1, p2 Genotype, r *rand.Rand) (c1, c2 Genotype) {
	n, c1, c2 := freshPair(p1, p2)
	cpoint := rand.Intn(n)
	copy(c1[:cpoint], p1[:cpoint])
	copy(c1[cpoint:], p2[cpoint:])

	copy(c2[:cpoint], p2[:cpoint])
	copy(c2[cpoint:], p1[cpoint:])
	return
}

var _ Crossover = OnePointCrossover

func TwoPointCrossover(p1, p2 Genotype, r *rand.Rand) (c1, c2 Genotype) {
	n, c1, c2 := freshPair(p1, p2)
	point1 := rand.Intn(n)
	point2 := rand.Intn(n)
	if point1 > point2 {
		point1, point2 = point2, point1
	}

	copy(c1[:point1], p1[:point1])
	copy(c1[point1:point2], p2[point1:point2])
	copy(c1[point2:], p1[point2:])

	copy(c2[:point1], p2[:point1])
	copy(c2[point1:point2], p1[point1:point2])
	copy(c2[point2:], p2[point2:])
	return
}

var _ Crossover = TwoPointCrossover

type Mutator func(Genotype, *rand.Rand)

//In-place chromosome mutation, by replacing some of the
//gene weights by a new random weight.
//Given a chromosome, RandomNorm(p, sigma)
//genes will mutate
//Mutations are done with replacement, meaning a weight
//can be mutated multiple times. The effect is that
//for a mu value close to the chromosome length,
//the actual number of mutated genes will be less.
//Hopefully that's not an intended usecase.
type NormalWeightMutator struct {
	Mean, StdDev float64
}

func (p NormalWeightMutator) Mutate(c Genotype, r *rand.Rand) {
	take := uint16(r.NormFloat64()*p.StdDev + p.Mean)
	for ; take > 0; take-- {
		i := rand.Intn(len(c))
		c[i].weight = rand.Float32()
	}
}

type NormalConfigMutator struct {
	Mean, StdDev float64
}

func (p NormalConfigMutator) Mutate(c Genotype, r *rand.Rand) {
	take := uint16(r.NormFloat64()*p.StdDev + p.Mean)
	for ; take > 0; take-- {
		i := rand.Intn(len(c))
		c[i].config = Join(rand.Intn(8))
	}
}

type CompoundWeightConfigMutator struct {
	Weight NormalWeightMutator
	Config NormalConfigMutator
}

func (c CompoundWeightConfigMutator) Mutate(g Genotype, r *rand.Rand) {
	c.Weight.Mutate(g, r)
	c.Config.Mutate(g, r)
}

var _ Mutator = NormalWeightMutator{}.Mutate
var _ Mutator = NormalConfigMutator{}.Mutate
var _ Mutator = CompoundWeightConfigMutator{}.Mutate
