package gcuts

//	"appengine/datastore"
import (
	"fmt"
	"github.com/crhym3/go-endpoints/endpoints"
	"github.com/rdarder/guillotine"
	"math"
	"math/rand"
	"net/http"
	"time"
)
const (
	gaTimeout = 10 * time.Second //Max time to spend per GeneticAlgorithm run.
)

// Greeting is a datastore entity that represents a single greeting.
// It also serves as (a part of) a response of GreetingService.
type Board struct {
	Width  uint `json:"width" endpoints:"req"`
	Height uint `json:"height" endpoints:"req"`
}
type BoardOrder struct {
	Board  Board `json:"board" endpoints:"req"`
	Amount uint  `json:"amount" endpoints:"req"`
}
type CutSpec struct {
	Orders   []BoardOrder            `json:"orders" endpoints:"req"`
	MaxWidth uint                    `json:"maxWidth"`
	Hints    *GeneticAlgorithmParams `json:"hints" endpoints:"req"`
}

type Placement struct {
	X       uint `json:"x"`
	Y       uint `json:"y"`
	Rotated bool `json:"rotated"`
}

type BoardPlacement struct {
	Orig      Board
	Oriented  Board     `json:"board" endpoints:"req"`
	Placement Placement `json:"placement" endpoints:"req"`
}

type CutResults struct {
	Placements   []BoardPlacement `json:"boardPlacements" endpoints:"required"`
	Sheet        Board            `json:"sheet"`
	Waste        uint             `json:"waste"`
	WastePercent float64          `json:"wastePercent"`
	RunDetails   RunDetails       `json:"runDetails" endpoints:"required"`
}
type RunDetails struct {
	Generations uint
}

type GeneticAlgorithmParams struct {
	ConfigMutateMean   float64 `endpoints:"d=5"`
	WeightMutateMean   float64 `endpoints:"d=5"`
	Crossover          string  `endpoints:"d=twopoint"`
	TournamentSize     uint    `endpoints:"d=8"`
	FittestProbability float32 `endpoints:"d=0.7"`
	Population         uint    `endpoints:"d=50"`
	Generations        uint    `endpoints:"d=100"`
	EliteSize          uint    `endpoints:"d=5"`
}

type Guillotine struct {
	r *rand.Rand
}

func NormUint(mean float64, stddev float64, r *rand.Rand) uint {
	return uint(math.Abs(r.NormFloat64()*stddev + mean))
}

// List responds with a list of all greetings ordered by Date field.
// Most recent greets come first.

func paramError(field string, v interface{}) error {
	return fmt.Errorf("Invalid %s option: <%v>", field, v)
}

var defaultHints GeneticAlgorithmParams = GeneticAlgorithmParams{
	ConfigMutateMean:   5,
	WeightMutateMean:   5,
	Crossover:          "twopoint",
	TournamentSize:     8,
	FittestProbability: 0.7,
	Population:         50,
	Generations:        200,
	EliteSize:          5,
}

func GetGeneticAlgorithm(spec *guillotine.CutSpec, params GeneticAlgorithmParams,
	r *rand.Rand) (*guillotine.GeneticAlgorithm, error) {

	if len(spec.Boards) < 2 {
		return nil, fmt.Errorf("Need at least two boards")
	}
	var breeder guillotine.Crossover
	switch params.Crossover {
	case "uniform":
		breeder = guillotine.UniformCrossover
	case "onepoint":
		breeder = guillotine.OnePointCrossover
	case "twopoint":
		breeder = guillotine.TwoPointCrossover
	default:
		return nil, paramError("Crossover", params.Crossover)
	}

	var evaluator guillotine.Fitness
	if spec.MaxWidth != 0 {
		evaluator = (*guillotine.LayoutTree).Height
	} else {
		evaluator = (*guillotine.LayoutTree).Area
	}

	if cMean := params.ConfigMutateMean; cMean < 0 {
		return nil, paramError("ConfigMutateMean", cMean)
	} else if wMean := params.WeightMutateMean; wMean < 0 {
		return nil, paramError("ConfigMutateMean", wMean)
	} else if population := params.Population; population < 1 || population > 1000 {
		return nil, paramError("Population", population)
	} else if tsize := params.TournamentSize; tsize < 1 || tsize > population {
		return nil, paramError("TournamentSize", tsize)
	} else if psel := params.FittestProbability; psel < 0 || psel >= 1 {
		return nil, paramError("fittestProbability", psel)
	} else if eliteSize := params.EliteSize; eliteSize < 0 || eliteSize > population {
		return nil, paramError("EliteSize", eliteSize)
	} else if generations := params.Generations; generations < 1 || generations > 10000 {
		return nil, paramError("Generations", generations)
	} else if genCost := int(population) * len(spec.Boards) * len(spec.Boards); genCost > 1000000 {
		//genCost approximately measures how much time it will take to process one generation
		//it's not an absolute time, but a setup with 2*genCost will be near 2*runtime
		//1MM is > 30boards * 1000 generations
		return nil, fmt.Errorf("Resource limits: Try lowering population or board count")
	} else {
		//
		return &guillotine.GeneticAlgorithm{
			Spec:      spec,
			Evaluator: evaluator,
			Mutator: guillotine.CompoundWeightConfigMutator{
				Weight: guillotine.NormalWeightMutator{
					Mean:   params.WeightMutateMean,
					StdDev: params.WeightMutateMean / 5,
				},
				Config: guillotine.NormalConfigMutator{
					Mean:   params.ConfigMutateMean,
					StdDev: params.ConfigMutateMean / 5,
				},
			}.Mutate,
			Breeder: breeder,
			SelectorBuilder: guillotine.NewTournamentSelectorBuilder(
				int(tsize), psel, r, true),
			R:              r,
			EliteSize:      eliteSize,
			PopulationSize: population,
			Generations:    generations,
		}, nil
	}
}

func CutSpecFromMessage(message *CutSpec) (*guillotine.CutSpec, error) {
	spec := &guillotine.CutSpec{
		Boards:   make([]guillotine.Board, 0, 100),
		MaxWidth: message.MaxWidth,
	}
	for i, order := range message.Orders {
		if order.Amount < 1 {
			return nil, fmt.Errorf("Invalid amount on order <%d>", i)
		} else {
			//this check should go in spec.Add()
			width, height := order.Board.Width, order.Board.Height
			if !spec.Fits(width, height) {
				return nil, fmt.Errorf("Invalid board dimensions: (%v, %v)", width, height)
			}
			for i := uint(0); i < order.Amount; i++ {
				spec.Add(width, height)
			}
		}
	}
	return spec, nil
}

func GetPlacements(lt *guillotine.LayoutTree) (sheet Board, bps []BoardPlacement) {
	boards := lt.Spec.Boards
	bps = make([]BoardPlacement, len(lt.Spec.Boards))
	drawing := guillotine.NewDrawer(lt).Draw()
	for i := range drawing.Boxes {
		bp := &bps[i]
		bp.Orig = Board{Width: boards[i].Width, Height: boards[i].Height}
		bp.Oriented = Board{drawing.Boxes[i].Width, drawing.Boxes[i].Height}
		bp.Placement.Rotated = lt.Picks[i].Rot
		bp.Placement.X = drawing.Boxes[i].X
		bp.Placement.Y = drawing.Boxes[i].Y
	}
	var width uint;
	if lt.Spec.MaxWidth != 0 { 
		// need to express limited/non-limited runs better, this spreads everywhere.
		width = lt.Spec.MaxWidth
	} else {
		width = drawing.Sheet.Width
	}
	sheet = Board{width, drawing.Sheet.Height}
	return sheet, bps
}

func (gn *Guillotine) Cut(r *http.Request, msg *CutSpec, resp *CutResults) error {
	if msg.Hints == nil {
		msg.Hints = &defaultHints
	}

	if cutSpec, err := CutSpecFromMessage(msg); err != nil {
		return err
	} else if ga, err := GetGeneticAlgorithm(cutSpec, *msg.Hints, gn.r); err != nil {
		return err
	} else {
		generations, layout := ga.TimeBoundedRun(gaTimeout)
		sheet, placements := GetPlacements(layout)
		resp.Waste = sheet.Width*sheet.Height - cutSpec.TotalArea
		resp.WastePercent = 100 * float64(resp.Waste) / float64(cutSpec.TotalArea)
		resp.Placements = placements
		resp.Sheet = sheet
		resp.RunDetails.Generations = generations
	}
	return nil
}

func (gn *Guillotine) RandomSpec(r *http.Request, p *endpoints.VoidMessage, spec *CutSpec) error{

	spec.MaxWidth = NormUint(200, 40, gn.r)
	for i := NormUint(5, 3, gn.r) + 2; i > 0; i-- {
		order := BoardOrder{
			Amount: NormUint(1, 0.5, gn.r) + 1,
			Board: Board{
				// at least one dimension below maxWidth
				Width:  NormUint(80, 40, gn.r) % spec.MaxWidth + 1, 
				Height: NormUint(100, 30, gn.r) + 1,
			},
		}
		spec.Orders = append(spec.Orders, order)
	}
	return nil
}
func (gn *Guillotine) DefaultHints(r *http.Request, p *endpoints.VoidMessage, hints *GeneticAlgorithmParams) error {
	*hints = defaultHints;
	return nil
}

func init() {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	guillotine := &Guillotine{r}

	api, err := endpoints.RegisterService(guillotine,
		"guillotine", "v1", "Guillotine Cuts API", true)
	if err != nil {
		panic(err.Error())
	}
	info := api.MethodByName("Cut").Info()
	info.Name, info.HTTPMethod, info.Path, info.Desc =
		"guillotine.cut", "POST", "cut", "Get a guillotine cut for the given boards."
	info = api.MethodByName("RandomSpec").Info()
	info.Name, info.HTTPMethod, info.Path, info.Desc =
		"guillotine.randomSpec", "GET", "randomSpec", "Generate a random cut spec."
	info = api.MethodByName("DefaultHints").Info()
	info.Name, info.HTTPMethod, info.Path, info.Desc =
		"guillotine.defaultHints", "GET", "defaultHints", "Get the default hints."
	endpoints.HandleHTTP()
}
