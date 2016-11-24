package multiswarm_pso

import "math/rand"

func randInt(min int, max int) int {
	return min + rand.Intn(max - min)
}

type QuadraticWorld struct {
	LeftTopEdgeCornerVector, RightDownEdgeCornerVector Position
}

func (w QuadraticWorld) DimensionSize() int {
	return len(w.LeftTopEdgeCornerVector)
}

func (w QuadraticWorld) Guard(position Position) bool {
	for i, v := range position {
		if (v < w.LeftTopEdgeCornerVector[i] || v > w.RightDownEdgeCornerVector[i]) {
			return false
		}
	}
	return true
}

func (w QuadraticWorld) RandomPosition() Position {
	v := make(Position, w.DimensionSize())
	for i := 0; i < w.DimensionSize(); i++ {
		v[i] = float64(randInt(int(w.LeftTopEdgeCornerVector[i]), int(w.RightDownEdgeCornerVector[i])))
	}

	return v
}
