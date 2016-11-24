package multiswarm_pso

import (
	"github.com/gonum/plot"
	"log"
	"math"
	"github.com/gonum/matrix/mat64"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"os"
	"image/gif"
	"image/color"
	"path/filepath"
	"image"
	"github.com/disintegration/imaging"
	"github.com/gonum/plot/palette"
	"bytes"
	"golang.org/x/image/colornames"
	"fmt"
	"io/ioutil"
	"strconv"
)

/**
 * THIS SCRIPT is BONUS for VISUALIZATION! ;)
 */

// this file creates gif from files (by glob(path))
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

// Create pipeline which waiting a part-result and creates after that gif and on second return value is chan which indicate gif si done
func CreateRenderPipe(filename string, parameters PSOParameters) ([](chan swarm), chan bool) {

	swarmCount := parameters.SwarmCount
	switchCount := parameters.SwitchCount

	partResultChannel := make([](chan swarm), swarmCount)
	for i := 0; i < swarmCount; i++ {
		partResultChannel[i] = make(chan swarm, switchCount + 1)
	}

	renderDoneChannel := make(chan bool, 1)

	tmpDir, err := ioutil.TempDir("", "pso-sim-gif" + strconv.Itoa(randInt(1, 1000)))
	if err != nil {
		log.Panic(err)
	}

	go func() {
		colors := make([]color.RGBA, swarmCount)
		renderPointsWaiter := make(chan bool, switchCount)
		for i := 0; i < swarmCount; i++ {
			colors[i] = colornames.Map[colornames.Names[i * 17 % len(colornames.Names)]]
		}
		for i := 0; i < switchCount; i++ {

			swarmsxx := make([]renderSwarm, swarmCount)
			for y := 0; y < swarmCount; y++ {
				s :=
					<-partResultChannel[y]
				swarmsxx[y] = renderSwarm{&(s), colors[y]}
			}

			go func (i int) {
				if (filename != "") {
					renderPoints2(filepath.Clean(fmt.Sprintf("%s/%d.png", tmpDir, i)), 300, 300, &parameters.World, swarmsxx[0].swarm, swarmsxx)
				}
				renderPointsWaiter <- true
			}(i)
		}

		for i := 0; i < switchCount; i++ {
			<- renderPointsWaiter
		}

		if (filename != "") {
			makeGif(filepath.Clean(fmt.Sprintf("%s/*.png", tmpDir)), filename)
		}

		renderDoneChannel <- true

		os.RemoveAll(tmpDir)

	}();


	return partResultChannel, renderDoneChannel
}

// generate 2D graph to png file
func renderPoints2(filename string, width, height int, world *World, swarm *swarm, swarms []renderSwarm) {

	w := (*world).(QuadraticWorld)

	if (w.DimensionSize() != 2) {
		return;
	}

	p, err := plot.New()
	if err != nil {
		log.Panic(err)
	}
	p.Title.Text = "Multiswarm on heat map"
	p.X.Min = w.LeftTopEdgeCornerVector[0]
	p.Y.Min = w.LeftTopEdgeCornerVector[1]
	p.X.Max = w.RightDownEdgeCornerVector[0]
	p.Y.Max = w.RightDownEdgeCornerVector[1]

	boardWidth := int(math.Ceil(p.X.Max - p.X.Min))
	boardHeight := int(math.Ceil(p.Y.Max - p.X.Min))

	// heat map

	v := make([]float64, boardWidth*boardHeight);
	for i := 0; i < boardWidth*boardHeight; i++ {
		v[i] = swarm.fcost(Position{p.X.Min+float64(i % boardWidth), p.Y.Min+float64(i / boardWidth)})
	}
	m := offsetUnitGrid{Data: mat64.NewDense(boardWidth, boardHeight, v)}
	h := plotter.NewHeatMap(m, palette.Heat(12, 1))
	p.Add(h)

	// points

	makePoints := func(points []particle) plotter.XYs {
		pts := make(plotter.XYs, len(points))
		for i := range points {
			pts[i].X = float64(points[i].position[0])
			pts[i].Y = float64(points[i].position[1])
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
