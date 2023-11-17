package scene

import (
	"github.com/DTLP/mini_tanks/internal/levels"

	"math"
	"sort"
	"github.com/hajimehoshi/ebiten/v2"
)

// rayCasting returns a slice of levels.Line originating from point cx, cy and intersecting with blocks
func rayCasting(cx, cy float64, blocks []levels.Block) []levels.Line {
	const rayLength = 10000 // something large enough to reach all blocks

	var rays []levels.Line
	for _, obj := range blocks {
		// Cast two rays per point
		for _, p := range obj.Points() {

			l := levels.Line{cx, cy, p[0], p[1]}
			angle := l.Angle()

			for _, offset := range []float64{-0.005, 0.005} {
				points := [][2]float64{}
				ray := newRay(cx, cy, rayLength, angle+offset)

				// Unpack all blocks
				for _, o := range blocks {
					for _, wall := range o.Walls {

						if px, py, ok := intersection(ray, wall); ok {

							points = append(points, [2]float64{px, py})

						}
					}
				}
				// Find the point closest to start of ray
				min := math.Inf(1)
				minI := -1
				for i, p := range points {
					d2 := (cx-p[0])*(cx-p[0]) + (cy-p[1])*(cy-p[1])
					if d2 < min {
						min = d2
						minI = i
					}
				}
				rays = append(rays, levels.Line{cx, cy, points[minI][0], points[minI][1]})
			}
		}
	}

	// Sort rays based on angle, otherwise light triangles will not come out right
	sort.Slice(rays, func(i int, j int) bool {
		return rays[i].Angle() < rays[j].Angle()
	})
	return rays
}

func newRay(x, y, length, angle float64) levels.Line {
	return levels.Line{
		X1: x,
		Y1: y,
		X2: x + length*math.Cos(angle),
		Y2: y + length*math.Sin(angle),
	}
}

// intersection calculates the intersection of given two levels.Lines.
func intersection(l1, l2 levels.Line) (float64, float64, bool) {
	// https://en.wikipedia.org/wiki/levels.Line%E2%80%93levels.Line_intersection#Given_two_points_on_each_levels.Line
	denom := (l1.X1-l1.X2)*(l2.Y1-l2.Y2) - (l1.Y1-l1.Y2)*(l2.X1-l2.X2)
	tNum := (l1.X1-l2.X1)*(l2.Y1-l2.Y2) - (l1.Y1-l2.Y1)*(l2.X1-l2.X2)
	uNum := -((l1.X1-l1.X2)*(l1.Y1-l2.Y1) - (l1.Y1-l1.Y2)*(l1.X1-l2.X1))

	if denom == 0 {
		return 0, 0, false
	}

	t := tNum / denom
	if t > 1 || t < 0 {
		return 0, 0, false
	}

	u := uNum / denom
	if u > 1 || u < 0 {
		return 0, 0, false
	}

	x := l1.X1 + t*(l1.X2-l1.X1)
	y := l1.Y1 + t*(l1.Y2-l1.Y1)

	return x, y, true
}

func rayVertices(x1, y1, x2, y2, x3, y3 float64) []ebiten.Vertex {
	return []ebiten.Vertex{
		{DstX: float32(x1), DstY: float32(y1), SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(x2), DstY: float32(y2), SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{DstX: float32(x3), DstY: float32(y3), SrcX: 0, SrcY: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}
}