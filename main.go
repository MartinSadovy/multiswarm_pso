package main

import (
	"math"
	crand "crypto/rand"
	"math/big"
	"math/rand"
	"fmt"
	"runtime"
	"sort"
	"github.com/gonum/matrix/mat64"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot"
	"log"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/palette"
	"golang.org/x/image/colornames"
	"image/color"
	"image"
	"bytes"
	"image/gif"
	"path/filepath"
	"github.com/disintegration/imaging"
	"os"
	_ "image/jpeg"
	_ "image/png"
)

type positionXY struct {
	x, y float64
}

func (p *positionXY) add(p2 positionXY) positionXY {
	return positionXY{x: p.x + p2.x, y: p.y + p2.y}
}

func (p *positionXY) addNum(p2 float64) positionXY {
	return positionXY{x: p.x + p2, y: p.y + p2}
}

func (p *positionXY) sub(p2 positionXY) positionXY {
	return positionXY{x: p.x - p2.x, y: p.y - p2.y}
}

func (p *positionXY) mul(p2 positionXY) positionXY {
	return positionXY{x: p.x * p2.x, y: p.y * p2.y}
}

func (p *positionXY) mulNum(p2 float64) positionXY {
	return positionXY{x: p.x * p2, y: p.y * p2}
}

type particle struct {
	position, best, velocity positionXY
}

type swarm struct {
	particles                  []particle
	leftTop, rightBottom, best positionXY
	fcost                      fcost
}

type fcost func(positionXY) float64;

func randInt(min int, max int) int {
	return min + rand.Intn(max - min)
}

func velocityCalc(swarmInstance *swarm, particle *particle) positionXY {
	v := positionXY{x: 0, y: 0};
	v = v.add(particle.velocity) // interial

	p := particle.best.sub(particle.position)
	v = v.add(p.mulNum(2 * rand.Float64())) // cognitive

	p2 := swarmInstance.best.sub(particle.position) // social
	v = v.add(p2.mulNum(2 * rand.Float64()))
	return v
}

func createSwarm(leftTop, rightBottom positionXY, particlesCount int, fcost fcost) *swarm {
	particles := make([]particle, particlesCount);

	var gBest positionXY;

	for i, _ := range particles {
		particles[i].position.x = float64(randInt(int(leftTop.x), int(rightBottom.x)))
		particles[i].position.y = float64(randInt(int(leftTop.y), int(rightBottom.y)))
		particles[i].velocity = positionXY{x: float64(randInt(-10, 10)), y: float64(randInt(-10, 10))}
		particles[i].best = particles[i].position;
		if (i == 0 || fcost(particles[i].best) < fcost(gBest)) {
			gBest = particles[i].best;
		}
	}

	var b *swarm
	b = &swarm{
		particles,
		leftTop,
		rightBottom,
		gBest,
		fcost,
	};
	return b
}

// iterate and update gBest for swarm
func iteration(swarmInstance *swarm) {

	for i, particle := range swarmInstance.particles {
		newVelocity := velocityCalc(swarmInstance, &particle)
		newPosition := particle.position.add(newVelocity);

		// guard
		if (
			newPosition.x >= swarmInstance.leftTop.x && newPosition.x <= swarmInstance.rightBottom.x &&
				newPosition.y >= swarmInstance.leftTop.y && newPosition.y <= swarmInstance.rightBottom.y) {
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


func (a *swarm) Len() int           { return len(a.particles) }
func (a *swarm) Swap(i, j int)      { a.particles[i], a.particles[j] = a.particles[j], a.particles[i] }
func (a *swarm) Less(i, j int) bool { return a.fcost(a.particles[i].best) < a.fcost(a.particles[j].best) }

type PSOParameters struct {
	swarmCount, particlesCountInSwarm, iterationCount, switchCount, switchBestParticleCount int
	fnc fcost
	leftTopEdgeCorner, rightDownEdgeCorner positionXY
}

func makeGif(path string, output string) {
	srcfilenames, err := filepath.Glob(path)
	if err != nil {
		log.Fatalf("Error in globbing source file pattern : %s", err)
	}

	var frames []*image.Paletted
	for _, filename := range srcfilenames {
		img, err := imaging.Open(filename)

		buf := bytes.Buffer{}
		if err := gif.Encode(&buf, img, nil); err != nil {
			log.Printf("Skipping file %s due to error in gif encoding:%s", filename, err)
			continue
		}

		tmpimg, err := gif.Decode(&buf)
		if err != nil {
			log.Printf("Skipping file %s due to weird error reading the temporary gif :%s", filename, err)
			continue
		}
		frames = append(frames, tmpimg.(*image.Paletted))
	}

	delays := make([]int, len(frames))
	for j, _ := range delays {
		delays[j] = 50
	}

	opfile, err := os.Create(output)
	if err != nil {
		log.Fatalf("Error creating the destination file out.gif : %s", err)
	}

	gifs := &gif.GIF{Image: frames, Delay: delays}
	if err := gif.EncodeAll(opfile, gifs); err != nil {
		log.Printf("Error encoding output into animated gif :%s", err)
	}
	opfile.Close()
}

func multiSwarm(parameters PSOParameters) positionXY {

	swarmCount := parameters.swarmCount
	particlesCountInSwarm := parameters.particlesCountInSwarm
	iterationCount := parameters.iterationCount
	switchCount := parameters.switchCount
	switchBestParticleCount := parameters.switchBestParticleCount

	fnc := parameters.fnc
	leftTopEdgeCorner := parameters.leftTopEdgeCorner
	rightDownEdgeCorner := parameters.rightDownEdgeCorner

	switchChannel := make([]chan particle, swarmCount)
	switchChannelSemaphore := make([](chan bool), swarmCount)
	doneSwarmChannel := make(chan *swarm)

	renderChannels := make([](chan swarm), swarmCount)
	for i:= 0; i < swarmCount; i++ {
		renderChannels[i] = make(chan swarm, switchCount+1)
	}

	renderChannelX := make(chan bool, 1)

	go func (){
		colors := make([]color.RGBA, swarmCount)
		colors[0] = colornames.Blue
		colors[1] = colornames.Green
		colors[2] = colornames.Black
		colors[3] = colornames.Purple

		for i := 0; i < switchCount; i++ {

			swarmsxx := make([]renderSwarm, swarmCount)
			for y := 0; y < swarmCount; y++ {
				s := <-renderChannels[y]
				swarmsxx[y] = renderSwarm{&(s), colors[y]}
			}

			RenderPoints2(fmt.Sprintf("tmp/%d-s.png", i), 300, 300, swarmsxx[0].swarm, swarmsxx)
		}

		makeGif("tmp/*-s.png", "out.gif")

		renderChannelX <- true

	}();

	for t := 0; t < swarmCount; t++ {

		switchChannel[t] = make(chan particle, swarmCount*switchCount)
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
					g.leftTop,
					g.rightBottom,
					g.best,
					g.fcost,
				}

				renderChannels[id] <- *gc
			}

			// 1) create swarm with N particles

			g := createSwarm(leftTopEdgeCorner, rightDownEdgeCorner, particlesCountInSwarm, fnc);
			renderFunc(g)

			switchIteration := switchCount + 1 // switch is between iteration, so + 1 iteration
			for i := 0; i < switchIteration; i++ {

				iterationsPart := iterationCount / switchIteration;
				if (i == switchIteration - 1) {
					iterationsPart += iterationCount % switchIteration
				}

				// 2) step N iterations

				for y := 0; y < iterationsPart; y++ {
					iteration(g)
				}

				renderFunc(g)

				// 3) switch between swarm M best particles

				// send my best particles
				<- switchChannelSemaphore[(id + i + 1) % swarmCount]
				sort.Sort(g)
				for y := 0; y < switchBestParticleCount; y++ {
					switchChannel[(id + i + 1) % swarmCount] <- g.particles[y]
				}
				switchChannelSemaphore[(id + i + 1) % swarmCount] <- true

				// pickup other best particles
				for y := 0; y < switchBestParticleCount; y++ {
					g.particles[y] = <- switchChannel[id]
				}
			}

			fmt.Printf("%d: done\n", id)

			doneSwarmChannel <- g

		}(t);
	}

	var bestPosition positionXY

	for t := 0; t < swarmCount; t++ {
		g := <-doneSwarmChannel
		fmt.Printf("Global best: %4.2f, %4.2f == %4.2f\n", g.best.x, g.best.y, g.fcost(g.best));
		if (t == 0 || g.fcost(g.best) < g.fcost(bestPosition)) {
			bestPosition = g.best
		}

	}

	<- renderChannelX

	return bestPosition

}

type offsetUnitGrid struct {
	Data mat64.Matrix
}

func (g offsetUnitGrid) Dims() (c, r int) {
	r, c = g.Data.Dims(); return c, r
}
func (g offsetUnitGrid) Z(c, r int) float64 {
	return g.Data.At(r, c)
}
func (g offsetUnitGrid) X(c int) float64 {
	_, n := g.Data.Dims()
	if c < 0 || c >= n {
		panic("index out of range")
	}
	return float64(c - n / 2)
}
func (g offsetUnitGrid) Y(r int) float64 {
	m, _ := g.Data.Dims()
	if r < 0 || r >= m {
		panic("index out of range")
	}
	return float64(r - m / 2)
}

type renderSwarm struct {
	swarm *swarm
	color color.Color
}

func RenderPoints2(filename string, width, height int, swarm *swarm, swarms []renderSwarm) {

	p, err := plot.New()
	if err != nil {
		log.Panic(err)
	}
	p.Title.Text = "Heat map"
	p.X.Min = swarm.leftTop.x
	p.Y.Min = swarm.leftTop.y
	p.X.Max = swarm.rightBottom.x
	p.Y.Max = swarm.rightBottom.y

	boardWidth := int(math.Ceil(swarm.rightBottom.x - swarm.leftTop.x))
	boardHeight := int(math.Ceil(swarm.rightBottom.y - swarm.leftTop.y))

	// heat map

	v := make([]float64, boardWidth*boardHeight);
	for i := 0; i < boardWidth*boardHeight; i++ {
		v[i] = swarm.fcost(positionXY{swarm.leftTop.x+float64(i % boardWidth), swarm.leftTop.y+float64(i / boardWidth)})
	}
	m := offsetUnitGrid{Data: mat64.NewDense(boardWidth, boardHeight, v)}
	h := plotter.NewHeatMap(m, palette.Heat(12, 1))
	p.Add(h)

	// points

	makePoints := func(points []particle) plotter.XYs {
		pts := make(plotter.XYs, len(points))
		for i := range points {
			pts[i].X = float64(points[i].position.x)
			pts[i].Y = float64(points[i].position.y)
		}
		return pts
	}

	for _, swarmRender := range swarms {
		scatterData := makePoints(swarmRender.swarm.particles)
		s, err := plotter.NewScatter(scatterData)
		if err != nil {
			log.Panic(err)
		}
		s.Color = swarmRender.color
		p.Add(s)
	}

	err = p.Save(vg.Length(width), vg.Length(height), filename)
	if err != nil {
		log.Panic(err)
	}
}


func dejonge(position positionXY) float64 {
	return math.Pow(float64(position.x), 2.0) + math.Pow(float64(position.y), 2.0);
}

func schwefel(position positionXY) float64 {
	return (-position.x*math.Sin(math.Sqrt(math.Abs(position.x)))) + (-position.y*math.Sin(math.Sqrt(math.Abs(position.y))))
}

func ackley(position positionXY) float64 {
	return 1/math.Pow(math.E, 5) * math.Sqrt(position.x*position.x + position.y*position.y) + 3 * (math.Cos(2*position.x)+math.Sin(2*position.y))
}


func main() {
	var v, _ = crand.Int(crand.Reader, big.NewInt(1000))
	rand.Seed(v.Int64())

	parameters := PSOParameters{}
	parameters.swarmCount = runtime.NumCPU()

	parameters.particlesCountInSwarm = 20
	parameters.iterationCount = 100

	parameters.switchBestParticleCount = 5
	parameters.switchCount = 10

	parameters.fnc = schwefel
	parameters.leftTopEdgeCorner = positionXY{-500, -500}
	parameters.rightDownEdgeCorner = positionXY{500, 500}

	multiSwarm(parameters)
}
