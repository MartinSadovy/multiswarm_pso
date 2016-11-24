package main

import (
	"math"
	crand "crypto/rand"
	"math/big"
	"math/rand"
	"runtime"
	"./multiswarm_pso"
	"fmt"
	"flag"
	"strings"
	"strconv"
	"os"
)

// this is only two dimension implementation function by prof. Zelinka in BIV silab

func dejong1(position multiswarm_pso.Position) float64 {
	return math.Pow(float64(position[0]), 2.0) + math.Pow(float64(position[1]), 2.0);
}

func schwefel(position multiswarm_pso.Position) float64 {
	return (-position[0]*math.Sin(math.Sqrt(math.Abs(position[0])))) + (-position[1]*math.Sin(math.Sqrt(math.Abs(position[1]))))
}

func ackley(position multiswarm_pso.Position) float64 {
	return 1/math.Pow(math.E, 5) * math.Sqrt(position[0]*position[0] + position[1]*position[1]) + 3 * (math.Cos(2*position[0])+math.Sin(2*position[1]))
}

// just parse parameters for size of world
func parseQuadraticWorld(input string) multiswarm_pso.QuadraticWorld {
	coords := strings.Split(input, ",")
	if (len(coords) % 2 != 0) {
		fmt.Println("Parameter quadraticWorld must have even count of parameters (coords).")
	}

	length := len(coords) / 2

	start := make(multiswarm_pso.Position, length)
	end := make(multiswarm_pso.Position, length)
	var err, err2 error
	for i := 0; i < length; i++ {
		start[i], err = strconv.ParseFloat(coords[i], 64)
		end[i], err2  = strconv.ParseFloat(coords[length + i], 64)
		if (err != nil || err2 != nil) {
			fmt.Printf("Parameter quadraticWorld has invalid number on position: %d or %d", i, length + i)
			runtime.Goexit()
		}

		if (start[i] >= end[i]) {
			fmt.Printf("Parameter quadraticWorld has inverse size (lower must be in first part) on position: %d and %d", i, length + i)
			runtime.Goexit()
		}
	}

	return multiswarm_pso.QuadraticWorld{
		LeftTopEdgeCornerVector: start,
		RightDownEdgeCornerVector: end,
	}

}


func main() {
	var v, _ = crand.Int(crand.Reader, big.NewInt(1000))
	rand.Seed(v.Int64())

	pFunc := flag.String("function", "dejong1", "Switch function will be used (dejong1, schwefel, ackley)")
	pQWorld := flag.String("quadraticWorld", "-500,-500,500,500", "Write corner points for hypercube, eg. 2D: X_start,Y_start,X_end,Y_end")
	pSwarms := flag.Int("swarms", runtime.NumCPU(), "Count of independent swarms (background is swarm concurrency). >0")
	pParticlesCountInSwarm := flag.Int("particlesInSwarm", 20, "Particles in swarm. >0")
	pIterationCount := flag.Int("iterations", 10000, "Total iterations per swarm. >0")
	pSwitchCount := flag.Int("switches", 5, "Count of switch after N iterations where N = iterations/(switches+1) except last iteration")
	pSwitchBestParticleCount := flag.Int("swichedParticle", 5, "Switch the best N particles between swarm. >0")
	pC1 := flag.Float64("pso_c1", 2, "Parameter C1 in velocity update")
	pC2 := flag.Float64("pso_c2", 2, "Parameter C2 in velocity update")
	pGif := flag.String("gif", "simulation.gif", "Where save gif simulation? Empty = no generating")
	pHelp := flag.Bool("help", false, "How can I help you? RTFM!")

	flag.Parse()

	var fnc multiswarm_pso.FCost
	switch *pFunc {
	case "dejong1": fnc = dejong1
	case "ackley": fnc = ackley
	case "schwefel": fnc = schwefel
	default:
		fmt.Fprintln(os.Stderr, "You want unknown function")
		return;
	}

	if (*pSwarms < 1 || *pParticlesCountInSwarm < 1 || *pIterationCount < 1 || *pSwitchBestParticleCount < 1) {
		fmt.Fprintln(os.Stderr, "Any value is not valid.")
		return;
	}

	parameters := multiswarm_pso.PSOParameters{
		SwarmCount: *pSwarms,
		ParticlesCountInSwarm: *pParticlesCountInSwarm,
		IterationCount: *pIterationCount,
		SwitchCount: *pSwitchCount,
		SwitchBestParticleCount: *pSwitchBestParticleCount,
		C1: *pC1,
		C2: *pC2,
		Function: fnc,
		World: parseQuadraticWorld(*pQWorld),
	}

	if (*pHelp) {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults();
		return;
	}

	if (parameters.World.DimensionSize() != 2) {
		fmt.Fprintln(os.Stderr, "Please create 2D world, gif generator is not able generate in other dimension, and valuation function is only for X,Y, but common algorithm is multi-dimensional.")
		return;
	}

	renderChan, doneGifChan := multiswarm_pso.CreateRenderPipe(*pGif, parameters) // generate pipeline for visualization
	best := multiswarm_pso.MultiSwarm(parameters, renderChan) // calculate swarm
	fmt.Printf("GBest %.5f found on position: %s\n", fnc(best), multiswarm_pso.PrintPosition(best)) // print result
	<-doneGifChan // stop main "thread/rutine" after gif is saved
}
