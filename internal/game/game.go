package game

import (
	"github.com/DTLP/mini_tanks/internal/levels"
	"github.com/DTLP/mini_tanks/internal/actors"
	"github.com/DTLP/mini_tanks/internal/scene"

	"fmt"
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
	player       = 1
	levelNum     = 0
	levelObjects = []levels.LevelBlock{}

	gameOver     = false
	gameMode     = 0
)


type Game struct {
	Tanks 		  []actors.Tank
	levelObjects  []levels.LevelBlock
}

func init() {
	//
	fmt.Printf("")
}

func NewGame() *Game {
	var tanks []actors.Tank
	tanks = append(tanks, actors.NewTank("player1"))

    g := &Game{
		Tanks: tanks,
		levelObjects: []levels.LevelBlock{},

    }

	levelObjects = levels.GetLevelObjects(levelNum)

    return g
}


func (g *Game) Update() error {

	// Main Menu logic
	if levelNum == 0 {
		mainMenu(&g.Tanks, &levelNum)
		if levelNum == 1 {
			levelObjects = levels.GetLevelObjects(levelNum)
		}
	}

	// Coop login
	if gameMode == 1 && !doesPlayer2Exist(g.Tanks) {
		g.Tanks = append(g.Tanks, actors.NewTank("player2"))
	}
	
	// Local: Read player input
	for i, _ := range g.Tanks {
		if g.Tanks[i].Player == player {
			actors.HandleMovement(&g.Tanks[i])
		}
	}

	// Coop: Client: Send input to server if in coop
	if gameMode == 1 && player == 2 {
		for i, _ := range g.Tanks {
			if g.Tanks[i].Player == 2  {
				sendTankState(g.Tanks[i])
			}
		}
	}
	// Coop: Server: Get input from client concurrently
	if gameMode == 1 && player == 1 {
		go func() {
			processUpdatesFromClient(&g.Tanks)
		}()
	}
	// Coop: Server: Send update to client
	if gameMode == 1 && player == 1 {
		go func() {
			// Send level num
			sendLevelNumtoClient(levelNum)

			// Create a copy of the slice
			tanksCopy := make([]actors.Tank, len(g.Tanks))
			copy(tanksCopy, g.Tanks)

			// Send tanks
			for i := 0; i < len(tanksCopy); i++ {
				if tanksCopy[i].Player != 2 {
					sendTankState(tanksCopy[i])
				}
			}

			// Send level objects
			for i := range levelObjects {
				sendLevelObjectToClient(levelObjects[i])
			}
		}()
	}
	// Coop: Client: Get update from server
	if gameMode == 1 && player == 2 {
		go func() {
			processUpdatesFromServer(&g.Tanks, &levelNum, &levelObjects)
		}()
	}
	

	// Tank hull and projectile collisions
	actors.HandleCollision(&g.Tanks, levelObjects)

	actors.Update(&g.Tanks)

	// Only host is processing enemy logic
	if player == 1 {
		actors.UpdateEnemyLogic(&g.Tanks, levelObjects)
	}

	g.Tanks = actors.GetUpdatedTankList(g.Tanks)

	// Spawn enemies if more needed
	g.Tanks = actors.CheckEnemyCount(g.Tanks)

	// If all enemies are dead, get new level layout
	if actors.NoEnemiesLeft(g.Tanks) && levelNum != 0 {
		// Progress to next level, reset tanks, get new level layout
		levelNum += 1
		actors.ResetPlayerPositions(&g.Tanks)
		// actors.LevelEnemyNames = actors.EnemyNames
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
		scene.DrawMainMenu(g.Tanks, menuStage, screen)
	}

	// Draw actors
	for i := range g.Tanks {
		scene.DrawActors(&g.Tanks[i], screen)
	}

	// Draw level objects
    for _, t := range g.Tanks {
        if t.IsPlayer && t.Player == player {
			scene.DrawLevel(levelObjects, t.X, t.Y, screen)

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
	// actors.LevelEnemyNames = actors.EnemyNames
	levelObjects = levels.GetLevelObjects(levelNum)

	return tanks, levelObjects
}
