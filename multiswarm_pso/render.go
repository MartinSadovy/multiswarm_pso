package multiswarm_pso
/*
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
)

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
/*
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
		v[i] = swarm.fcost(PositionXY{swarm.leftTop.x+float64(i % boardWidth), swarm.leftTop.y+float64(i / boardWidth)})
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
*/