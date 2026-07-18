package game

import (
	"github.com/DTLP/mini_tanks/internal/actors"

	"github.com/hajimehoshi/ebiten/v2"
	"os"
)

var menuStage = "init"

// joinAddr is the address typed in the join_input menu stage.
var joinAddr = "127.0.0.1:54050"

// bsPrev and enterPrev edge-detect Backspace/Enter in the join_input stage.
var bsPrev = false
var enterPrev = false

type Coordinates struct {
	X float64
	Y float64
	Width float64
	Height float64
}

var (
	playButton   Coordinates
	soloButton   Coordinates
	coopButton   Coordinates
	hostButton   Coordinates
	joinButton   Coordinates
	backButton   Coordinates
	exitButton   Coordinates
)

func init() {
	playButton = Coordinates{
		X: 700.0,
		Y: 450.0,
		Width: 250.0,
		Height: 74.0,
	}
	soloButton = Coordinates{
		X: 700.0,
		Y: 450.0,
		Width: 250.0,
		Height: 74.0,
	}
	coopButton = Coordinates{
		X: 700.0,
		Y: 550.0,
		Width: 250.0,
		Height: 74.0,
	}
	hostButton = Coordinates{
		X: 700.0,
		Y: 450.0,
		Width: 250.0,
		Height: 74.0,
	}
	joinButton = Coordinates{
		X: 700.0,
		Y: 550.0,
		Width: 250.0,
		Height: 74.0,
	}
	backButton = Coordinates{
		X: 700.0,
		Y: 650.0,
		Width: 250.0,
		Height: 74.0,
	}
	exitButton = Coordinates{
		X: 700.0,
		Y: 875.0,
		Width: 250.0,
		Height: 74.0,
	}
}

func mainMenu(tanks *[]actors.Tank, levelNum *int) {
	actors.MaxEnemies = 0

	switch menuStage {
	case "host_wait":
		handleHostWait(tanks, levelNum)
	case "join_input":
		handleJoinInput()
	case "join_connect":
		handleJoinConnect()
	}

	checkIfMenuButtonIsSelected(tanks, levelNum)
}

// handleHostWait polls the connection and auto-starts the game once the client
// is connected. Escape cancels hosting and returns to the coop menu.
func handleHostWait(tanks *[]actors.Tank, levelNum *int) {
	if netConn == nil {
		menuStage = "coop"
		return
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		netConn.Close()
		netConn = nil
		role = RoleSolo
		menuStage = "coop"
		return
	}

	if netConn.IsConnected() {
		// Connection established: start the game. The host Update branch
		// takes over from here.
		if !doesPlayer2Exist(*tanks) {
			*tanks = append(*tanks, actors.NewTank("player2"))
		}
		actors.ResetPlayerPositions(tanks)
		*levelNum = 1
		actors.MaxEnemies = 3
	}
}

// handleJoinInput samples the keyboard to build the join address. Enter dials
// the host, Backspace edits the address, Escape returns to the coop menu.
func handleJoinInput() {
	joinAddr += string(ebiten.InputChars())

	if ebiten.IsKeyPressed(ebiten.KeyBackspace) {
		if !bsPrev && len(joinAddr) > 0 {
			joinAddr = joinAddr[:len(joinAddr)-1]
		}
	}
	bsPrev = ebiten.IsKeyPressed(ebiten.KeyBackspace)

	if ebiten.IsKeyPressed(ebiten.KeyEnter) {
		if !enterPrev {
			nc, err := Join(joinAddr)
			if err != nil {
				// Lobby dial error: surface as red hint text in the
				// join_input stage, NOT the full-screen disconnect
				// overlay (that's reserved for mid-game disconnects).
				statusMessage = err.Error()
			} else {
				netConn = nc
				role = RoleClient
				menuStage = "join_connect"
				statusMessage = ""
				statusVisible = false
			}
		}
	}
	enterPrev = ebiten.IsKeyPressed(ebiten.KeyEnter)

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		menuStage = "coop"
		statusMessage = ""
	}
}

// handleJoinConnect waits while the client connects. The game starts when the
// first snapshot with levelNum > 0 arrives in the client Update branch.
func handleJoinConnect() {
	if netConn == nil {
		menuStage = "coop"
		return
	}

	if netConn.Err() != nil {
		// Lobby connection error: surface as red hint text in the
		// join_input stage, not the mid-game disconnect overlay.
		statusMessage = netConn.Err().Error()
		netConn.Close()
		netConn = nil
		role = RoleSolo
		menuStage = "join_input"
		return
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		netConn.Close()
		netConn = nil
		role = RoleSolo
		menuStage = "join_input"
		// Reset the address bar to the prefill so re-entering is clean.
		joinAddr = "127.0.0.1:54050"
		statusMessage = ""
	}
}

// startHostMenu starts hosting and enters the host_wait stage.
func startHostMenu(addr string) {
	nc, err := Host(addr)
	if err != nil {
		statusMessage = err.Error()
		statusVisible = true
		return
	}
	netConn = nc
	role = RoleHost
	menuStage = "host_wait"
	statusMessage = ""
	statusVisible = false
}

func checkIfMenuButtonIsSelected(tanks *[]actors.Tank, levelNum *int) {
	for ti, t := range *tanks {
		for pi, p := range t.Projectiles {
			pX := p.X / gameLogicToScreenXOffset
			pY := p.Y / gameLogicToScreenYOffset

			if menuStage == "init" {
				// Play button
				if checkMenuCollision(pX, pY, playButton.X, playButton.Y, playButton.X+playButton.Width, playButton.Y+playButton.Height) {
					(*tanks)[ti].Projectiles[pi].Collided = true
					menuStage = "play"
					continue
				}
			}
			if menuStage == "play" {
				// Solo button
				if checkMenuCollision(pX, pY, soloButton.X, soloButton.Y, soloButton.X+soloButton.Width, soloButton.Y+soloButton.Height) {
					(*tanks)[ti].Projectiles[pi].Collided = true
					// Start game
					actors.ResetPlayerPositions(tanks)
					*levelNum = 1
					actors.MaxEnemies = 3
				}
				// Coop button
				if checkMenuCollision(pX, pY, coopButton.X, coopButton.Y, coopButton.X+coopButton.Width, coopButton.Y+coopButton.Height) {
					(*tanks)[ti].Projectiles[pi].Collided = true
					menuStage = "coop"
					continue
				}
			}
			if menuStage == "coop" {
				// Host button
				if checkMenuCollision(pX, pY, hostButton.X, hostButton.Y, hostButton.X+hostButton.Width, hostButton.Y+hostButton.Height) {
					(*tanks)[ti].Projectiles[pi].Collided = true
					startHostMenu("0.0.0.0:54050")
					continue
				}
// Join button
			if checkMenuCollision(pX, pY, joinButton.X, joinButton.Y, joinButton.X+joinButton.Width, joinButton.Y+joinButton.Height) {
				(*tanks)[ti].Projectiles[pi].Collided = true
				// Reset the address bar to the default prefill each time we
				// enter the join screen so testing is a one-Enter affair.
				joinAddr = "127.0.0.1:54050"
				statusMessage = ""
				menuStage = "join_input"
				continue
			}
				// Back button
				if checkMenuCollision(pX, pY, backButton.X, backButton.Y, backButton.X+backButton.Width, backButton.Y+backButton.Height) {
					(*tanks)[ti].Projectiles[pi].Collided = true
					menuStage = "play"
					continue
				}
			}
			// Exit Game button
			if checkMenuCollision(pX, pY, exitButton.X, exitButton.Y, exitButton.X+exitButton.Width, exitButton.Y+exitButton.Height) {
				// Close game
				os.Exit(0)
			}
		}
	}
}

func checkMenuCollision(pX, pY, x1, y1, x2, y2 float64) bool {
	return pX > x1 && pX < x2 && pY > y1 && pY < y2
}
