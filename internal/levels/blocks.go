package levels

import (
	// "fmt"
    "math"
	// "github.com/hajimehoshi/ebiten/v2"
)

const (
    ScreenWidth  = 1000
    ScreenHeight = 1000
    padding      = 20
)

type Line struct {
	X1, Y1, X2, Y2 float64
}

func (l *Line) Angle() float64 {
	return math.Atan2(l.Y2-l.Y1, l.X2-l.X1)
}

type Block struct {
	Walls  []Line
}

var Blocks []Block

type LevelBlock struct {
	X, Y       	  float64
	Width, Height float64
	Blocks		  []Block
	Destructible  bool
	Collidable    bool
	Base          bool
	Health     	  int
	Border		  bool
	Image		  Image
}

var LevelObjects []LevelBlock

type Image struct {
	X      int
	Y      int
	Width  int
	Height int
	Path   string
}

func AddLevelBorder(x, y, w, h int) LevelBlock {
	return LevelBlock{
		X:      float64(x),
		Y:      float64(y),
		Width:  float64(w),
		Height: float64(h),
		Image:  Image{
				Path: "resources/empty.png",
		},
		Border: true,
		Blocks: []Block{
			{
				Walls: rect(float64(x), float64(y), float64(w), float64(h)),
			},
		},
	}
}

func AddLevelBlock(x, y int, blockType string) LevelBlock {
    // Default preset
    block := LevelBlock{
        X:            float64(x),
        Y:            float64(y),
        Width:        64.0,
        Height:       64.0,
        Image:        Image{X: 0, Y: 0, Width: 64, Height: 64, Path: "resources/empty.png"},
        Destructible: true,
		Collidable:   true,
		Base:         false,
        Health:       100,
        Border:       false,
        Blocks: []Block{
            {
                Walls: rect(float64(x), float64(y), 64.0, 64.0),
            },
        },
    }

    switch blockType {
	// Bricks
    case "b":
        block.Image.Path   = "resources/Brick_Block_small.png"
	// Camo net
	case "c":
		block.Image.Path   = "resources/camo_net.png"
		block.Destructible = false
		block.Collidable   = false
		// Avoid casting shadow
        block.Blocks = []Block{{Walls: rect(float64(x), float64(y), 0.0, 0.0)}}
	// Player base
    case "e":
        block.Image.Path   = "resources/eagle.png"
		block.Base         = true
		block.Blocks = []Block{{Walls: rect(float64(x), float64(y), 0.0, 0.0)}}
    // Forest
	case "f":
		block.Image.Path   = "resources/leaves.png"
		block.Destructible = false
		block.Collidable   = false
	// Water
	case "w":
		block.Image.Path   = "resources/water.png"
		block.Destructible = false
        block.Blocks = []Block{{Walls: rect(float64(x), float64(y), 0.0, 0.0)}}
    }

    return block
}

func GetLevelObjects(levelNum int) []LevelBlock {
	levelLayout := GetLevelLayout(levelNum)

	var LevelObjects []LevelBlock
    brickWidth 	:= 64
    brickHeight := 64
    
	// Add outer walls
	LevelObjects = append(LevelObjects, AddLevelBorder(padding, padding, ScreenWidth-2*padding, ScreenHeight-2*padding))

    for row, rowStr := range levelLayout {
        for col, char := range rowStr {
			if char != ' ' {
				// Calculate the coordinates based on grid position
				x := int(col*brickWidth + padding)
				y := int(row*brickHeight + padding)

				// Rectangles
				LevelObjects = append(LevelObjects, AddLevelBlock(x, y, string(char)))
			}
        }
    }

	return LevelObjects
}

func rect(x, y, w, h float64) []Line {
	return []Line{
		{x, y, x, y + h},
		{x, y + h, x + w, y + h},
		{x + w, y + h, x + w, y},
		{x + w, y, x, y},
	}
}

func (b Block) Points() [][2]float64 {
	// Get one of the endpoints for all segments,
	// + the startpoint of the first one, for non-closed paths
	var points [][2]float64
	for _, wall := range b.Walls {
		points = append(points, [2]float64{wall.X2, wall.Y2})
	}
	p := [2]float64{b.Walls[0].X1, b.Walls[0].Y1}
	if p[0] != points[len(points)-1][0] && p[1] != points[len(points)-1][1] {
		points = append(points, [2]float64{b.Walls[0].X1, b.Walls[0].Y1})
	}
	return points
}

// Update object health after it gets shot
func SetHealth(brick *LevelBlock, health int) {
	brick.Health = health
}

func DeformBlock(block *LevelBlock, side string) {
// 0 - Left
// X1 Y1
// X2 Y2
// 1 - Bottom
// X1 Y1  X2 Y2
// 2 - Right
// X2 Y2
// X1 Y1 
// 3 - Top
// X2 Y2  X1 Y2

	// Update object cooldinates depending on which side got destroyed
	// X Y Width Height = Image coordinates
	// Walls[x]X1 X2 Y1 Y2 = Collision / Ray casting point coordinates
	switch side {
		case "l":
			// Image coordinates
			block.X += 32
			block.Width -= 32
			block.Image.X += 32
			// Collision / ray casting point coordinates
			block.Blocks[0].Walls[0].X1  += 32
			block.Blocks[0].Walls[0].X2  += 32
			block.Blocks[0].Walls[1].X1  += 32
			block.Blocks[0].Walls[3].X2  += 32
		case "r":
			block.Width -= 32
			block.Image.X = 0
			block.Image.Width -= 32
			block.Blocks[0].Walls[1].X2  -= 32
			block.Blocks[0].Walls[2].X1  -= 32
			block.Blocks[0].Walls[2].X2  -= 32
			block.Blocks[0].Walls[3].X1  -= 32
		case "t":
			block.Y += 32
			block.Height -= 32
			block.Image.Y += 32
			block.Blocks[0].Walls[0].Y1  += 32
			block.Blocks[0].Walls[2].Y2  += 32
			block.Blocks[0].Walls[3].Y1  += 32
			block.Blocks[0].Walls[3].Y2  += 32
		case "b":
			block.Height -= 32
			block.Image.Height -= 32
			block.Blocks[0].Walls[0].Y2  -= 32
			block.Blocks[0].Walls[1].Y1  -= 32
			block.Blocks[0].Walls[1].Y2  -= 32
			block.Blocks[0].Walls[2].Y1  -= 32
	}
}

func UpdateLevelObjects(levelObjects []LevelBlock) {
	// Destroy any block that's lost its width/height
	for i, object := range levelObjects {
		if object.Height == 0 || object.Width == 0 {
			levelObjects[i].Health = 0
		}
	}
}

func IsBaseDestroyed(levelObjects []LevelBlock) bool {
	// Destroy any block that's lost its width/height
	for _, object := range levelObjects {
		if object.Base && object.Health == 0 {
			return true
		}
	}

	return false
}

// PlayerBaseLevelBlock returns a LevelBlock instance with player base coordinates
func PlayerBaseLevelBlock() LevelBlock {
	return LevelBlock{
		Blocks: []Block{
			{
				Walls: []Line{
					{468, 916, 468, 916},
					{468, 916, 468, 916},
					{468, 916, 468, 916},
					{468, 916, 468, 916},
				},
			},
		},
	}
}