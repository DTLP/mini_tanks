package actors

import (
	"os"
	// "fmt"
)

var MenuStage = "init"

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
	exitButton = Coordinates{
		X: 700.0,
		Y: 875.0,
		Width: 250.0,
		Height: 74.0,
	}
}

func MainMenu(tanks *[]Tank, levelNum *int) {
	maxEnemies = 0
	checkIfMenuButtonIsSelected(tanks, levelNum)
}

func checkIfMenuButtonIsSelected(tanks *[]Tank, levelNum *int) {
	for ti, t := range *tanks {
		for pi, p := range t.Projectiles {
			pX := p.X / gameLogicToScreenXOffset
			pY := p.Y / gameLogicToScreenYOffset

			if MenuStage == "init" {
				// Play button
				if checkMenuCollision(pX, pY, playButton.X, playButton.Y, playButton.X+playButton.Width, playButton.Y+playButton.Height) {
					(*tanks)[ti].Projectiles[pi].Collided = true
					MenuStage = "play"
					continue
				}
			}
			if MenuStage == "play" {
				// Solo button
				if checkMenuCollision(pX, pY, soloButton.X, soloButton.Y, soloButton.X+soloButton.Width, soloButton.Y+soloButton.Height) {
					(*tanks)[ti].Projectiles[pi].Collided = true
					// Start game
					*levelNum = 1
					ResetPlayerPositions(tanks)
					maxEnemies = 3
					continue
				}
				// Coop button - not available yet
				// if checkMenuCollision(pX, pY, coopButton.X, coopButton.Y, coopButton.X+coopButton.Width, coopButton.Y+coopButton.Height) {
				// 	(*tanks)[ti].Projectiles[pi].Collided = true
				// 	MenuStage = "coop"
				// 	continue
				// }
			}
			if MenuStage == "coop" {
				// Check if p.X and p.Y are within hostButton rectangle
				if checkMenuCollision(pX, pY, hostButton.X, hostButton.Y, hostButton.X+hostButton.Width, hostButton.Y+hostButton.Height) {

					continue
				}
				// Check if p.X and p.Y are within joinButton rectangle
				if checkMenuCollision(pX, pY, joinButton.X, joinButton.Y, joinButton.X+joinButton.Width, joinButton.Y+joinButton.Height) {

					continue
				}
			}
			// Check if p.X and p.Y are within exitButton rectangle
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