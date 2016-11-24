package multiswarm_pso

import (
	"math/rand"
	"sort"
	"github.com/gonum/floats"
	"fmt"
	"bytes"
)

type Position []float64

func PrintPosition(p Position) string {
	var buffer bytes.Buffer
	for i, v := range p {
		buffer.WriteString(fmt.Sprintf("%.5f", v))

		if (len(p) != i + 1) {
			buffer.WriteString(", ")
		}
	}
	return buffer.String()
}

type World interface {
	DimensionSize() int
	Guard(position Position) bool
	RandomPosition() Position
}

type PSOParameters struct {
	SwarmCount, ParticlesCountInSwarm, IterationCount, SwitchCount, SwitchBestParticleCount int
	C1, C2											float64
	Function                                                                                FCost
	World                                                                                   World
}

type particle struct {
	position, best, velocity Position
}

type swarm struct {
	particles []particle
	best      Position
	fcost     FCost
}

type FCost func(Position) float64;

// formula for calc new velocity (from BIV silab prof. Zelinka)
func velocityCalc(parameters PSOParameters, swarmInstance *swarm, particle *particle) Position {
	v := make(Position, parameters.World.DimensionSize())
	v2 := make(Position, parameters.World.DimensionSize())
	v3 := make(Position, parameters.World.DimensionSize())

	floats.SubTo(v2, particle.best, particle.position)
	floats.SubTo(v3, swarmInstance.best, particle.position)

	floats.Add(v, particle.velocity) // interial
	floats.AddScaled(v, parameters.C1 * rand.Float64(), v2) // cognitive
	floats.AddScaled(v, parameters.C2 * rand.Float64(), v3) // social
	return v
}

// create swarm instance
func createSwarm(parameters PSOParameters) *swarm {
	particles := make([]particle, parameters.ParticlesCountInSwarm);

	var gBest Position;

	for i, _ := range particles {
		particles[i].position = parameters.World.RandomPosition()
		particles[i].velocity = parameters.World.RandomPosition()
		particles[i].best = particles[i].position;
		if (i == 0 || parameters.Function(particles[i].best) < parameters.Function(gBest)) {
			gBest = particles[i].best;
		}
	}

	var b *swarm
	b = &swarm{
		particles,
		gBest,
		parameters.Function,
	};
	return b
}

// Sorting method

func (a *swarm) Len() int {
	return len(a.particles)
}
func (a *swarm) Swap(i, j int) {
	a.particles[i], a.particles[j] = a.particles[j], a.particles[i]
}
func (a *swarm) Less(i, j int) bool {
	return a.fcost(a.particles[i].best) < a.fcost(a.particles[j].best)
}

// Clone swarm (use to rendering)
func (g *swarm) CreateCopy() *swarm {
	copyParticles := make([]particle, len(g.particles))
	for y, p := range g.particles {
		copyParticles[y] = particle{
			p.position,
			p.best,
			p.velocity,
		}
	}

	return &swarm{
		copyParticles,
		g.best,
		g.fcost,
	}
}

// iterate and update gBest for swarm
func iteration(parameters PSOParameters, swarmInstance *swarm) {

	for i, particle := range swarmInstance.particles {
		newVelocity := velocityCalc(parameters, swarmInstance, &particle)
		newPosition := make(Position, parameters.World.DimensionSize())
		floats.AddTo(newPosition, particle.position, newVelocity)


		// guard - is new position in world?
		if (parameters.World.Guard(newPosition)) {
			particle.position = newPosition
		}

		// update pbest (and gbest in swarm)
		fitness := swarmInstance.fcost(particle.position)
		if (fitness < swarmInstance.fcost(particle.best)) {
			particle.best = particle.position;

			if (swarmInstance.fcost(swarmInstance.best) > swarmInstance.fcost(particle.best)) {
				swarmInstance.best = particle.best;
			}
		}

		swarmInstance.particles[i] = particle
	}

}

// Multiswarm simulation in routines by swarm
func MultiSwarm(parameters PSOParameters, partResultChannel [](chan swarm)) Position {

	swarmCount := parameters.SwarmCount
	iterationCount := parameters.IterationCount
	switchCount := parameters.SwitchCount
	switchBestParticleCount := parameters.SwitchBestParticleCount

	switchChannel := make([]chan particle, swarmCount)
	switchChannelSemaphore := make([](chan bool), swarmCount)
	doneSwarmChannel := make(chan *swarm)

	for t := 0; t < swarmCount; t++ {

		switchChannel[t] = make(chan particle, swarmCount * switchCount)
		switchChannelSemaphore[t] = make(chan bool, 1)
		switchChannelSemaphore[t] <- true

		go func(id int) {

			// 1) create swarm with N particles

			g := createSwarm(parameters);
			partResultChannel[id] <- *g.CreateCopy()

			switchIteration := switchCount + 1 // switch is between iteration, so + 1 iteration
			for i := 0; i < switchIteration; i++ {

				iterationsPart := iterationCount / switchIteration;
				if (i == switchIteration - 1) {
					iterationsPart += iterationCount % switchIteration
				}

				// 2) step N iterations

				for y := 0; y < iterationsPart; y++ {
					iteration(parameters, g)
				}

				// send part result to listener
				partResultChannel[id] <- *g.CreateCopy()

				// 3) switch between swarm M best particles

				// send my best particles to next swarm
				// This was consulted with you after presentation how to multiswarm works
				<-switchChannelSemaphore[(id + i + 1) % swarmCount]
				sort.Sort(g)
				for y := 0; y < switchBestParticleCount; y++ {
					switchChannel[(id + i + 1) % swarmCount] <- g.particles[y]
				}
				switchChannelSemaphore[(id + i + 1) % swarmCount] <- true

				// pickup other best particles
				for y := 0; y < switchBestParticleCount; y++ {
					g.particles[y] = <-switchChannel[id]
				}
			}

			doneSwarmChannel <- g

		}(t);
	}

	var bestPosition Position

	for t := 0; t < swarmCount; t++ {
		g := <-doneSwarmChannel
		if (t == 0 || g.fcost(g.best) < g.fcost(bestPosition)) {
			bestPosition = g.best
		}

	}

	return bestPosition

}
