package scene

import (
	"github.com/DTLP/mini_tanks/internal/actors"

	// "fmt"
	"math"
    
	"image"
    "time"
    "strings"
	"image/color"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
)

var (
    turretImg, projectileImg, explosionImg      *ebiten.Image
    // baseWidth, baseHeight   	= 64, 64 // Set the size of the tank base
    // turretWidth, turretHeight 	= 64, 64 // Set the size of the turret

    projectileWidth = 32.0
    projectileHeight = 32.0
)

func DrawActors(t *actors.Tank, screen *ebiten.Image) {

    drawTankHull(t, screen)

    drawTankTurret(t, screen)

    drawHealthBar(t, screen)

    drawReloadBar(t, screen)

	drawTankPojectiles(t, screen)

    drawExplosions(t, screen)

	// drawDebugStuffActors(t, screen)
}

func drawHealthBar(t *actors.Tank, screen *ebiten.Image) {
    percentage := float64(t.Health) / float64(t.MaxHealth)

    // Define colors for the progress bar based on the HP percentage
    var filledColor color.RGBA
    if percentage >= 0.75 {
        filledColor = color.RGBA{255, 255, 255, 255} // White
    } else if percentage >= 0.5 {
        filledColor = color.RGBA{255, 255, 0, 255} // Yellow
    } else if percentage >= 0.25 {
        filledColor = color.RGBA{255, 165, 0, 255} // Orange
    } else if percentage > 0 {
        filledColor = color.RGBA{255, 0, 0, 255} // Red
    } else {
        filledColor = color.RGBA{0, 0, 0, 0} // Transparent
    }

    filledWidth := 1 + int(float64(t.HealthBarWidth) * percentage)
    // Draw the filled portion of the progress bar
    filledRect := image.NewRGBA(image.Rect(0, 0, int(filledWidth), int(t.HealthBarHeight)))
    for x := 0; x < int(filledWidth); x++ {
        for y := 0; y < int(t.HealthBarHeight); y++ {
            filledRect.Set(x, y, filledColor)
        }
    }

    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(float64(t.Hull.X / gameLogicToScreenXOffset - 32),
			float64(t.Hull.Y / gameLogicToScreenYOffset + 35))
    screen.DrawImage(ebiten.NewImageFromImage(filledRect), op)

}

func drawReloadBar(t *actors.Tank, screen *ebiten.Image) {
    // Calculate the percentage of reload timer completion
    percentage := float64(t.Turret.ReloadTimer) / float64(t.Turret.ReloadTime)

    // Define colors for the progress bar based on the percentage
    var filledColor color.RGBA
    if percentage >= 0.75 {
        filledColor = color.RGBA{255, 0, 0, 255} // Red
    } else if percentage >= 0.5 {
        filledColor = color.RGBA{255, 165, 0, 255} // Orange
    } else if percentage >= 0.25 {
        filledColor = color.RGBA{255, 255, 0, 255} // Yellow
    } else if percentage > 0 {
        filledColor = color.RGBA{255, 255, 255, 255} // White
    } else {
        filledColor = color.RGBA{0, 0, 0, 0} // Transparent
    }

    filledWidth := 1 + int(float64(t.ReloadBarWidth) * percentage)
    // Draw the filled portion of the progress bar
    filledRect := image.NewRGBA(image.Rect(0, 0, int(filledWidth), int(t.ReloadBarHeight)))
    for x := 0; x < int(filledWidth); x++ {
        for y := 0; y < int(t.ReloadBarHeight); y++ {
            filledRect.Set(x, y, filledColor)
        }
    }

    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(float64(t.Hull.X / gameLogicToScreenXOffset - 32),
			float64(t.Hull.Y / gameLogicToScreenYOffset + 40))
    screen.DrawImage(ebiten.NewImageFromImage(filledRect), op)

}

func drawTankHull(t *actors.Tank, screen *ebiten.Image) {
    op := &ebiten.DrawImageOptions{}
	bodyImg, _, _ := ebitenutil.NewImageFromFile(t.Hull.Image)

    // Calculate the offset to the center of the tank hull
    baseOffsetX := float64(bodyImg.Bounds().Dx()) / 2
    baseOffsetY := float64(bodyImg.Bounds().Dy()) / 2
    // Set the center of rotation to the center of the tank hull
    op.GeoM.Translate(-baseOffsetX, -baseOffsetY)
    // Rotate the hull
    op.GeoM.Rotate(t.Hull.Angle * math.Pi / 180.0)
    // Translate to the final position
    op.GeoM.Translate(t.Hull.X, t.Hull.Y)
    // Scale tank's hull
    op.GeoM.Scale(float64(t.Hull.Width)/float64(bodyImg.Bounds().Dx()), float64(t.Hull.Height)/float64(bodyImg.Bounds().Dy()))
    // Draw tank hull
    screen.DrawImage(bodyImg, op)

    // // Print debugging stuff
	// if t.Player {
	// 	tHullXStr := fmt.Sprintf("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n t.Hull.X: %+v", t.Hull.X)
	// 	ebitenutil.DebugPrint(screen, tHullXStr)
	// 	tHullYStr := fmt.Sprintf("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n t.Hull.Y: %+v", t.Hull.Y)
	// 	ebitenutil.DebugPrint(screen, tHullYStr)
	// 	tHullAngleStr := fmt.Sprintf("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n t.Hull.Angle: %+v", t.Hull.Angle)
	// 	ebitenutil.DebugPrint(screen, tHullAngleStr)

	// 	tx := ebiten.GeoM{}
	// 	tx.Concat(op.GeoM)
	// 	opxStr := fmt.Sprintf("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n   op X: %f", tx.Element(0, 2))
	// 	ebitenutil.DebugPrint(screen, opxStr)
	// 	ty := ebiten.GeoM{}
	// 	ty.Concat(op.GeoM)
	// 	opyStr := fmt.Sprintf("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n   op Y: %f", ty.Element(1, 2))
	// 	ebitenutil.DebugPrint(screen, opyStr)
	// }
}

func drawTankTurret(t *actors.Tank, screen *ebiten.Image) {
    op := &ebiten.DrawImageOptions{}
	turretImg, _, _ := ebitenutil.NewImageFromFile(t.Turret.Image)
    // Calculate the offset to the center of the tank turret
    turretOffsetX := float64(turretImg.Bounds().Dx()) / 2
    turretOffsetY := float64(turretImg.Bounds().Dy()) / 2

    op.GeoM.Translate(-turretOffsetX, -turretOffsetY)
    op.GeoM.Rotate(t.Turret.Angle * math.Pi / 180.0)
    op.GeoM.Translate(t.Turret.X, t.Turret.Y)
    op.GeoM.Scale(float64(t.Turret.Width)/float64(turretImg.Bounds().Dx()), float64(t.Turret.Height)/float64(turretImg.Bounds().Dy()))
    screen.DrawImage(turretImg, op)

    // // Print debugging stuff
	// coordsStr := fmt.Sprintf("Base X: %.2f, Y: %.2f\nTurret X: %.2f, Y: %.2f", t.Hull.X, t.Hull.Y, t.Turret.X, t.Turret.Y)
	// ebitenutil.DebugPrint(screen, coordsStr)
    // angleStr := fmt.Sprintf("\n\nTurret Angle: %.2f°", t.Turret.Angle)
    // ebitenutil.DebugPrint(screen, angleStr)
    // tTurretXStr := fmt.Sprintf("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\nt.Hull.X: %+v", t.Turret.X)
    // ebitenutil.DebugPrint(screen, tTurretXStr)
}

func drawTankPojectiles(t *actors.Tank, screen *ebiten.Image) {
	// turretImg, _, _ := ebitenutil.NewImageFromFile(t.Turret.Image)

    // Calculate the offset to the center of the tank turret
    turretOffsetX := float64(turretImg.Bounds().Dx()) / 2
    turretOffsetY := float64(turretImg.Bounds().Dy()) / 2

    for _, projectile := range t.Projectiles {
        screenX := projectile.X / gameLogicToScreenXOffset
        screenY := projectile.Y / gameLogicToScreenYOffset


        // Spawn projectile at turret's center facing same way as turret
        op := &ebiten.DrawImageOptions{}
        op.GeoM.Translate(-turretOffsetX, -turretOffsetY)
        projectileAngleRad := projectile.Angle * math.Pi / 180.0
        op.GeoM.Rotate(projectileAngleRad)

        // Adjust the translation based on projectile angle
        offsetX := turretOffsetX*math.Cos(projectileAngleRad) - turretOffsetY*math.Sin(projectileAngleRad)
        offsetY := turretOffsetX*math.Sin(projectileAngleRad) + turretOffsetY*math.Cos(projectileAngleRad)

        op.GeoM.Translate(screenX+offsetX, screenY+offsetY)
        screen.DrawImage(projectileImg, op)

        // Debugging: Draw a red square at the projectile's position
        // ebitenutil.DrawRect(screen, screenX, screenY, 5, 5, color.RGBA{255, 0, 0, 255})
    }

    // // Print debugging stuff
    // reloadTimerStr := fmt.Sprintf("\n\n\nt.Turret.ReloadTimer: %.2f", t.Turret.ReloadTimer)
    // ebitenutil.DebugPrint(screen, reloadTimerStr)
    // reloadTimeStr := fmt.Sprintf("\n\n\n\nt.Turret.ReloadTimer: %.2f", t.Turret.ReloadTime)
    // ebitenutil.DebugPrint(screen, reloadTimeStr)
    // projectilesStr := fmt.Sprintf("\n\n\n\n\nNumber of projectiles: %d", len(projectiles))
    // ebitenutil.DebugPrint(screen, projectilesStr)
}

func drawExplosions(t *actors.Tank, screen *ebiten.Image) {
    frameOX     := 0
    frameOY     := 0
    frameWidth 	:= 64
    frameHeight := 64
    frameCount 	:= 16

    for _, expl := range t.Explosions {
        // Ensure that the frame is within bounds
        frameIndex := expl.Frame % frameCount
        if frameIndex < 0 || frameIndex >= frameCount {
            continue
        }

        screenX := expl.X / gameLogicToScreenXOffset
        screenY := expl.Y / gameLogicToScreenYOffset

        op := &ebiten.DrawImageOptions{}
        op.GeoM.Translate(screenX, screenY)

        sy := frameOY + (frameIndex / 4) * frameHeight
        sx := frameOX + (frameIndex % 4) * frameWidth
    
        subImg := explosionImg.SubImage(image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)).(*ebiten.Image)
    
        // Use the sub-image directly without additional translation
        screen.DrawImage(subImg, op)
    }
}

func DrawKillFeed(screen *ebiten.Image) {
    currentTime := time.Now()
    y := 50

    // Iterate over the kill feed entries
    for i, entry := range actors.TanksKilled {
        // Calculate the time difference
        timeDiff := currentTime.Sub(entry.TimeAdded)

        // Display the entry if it's less than 5 seconds old
        if timeDiff < 5*time.Second {

            isPlayer := strings.Contains(entry.TankName, "player")
            // Set text color based on whether it's a player or an enemy
            textColor := color.RGBA{255, 255, 255, 255} // Default to white
            if isPlayer {
                textColor = color.RGBA{0, 0, 255, 255} // Blue
            } else {
                textColor = color.RGBA{255, 0, 0, 255} // Red
            }

            // Calculate the width of the tank name
            bounds := text.BoundString(mplusNormalFont, entry.TankName)
            msgWidth := bounds.Max.X - bounds.Min.X

            // Draw the tank name with the specified color
            text.Draw(screen, entry.TankName, mplusNormalFont, 20, y, textColor)
            // Append the "destroyed" text in white after the tank name
            text.Draw(screen, " destroyed", mplusNormalFont, 20+msgWidth, y, color.White)

            // Move to the next line in the kill feed
            y += 30
        } else {
            // Remove entries older than 5 seconds
            actors.TanksKilled = append(actors.TanksKilled[:i], actors.TanksKilled[i+1:]...)
        }
    }
}

func init() {
    loadImages()

	tt, _ := opentype.Parse(fonts.MPlus1pRegular_ttf)
	mplusNormalFont, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    3,
		DPI:     10,
	})
}

func loadImages() {
    turretImage, _, err := ebitenutil.NewImageFromFile("resources/green_tank_turret.png")
	if err != nil {
		panic(err)
	}
    turretImg = turretImage

	projectileImage, _, err := ebitenutil.NewImageFromFile("resources/projectile.png")
	if err != nil {
		panic(err)
	}
	projectileImg = projectileImage

	explosionImage, _, err := ebitenutil.NewImageFromFile("resources/explosion.png")
	if err != nil {
		panic(err)
	}
	explosionImg = explosionImage
}

func drawDebugStuffActors(t *actors.Tank, screen *ebiten.Image) {
	// Draw player tank's collision box
	vector.StrokeLine(screen, float32(t.Hull.CollisionX1), 
		float32(t.Hull.CollisionY1),
		float32(t.Hull.CollisionX2),
		float32(t.Hull.CollisionY2), 1, color.RGBA{0, 65, 250, 255}, true)
	vector.StrokeLine(screen, float32(t.Hull.CollisionX2), 
		float32(t.Hull.CollisionY2),
		float32(t.Hull.CollisionX3),
		float32(t.Hull.CollisionY3), 1, color.RGBA{0, 65, 250, 255}, true)
	vector.StrokeLine(screen, float32(t.Hull.CollisionX3), 
		float32(t.Hull.CollisionY3),
		float32(t.Hull.CollisionX4),
		float32(t.Hull.CollisionY4), 1, color.RGBA{0, 65, 250, 255}, true)
	vector.StrokeLine(screen, float32(t.Hull.CollisionX4), 
		float32(t.Hull.CollisionY4),
		float32(t.Hull.CollisionX1),
		float32(t.Hull.CollisionY1), 1, color.RGBA{0, 65, 250, 255}, true)
}

