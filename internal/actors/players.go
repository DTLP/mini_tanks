package actors

import (
    // "fmt"
    "math"
    "github.com/hajimehoshi/ebiten/v2"
)

const (
    ScreenWidth              = 1000
    ScreenHeight             = 1000
    minXCoordinates          = 270
    minYCoordinates          = 270
    maxXCoordinates          = 4900
    maxYCoordinates          = 4900
    gameLogicToScreenXOffset = 5.12
    gameLogicToScreenYOffset = 5.12

    degrees     = 180.0 / math.Pi
    radians     = degrees * math.Pi / 180.0

    baseOffsetX = 160.0 // float64(baseImage.Bounds().Dx()) / 2
	baseOffsetY = 160.0 // float64(baseImage.Bounds().Dy()) / 2
)


// Define a projectile struct
type Projectile struct {
    X         float64
    Y         float64
    VelocityX float64
    VelocityY float64
    Angle     float64
    Width     float64
    Height    float64
    Collided  bool
}

type Explosion struct {
    X     float64
    Y     float64
    Frame int
}


func HandleMovement(t *Tank) {
    if t.CanMove {
        if ebiten.IsKeyPressed(ebiten.KeyA) {
            t.Hull.Angle -= t.Hull.RotationSpeed
            updateCollisionBox(t)
        }
        if ebiten.IsKeyPressed(ebiten.KeyD) {
            t.Hull.Angle += t.Hull.RotationSpeed
            updateCollisionBox(t)
        }
        if ebiten.IsKeyPressed(ebiten.KeyW) {
            t.Hull.PrevX = t.Hull.X
            t.Hull.PrevY = t.Hull.Y
            t.Hull.X += t.Hull.Speed * math.Cos(-t.Hull.Angle*math.Pi/180.0)
            t.Hull.Y += t.Hull.Speed * math.Sin(t.Hull.Angle*math.Pi/180.0)
            updateCollisionBox(t)
        }
        if ebiten.IsKeyPressed(ebiten.KeyS) {
            t.Hull.PrevX = t.Hull.X
            t.Hull.PrevY = t.Hull.Y
            t.Hull.X -= t.Hull.ReverseSpeed * math.Cos(-t.Hull.Angle*math.Pi/180.0)
            t.Hull.Y -= t.Hull.ReverseSpeed * math.Sin(t.Hull.Angle*math.Pi/180.0)
            updateCollisionBox(t)
        }

        if ebiten.IsKeyPressed(ebiten.KeyLeft) {
            t.Turret.Angle -= t.Turret.RotationSpeed
        }
        if ebiten.IsKeyPressed(ebiten.KeyRight) {
            t.Turret.Angle += t.Turret.RotationSpeed
        }
    
        // Ensure the tank stays within the game world bounds
        if t.Hull.X < minXCoordinates {
            t.Hull.X = minXCoordinates
        }
        if t.Hull.X > maxXCoordinates {
            t.Hull.X = maxXCoordinates
        }
        if t.Hull.Y < minYCoordinates {
            t.Hull.Y = minYCoordinates
        }
        if t.Hull.Y > maxYCoordinates {
            t.Hull.Y = maxYCoordinates
        }
    
        // Update the turret's position relative to the base
        t.Turret.X = t.Hull.X
        t.Turret.Y = t.Hull.Y
    }

    if ebiten.IsKeyPressed(ebiten.KeySpace) && t.Turret.ReloadTimer == 0 {
        shoot(t)
    }
}

// updateCollisionBox updates the tank's collision box based on its position and rotation.
func updateCollisionBox (t *Tank) {

    // Offset from the center of the tank's base
    offsetX := float64(t.Hull.Width) / 2
    offsetY := float64(t.Hull.Height) / 2

    // Convert tank's game logic coordinates to screen coordinates
    tankXScreen := t.Hull.X / gameLogicToScreenXOffset
    tankYScreen := t.Hull.Y / gameLogicToScreenYOffset

    // Calculate the rotation angle in radians
    angleRad := t.Hull.Angle * math.Pi / 180

    // Update the collision coordinates based on the tank's current position and rotation
    t.Hull.CollisionX1 = tankXScreen - offsetX*math.Cos(angleRad) + offsetY*math.Sin(angleRad)
    t.Hull.CollisionY1 = tankYScreen - offsetX*math.Sin(angleRad) - offsetY*math.Cos(angleRad)

    t.Hull.CollisionX2 = tankXScreen + offsetX*math.Cos(angleRad) + offsetY*math.Sin(angleRad)
    t.Hull.CollisionY2 = tankYScreen + offsetX*math.Sin(angleRad) - offsetY*math.Cos(angleRad)

    t.Hull.CollisionX3 = tankXScreen + offsetX*math.Cos(angleRad) - offsetY*math.Sin(angleRad)
    t.Hull.CollisionY3 = tankYScreen + offsetX*math.Sin(angleRad) + offsetY*math.Cos(angleRad)

    t.Hull.CollisionX4 = tankXScreen - offsetX*math.Cos(angleRad) - offsetY*math.Sin(angleRad)
    t.Hull.CollisionY4 = tankYScreen - offsetX*math.Sin(angleRad) + offsetY*math.Cos(angleRad)

}

func NoPlayersLeft(tanks []Tank) bool {
    if CountPlayerTanks(tanks) == 0 {
        return true
    }

    return false
}

func CountPlayerTanks(tanks []Tank) int {
	count := 0
	for _, tank := range tanks {
		if tank.Player && tank.Health > 0 {
			count++
		}
	}

	return count
}

func ResetPlayerPositions(tanks *[]Tank) {
    for i := range *tanks {
        t := &(*tanks)[i]
		if t.Name == "player1" {
			t.Hull.X   = 1850.0
            t.Hull.Y   = 4730.0
		}
        if t.Name == "player2" {
			t.Hull.X   = 3280.0
            t.Hull.Y   = 4730.0
		}
        t.Hull.Angle   = -90.0
        t.Turret.Angle = -90.0
        t.Health       = t.MaxHealth
        updateCollisionBox(t)
	}
}
