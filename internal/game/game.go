package game

import (
	"github.com/DTLP/mini_tanks/internal/levels"
	"github.com/DTLP/mini_tanks/internal/actors"
	"github.com/DTLP/mini_tanks/internal/scene"

	// "fmt"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
    ScreenWidth  			 = 1000
    ScreenHeight 			 = 1000
    minXCoordinates 		 = 270
    minYCoordinates 		 = 270
    maxXCoordinates 		 = 4730
    maxYCoordinates 		 = 4730
    gameLogicToScreenXOffset = 5.12
    gameLogicToScreenYOffset = 5.12
    padding      			 = 20
)

var (
	player       = 0
	levelNum     = 0
	levelObjects = []levels.LevelBlock{}

	gameOver     = false
)

type Game struct {
	Tanks 		  []actors.Tank
	levelObjects  []levels.LevelBlock
}

func init() {
	//
}

func NewGame() *Game {
	var tanks []actors.Tank
	tanks = append(tanks, actors.NewTank("player1"))

    g := &Game{
		Tanks: tanks,
		levelObjects: []levels.LevelBlock{},

    }

    return g
}


func (g *Game) Update() error {

	// Main Menu logic
	if levelNum == 0 {
		// actors.TotalEnemiesNum = 0

		actors.MainMenu(&g.Tanks, &levelNum)
	}

	levelObjects = levels.GetLevelObjects(levelNum)

	// Read player input
	for i, _ := range g.Tanks {
		if g.Tanks[i].Player {
			actors.HandleMovement(&g.Tanks[i])
		}
	}

	// Tank hull and projectile collisions
	actors.HandleCollision(&g.Tanks, levelObjects)

	actors.Update(&g.Tanks)

	actors.UpdateEnemyLogic(&g.Tanks, levelObjects)

	g.Tanks = actors.GetUpdatedTankList(g.Tanks)

	// Spawn enemies if more needed
	g.Tanks = actors.CheckEnemyCount(g.Tanks)

	// If all enemies are dead, get new level layout
	if actors.NoEnemiesLeft(g.Tanks) && levelNum != 0 {
		// Progress to next level, reset tanks, get new level layout
		levelNum += 1
		actors.ResetPlayerPositions(&g.Tanks)
		levelObjects = levels.GetLevelObjects(levelNum)
	}

	// Deform / remove level objects
	levels.UpdateLevelObjects(levelObjects)

	// Check Game Over conditions
	if actors.NoPlayersLeft(g.Tanks) || levels.IsBaseDestroyed(levelObjects) {
		gameOver = true
	}

	// During the game over screen
	if gameOver && ebiten.IsKeyPressed(ebiten.KeyEnter){
		g.Tanks, g.levelObjects = restartGame()
	}

    return nil

}

func (g *Game) Draw(screen *ebiten.Image) {
	scene.DrawScreen(levelNum, screen)

	// Draw Main Menu
	if levelNum == 0 {
		scene.DrawLevelZero(screen)
	}

	// Draw actors
	for i := range g.Tanks {
		scene.DrawActors(&g.Tanks[i], screen)
	}

	// Draw level objects
    for _, t := range g.Tanks {
        if t.Player {
			scene.DrawLevel(levelObjects, t.Hull.X, t.Hull.Y, screen)

		}
	}

	scene.DrawKillFeed(screen)

	// Check Game Over conditions
	if gameOver {
		scene.DrawGameOverScreen(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {

    return ScreenWidth, ScreenHeight
}

func restartGame() ([]actors.Tank, []levels.LevelBlock) {
    levelNum = 0
	gameOver = false
	actors.ResetCounter()

	var tanks []actors.Tank
	tanks = append(tanks, actors.NewTank("player1"))
	levelObjects = levels.GetLevelObjects(levelNum)

	return tanks, levelObjects
}
