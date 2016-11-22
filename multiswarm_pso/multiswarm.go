package multiswarm_pso

import (
	"math/rand"
	"sort"
	"image/color"
	"golang.org/x/image/colornames"
)

type PositionSystem interface {
	init() PositionSystem
	add(p2 PositionSystem) PositionSystem
	addNum(p2 float64) PositionSystem
	sub(p2 PositionSystem) PositionSystem
	mul(p2 PositionSystem) PositionSystem
	mulNum(p2 float64) PositionSystem
}

type World interface {
	zeroPosition() PositionSystem
	guard(position PositionSystem) bool
	randomPosition() PositionSystem
}

type PSOParameters struct {
	SwarmCount, ParticlesCountInSwarm, IterationCount, SwitchCount, SwitchBestParticleCount int
	Function                                                                                FCost
	World                                                                                   World
}

type particle struct {
	position, best, velocity PositionSystem
}

type swarm struct {
	particles []particle
	best      PositionSystem
	fcost     FCost
}

type FCost func(PositionSystem) float64;

func velocityCalc(parameters PSOParameters, swarmInstance *swarm, particle *particle) PositionSystem {
	v := parameters.World.zeroPosition()
	v = v.add(particle.velocity) // interial

	p := particle.best.sub(particle.position)
	v = v.add(p.mulNum(2 * rand.Float64())) // cognitive

	p2 := swarmInstance.best.sub(particle.position) // social
	v = v.add(p2.mulNum(2 * rand.Float64()))
	return v
}

func createSwarm(parameters PSOParameters) *swarm {
	particles := make([]particle, parameters.ParticlesCountInSwarm);

	var gBest PositionSystem;

	for i, _ := range particles {
		particles[i].position = parameters.World.randomPosition()
		particles[i].velocity = parameters.World.randomPosition()
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

func (a *swarm) Len() int {
	return len(a.particles)
}
func (a *swarm) Swap(i, j int) {
	a.particles[i], a.particles[j] = a.particles[j], a.particles[i]
}
func (a *swarm) Less(i, j int) bool {
	return a.fcost(a.particles[i].best) < a.fcost(a.particles[j].best)
}


// iterate and update gBest for swarm
func iteration(parameters PSOParameters, swarmInstance *swarm) {

	for i, particle := range swarmInstance.particles {
		newVelocity := velocityCalc(parameters, swarmInstance, &particle)
		newPosition := particle.position.add(newVelocity);

		// guard
		if (parameters.World.guard(newPosition)) {
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

func MultiSwarm(parameters PSOParameters) PositionSystem {

	swarmCount := parameters.SwarmCount
	iterationCount := parameters.IterationCount
	switchCount := parameters.SwitchCount
	switchBestParticleCount := parameters.SwitchBestParticleCount

	switchChannel := make([]chan particle, swarmCount)
	switchChannelSemaphore := make([](chan bool), swarmCount)
	doneSwarmChannel := make(chan *swarm)

	renderChannels := make([](chan swarm), swarmCount)
	for i := 0; i < swarmCount; i++ {
		renderChannels[i] = make(chan swarm, switchCount + 1)
	}

	renderChannelX := make(chan bool, 1)

	go func() {
		colors := make([]color.RGBA, swarmCount)
		colors[0] = colornames.Blue
		colors[1] = colornames.Green
		colors[2] = colornames.Black
		colors[3] = colornames.Purple

		for i := 0; i < switchCount; i++ {

//			swarmsxx := make([]renderSwarm, swarmCount)
			for y := 0; y < swarmCount; y++ {
				//s :=
				<-renderChannels[y]
				//swarmsxx[y] = renderSwarm{&(s), colors[y]}
			}

			//RenderPoints2(fmt.Sprintf("tmp/%d-s.png", i), 300, 300, swarmsxx[0].swarm, swarmsxx)
		}

		//makeGif("tmp/*-s.png", "out.gif")

		renderChannelX <- true

	}();

	for t := 0; t < swarmCount; t++ {

		switchChannel[t] = make(chan particle, swarmCount * switchCount)
		switchChannelSemaphore[t] = make(chan bool, 1)
		switchChannelSemaphore[t] <- true

		go func(id int) {

			renderFunc := func(g *swarm) {
				copyParticles := make([]particle, len(g.particles))
				for y, p := range g.particles {
					copyParticles[y] = particle{
						p.position,
						p.best,
						p.velocity,
					}
				}

				gc := &swarm{
					copyParticles,
					g.best,
					g.fcost,
				}

				renderChannels[id] <- *gc
			}

			// 1) create swarm with N particles

			g := createSwarm(parameters);
			renderFunc(g)

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

				renderFunc(g)

				// 3) switch between swarm M best particles

				// send my best particles
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

			//fmt.Printf("%d: done\n", id)

			doneSwarmChannel <- g

		}(t);
	}

	var bestPosition PositionSystem

	for t := 0; t < swarmCount; t++ {
		g := <-doneSwarmChannel
		//fmt.Printf("Global best: %4.2f, %4.2f == %4.2f\n", g.best.x, g.best.y, g.fcost(g.best));
		if (t == 0 || g.fcost(g.best) < g.fcost(bestPosition)) {
			bestPosition = g.best
		}

	}

	<-renderChannelX

	return bestPosition

}
