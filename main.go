package main

import (
	"math"
	crand "crypto/rand"
	"math/big"
	"math/rand"
	"runtime"
	"./multiswarm_pso"
)


type PositionSystem interface {
	init() PositionSystem
	add(p2 PositionSystem) PositionSystem
	addNum(p2 float64) PositionSystem
	sub(p2 PositionSystem) PositionSystem
	mul(p2 PositionSystem) PositionSystem
	mulNum(p2 float64) PositionSystem
}


type PositionXY struct {
	x, y float64
}

func (p *PositionXY) init() PositionSystem {
	return PositionSystem(PositionXY{x: 0, y: 0})
}

func (p *PositionXY) add(pos PositionSystem) PositionSystem {
	p2 := pos.(PositionXY)
	return PositionSystem(PositionXY{x: p.x + p2.x, y: p.y + p2.y})
}

func (p *PositionXY) addNum(p2 float64) PositionSystem {
	return PositionSystem(PositionXY{x: p.x + p2, y: p.y + p2})
}

func (p *PositionXY) sub(pos PositionSystem) PositionSystem {
	p2 := pos.(PositionXY)
	return PositionSystem(PositionXY{x: p.x - p2.x, y: p.y - p2.y})
}

func (p *PositionXY) mul(pos PositionSystem) PositionSystem {
	p2 := pos.(PositionXY)
	return PositionSystem(PositionXY{x: p.x * p2.x, y: p.y * p2.y})
}

func (p *PositionXY) mulNum(p2 float64) PositionSystem {
	return PositionSystem(PositionXY{x: p.x * p2, y: p.y * p2})
}

func randInt(min int, max int) int {
	return min + rand.Intn(max - min)
}



type WorldXY struct{
	leftTopEdgeCorner, rightDownEdgeCorner PositionXY
}

func (w *WorldXY) zeroPosition() PositionXY {
	return PositionXY{x: 0, y: 0}
}


func (w *WorldXY) guard(position PositionXY) bool {
	return position.x >= w.leftTopEdgeCorner.x && position.x <= w.rightDownEdgeCorner.x &&
		position.y >= w.leftTopEdgeCorner.y && position.y <= w.rightDownEdgeCorner.y
}


func (w *WorldXY) randomPosition() PositionXY {
	return PositionXY{
		x: float64(randInt(int(w.leftTopEdgeCorner.x), int(w.rightDownEdgeCorner.x))),
		y: float64(randInt(int(w.leftTopEdgeCorner.y), int(w.rightDownEdgeCorner.y)))}
}

func dejonge(position PositionXY) float64 {
	return math.Pow(float64(position.x), 2.0) + math.Pow(float64(position.y), 2.0);
}

func schwefel(position PositionXY) float64 {
	return (-position.x*math.Sin(math.Sqrt(math.Abs(position.x)))) + (-position.y*math.Sin(math.Sqrt(math.Abs(position.y))))
}

func ackley(position PositionXY) float64 {
	return 1/math.Pow(math.E, 5) * math.Sqrt(position.x*position.x + position.y*position.y) + 3 * (math.Cos(2*position.x)+math.Sin(2*position.y))
}

func main() {
	var v, _ = crand.Int(crand.Reader, big.NewInt(1000))
	rand.Seed(v.Int64())

	var pos PositionSystem
	pos = PositionXY{-500, -500}


	var world multiswarm_pso.World
	world = WorldXY{
		leftTopEdgeCorner: pos,
		rightDownEdgeCorner: PositionXY{500, 500},
	}

	var fnc multiswarm_pso.FCost
	fnc = schwefel

	parameters := multiswarm_pso.PSOParameters{
		SwarmCount: runtime.NumCPU(),
		ParticlesCountInSwarm: 20,
		IterationCount: 100,
		SwitchBestParticleCount: 5,
		SwitchCount: 10,
		Function: fnc,
		World: world,
	}

	multiswarm_pso.MultiSwarm(parameters)
}
