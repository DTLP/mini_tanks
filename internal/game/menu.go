package game

import (
	"github.com/DTLP/mini_tanks/internal/actors"

	"os"
	// "fmt"
	// "time"
)

var menuStage = "init"

type Coordinates struct {
	X float64
	Y float64
	Width float64
	Height float64
}

var (
	playButton Coordinates
	soloButton Coordinates
	coopButton Coordinates
	hostButton Coordinates
	joinButton Coordinates
	backButton Coordinates
	exitButton Coordinates
)

func init () {
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
	checkIfMenuButtonIsSelected(tanks, levelNum)
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
					menuStage = "host"
					// hostNewCoopGame()
					go startServer()
					continue
				}
				// Join button
				if checkMenuCollision(pX, pY, joinButton.X, joinButton.Y, joinButton.X+joinButton.Width, joinButton.Y+joinButton.Height) {
					(*tanks)[ti].Projectiles[pi].Collided = true
					menuStage = "join"
					// joinNewCoopGame()
					go joinServer()
					continue
				}
				// Back button
				if checkMenuCollision(pX, pY, backButton.X, backButton.Y, backButton.X+backButton.Width, backButton.Y+backButton.Height) {
					(*tanks)[ti].Projectiles[pi].Collided = true
					menuStage = "play"
					continue
				}
			}
			if menuStage == "host" || menuStage == "join" {
				// Add message:
				// Waiting for your friend to connect
				// Give them this IP address: 
				if actors.CountPlayerTanks(*tanks) == 2 {
					// Play button
					if checkMenuCollision(pX, pY, playButton.X, playButton.Y, playButton.X+playButton.Width, playButton.Y+playButton.Height) {
						(*tanks)[ti].Projectiles[pi].Collided = true
						// Start game
						actors.ResetPlayerPositions(tanks)
						*levelNum = 1
						actors.MaxEnemies = 3
						}
				}
				// Back button
				if checkMenuCollision(pX, pY, backButton.X, backButton.Y, backButton.X+backButton.Width, backButton.Y+backButton.Height) {
					(*tanks)[ti].Projectiles[pi].Collided = true
					menuStage = "coop"
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