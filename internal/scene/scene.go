package scene

import (
	"github.com/DTLP/mini_tanks/internal/actors"
	"github.com/DTLP/mini_tanks/internal/levels"
       
    // "os"
    // "io"
	"fmt"
	// "bytes"
	"image"
	"image/color"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
)

const (
    screenWidth  			 = 1000
    screenHeight 			 = 1000
    minXCoordinates 		 = 270
    minYCoordinates 		 = 270
    gameLogicToScreenXOffset = 5.12
    gameLogicToScreenYOffset = 5.12
    padding      			 = 20
)

type GameState int

const (
    GameStateMainMenu GameState = iota
    GameStateGameplay
    GameStateSettings
)

var (
	logoImg, bgImage *ebiten.Image
	playButtonImg, exitButtonImg, soloButtonImg, coopButtonImg *ebiten.Image
	hostButtonImg, joinButtonImg, backButtonImg *ebiten.Image
	shadowImage   = ebiten.NewImage(screenWidth, screenHeight)
	triangleImage = ebiten.NewImage(screenWidth, screenHeight)
)

var (
	mplusNormalFont, mplusSmallFont  font.Face
)

func DrawScreen(levelNum int, screen *ebiten.Image) {
    screen.Fill(color.RGBA{240, 222, 180, 255}) // Desert background
	// Reset the shadowImage
	shadowImage.Fill(color.RGBA{50, 50, 50, 255}) // Grey shadowns
	// Draw level number
	if levelNum > 0 {
		levelText := fmt.Sprintf("Level: %v", levelNum)
		text.Draw(screen, levelText, mplusNormalFont, 20, 980, color.White)
	}
}

func DrawLevel(levelObjects []levels.LevelBlock, tankX, tankY float64, screen *ebiten.Image) {

	drawShadows(levelObjects, tankX, tankY, screen)

	drawLevelObjects(levelObjects, screen)

	// drawDebugStuffLevels(levelObjects, screen)
}

func drawShadows(levelObjects []levels.LevelBlock, tankX, tankY float64, screen *ebiten.Image) {
	// Create a new slice to match the structure of 'levels []Block'
	var blocks []levels.Block

	for _, object := range levelObjects {
		if !object.Border && object.Health == 0 {
			// Skip destroyed objects
			continue
		}

		for _, block := range object.Blocks {
			// Append 'block' to 'objects'
			blocks = append(blocks, block)
		}
	}

	rays := rayCasting(
		float64(tankX / gameLogicToScreenXOffset),
		float64(tankY / gameLogicToScreenYOffset),
		blocks,
	)

	// Subtract ray triangles from shadow
	opt := &ebiten.DrawTrianglesOptions{}
	opt.Address = ebiten.AddressRepeat
	opt.Blend = ebiten.BlendSourceOut
	for i, line := range rays {
		nextLine := rays[(i+1)%len(rays)]

		// Draw triangle of area between rays
		v := rayVertices(float64(tankX / gameLogicToScreenXOffset), float64(tankY / gameLogicToScreenYOffset), nextLine.X2, nextLine.Y2, line.X2, line.Y2)
		shadowImage.DrawTriangles(v, []uint16{0, 1, 2}, triangleImage, opt)
	}

	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(1)
	screen.DrawImage(shadowImage, op)

	// Debug stuff
	// ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()), padding, 150)
	// ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rays: 2*%d", len(rays)/2), padding, 222)
	// // Draw rays
	// for _, r := range rays {
	// 	vector.StrokeLine(screen, float32(r.X1), float32(r.Y1), float32(r.X2), float32(r.Y2), 1, color.RGBA{255, 255, 0, 150}, true)
	// }
}

func drawLevelObjects(levelObjects []levels.LevelBlock, screen *ebiten.Image) {
	for _, object := range levelObjects {
        // Skip destroyed blocks
        if !object.Border && object.Health == 0 {
            continue
        }

		originalImg, _, _ := ebitenutil.NewImageFromFile(object.Image.Path)

		// Create a new image representing a sub-image of the original image
		subImg := originalImg.SubImage(image.Rect(object.Image.X, object.Image.Y,
					object.Image.Width, object.Image.Height)).(*ebiten.Image)
		// Draw the sub-image on the screen
		options := &ebiten.DrawImageOptions{}
		options.GeoM.Translate(object.X, object.Y)
		screen.DrawImage(subImg, options)
    }
}

// func drawDebugStuffLevels(levelObjects []levels.LevelBlock, screen *ebiten.Image) {
	// // Draw walls - raycasting red lines
	// for _, obj := range levelObjects {
	// 	for _, block := range obj.Blocks {
	// 		for _, w := range block.Walls {
	// 			vector.StrokeLine(screen, float32(w.X1), float32(w.Y1), float32(w.X2), float32(w.Y2), 1, color.RGBA{255, 0, 0, 255}, true)
	// 		}
	// 	}
	// }
// }

func init() {
	tt, _ := opentype.Parse(fonts.MPlus1pRegular_ttf)
	mplusNormalFont, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     100,
	})
	mplusSmallFont, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    12,
		DPI:     100,
	})

	// Load images
	logoImage, _, _ := ebitenutil.NewImageFromFile("resources/logo.png")
	logoImg = logoImage
	playButtonImage, _, _ := ebitenutil.NewImageFromFile("resources/play_button.png")
	playButtonImg = playButtonImage
	soloButtonImage, _, _ := ebitenutil.NewImageFromFile("resources/solo_button.png")
	soloButtonImg = soloButtonImage
	coopButtonImage, _, _ := ebitenutil.NewImageFromFile("resources/coop_button.png")
	coopButtonImg = coopButtonImage
	hostButtonImage, _, _ := ebitenutil.NewImageFromFile("resources/host_button.png")
	hostButtonImg = hostButtonImage
	joinButtonImage, _, _ := ebitenutil.NewImageFromFile("resources/join_button.png")
	joinButtonImg = joinButtonImage
	backButtonImage, _, _ := ebitenutil.NewImageFromFile("resources/back_button.png")
	backButtonImg = backButtonImage
	exitButtonImage, _, _ := ebitenutil.NewImageFromFile("resources/exit_game_button.png")
	exitButtonImg = exitButtonImage
}

func DrawMainMenu(tanks []actors.Tank, menuStage string, screen *ebiten.Image) {

	// Draw Game logo
	logoOp := &ebiten.DrawImageOptions{}
	logoOp.GeoM.Scale(.25, .25)
	logoOp.GeoM.Translate(40, 50)
	screen.DrawImage(logoImg, logoOp)

	drawKeyboard(screen)

	drawMenuButtons(tanks, menuStage, screen)

	credsText := "github.com/DTLP"
	text.Draw(screen, credsText, mplusNormalFont, 700, 50, color.Black)
}

func drawKeyboard(screen *ebiten.Image) {
	keySize := 25.0
	keyOutline := 2.0
	colorBg := color.RGBA{240, 222, 180, 255}
	// colorPink := color.RGBA{255, 180, 180, 255}
	colorPressed := color.RGBA{240, 205, 130, 255}


	// Key W
	ebitenutil.DrawRect(screen, 80, 350, keySize, keySize, color.Black)
	ebitenutil.DrawRect(screen, 81, 351, keySize-keyOutline, keySize-keyOutline, colorBg)
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		ebitenutil.DrawRect(screen, 81, 351, keySize-keyOutline, keySize-keyOutline, colorPressed)
	}
	keyWText := "W"
	text.Draw(screen, keyWText, mplusSmallFont, 84, 369, color.Black)

	// Key A
	ebitenutil.DrawRect(screen, 60, 377, keySize, keySize, color.Black)
	ebitenutil.DrawRect(screen, 61, 378, keySize-keyOutline, keySize-keyOutline, colorBg)
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		ebitenutil.DrawRect(screen, 61, 378, keySize-keyOutline, keySize-keyOutline, colorPressed)
	}
	keyAText := "A"
	text.Draw(screen, keyAText, mplusSmallFont, 66, 396, color.Black)

	// Key S
	ebitenutil.DrawRect(screen, 87, 377, keySize, keySize, color.Black)
	ebitenutil.DrawRect(screen, 88, 378, keySize-keyOutline, keySize-keyOutline, colorBg)
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		ebitenutil.DrawRect(screen, 88, 378, keySize-keyOutline, keySize-keyOutline, colorPressed)
	}
	keySText := "S"
	text.Draw(screen, keySText, mplusSmallFont, 93, 396, color.Black)

	// Key D
	ebitenutil.DrawRect(screen, 114, 377, keySize, keySize, color.Black)
	ebitenutil.DrawRect(screen, 115, 378, keySize-keyOutline, keySize-keyOutline, colorBg)
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		ebitenutil.DrawRect(screen, 115, 378, keySize-keyOutline, keySize-keyOutline, colorPressed)
	}
	keyDText := "D"
	text.Draw(screen, keyDText, mplusSmallFont, 121, 396, color.Black)

	// Key Space
	ebitenutil.DrawRect(screen, 168, 377, keySize*5, keySize, color.Black)
	ebitenutil.DrawRect(screen, 169, 378, keySize*5-keyOutline, keySize-keyOutline, colorBg)
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		ebitenutil.DrawRect(screen, 169, 378, keySize*5-keyOutline, keySize-keyOutline, colorPressed)
	}

	// Key Up Arrow
	ebitenutil.DrawRect(screen, 347, 350, keySize, keySize, color.Black)
	ebitenutil.DrawRect(screen, 348, 351, keySize-keyOutline, keySize-keyOutline, colorBg)

	// Key Left Arrow
	ebitenutil.DrawRect(screen, 320, 377, keySize, keySize, color.Black)
	ebitenutil.DrawRect(screen, 321, 378, keySize-keyOutline, keySize-keyOutline, colorBg)
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		ebitenutil.DrawRect(screen, 321, 378, keySize-keyOutline, keySize-keyOutline, colorPressed)
	}
	keyLeftText := "<"
	text.Draw(screen, keyLeftText, mplusSmallFont, 326, 396, color.Black)

	// Key Down Arrow
	ebitenutil.DrawRect(screen, 347, 377, keySize, keySize, color.Black)
	ebitenutil.DrawRect(screen, 348, 378, keySize-keyOutline, keySize-keyOutline, colorBg)

	// Key Right Arrow
	ebitenutil.DrawRect(screen, 374, 377, keySize, keySize, color.Black)
	ebitenutil.DrawRect(screen, 375, 378, keySize-keyOutline, keySize-keyOutline, colorBg)
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		ebitenutil.DrawRect(screen, 375, 378, keySize-keyOutline, keySize-keyOutline, colorPressed)
	}
	keyRightText := ">"
	text.Draw(screen, keyRightText, mplusSmallFont, 380, 396, color.Black)

	keyControlsText := "Move                   Shoot                  Aim"
	text.Draw(screen, keyControlsText, mplusSmallFont, 75, 420, color.Black)

}

func drawMenuButtons(tanks []actors.Tank, menuStage string, screen *ebiten.Image) {
	// Play
	if menuStage == "init" {
		buttonPlayOp := &ebiten.DrawImageOptions{}
		buttonPlayOp.GeoM.Translate(700, 450)
		screen.DrawImage(playButtonImg, buttonPlayOp)
	}
	if menuStage == "play" {
		// Solo
		buttonSoloOp := &ebiten.DrawImageOptions{}
		buttonSoloOp.GeoM.Translate(700, 450)
		screen.DrawImage(soloButtonImg, buttonSoloOp)
		// Co-op
		buttonCoopOp := &ebiten.DrawImageOptions{}
		buttonCoopOp.GeoM.Translate(700, 550)
		// alpha := float64(128) / 255.0
		// buttonCoopOp.ColorM.Scale(1, 1, 1, alpha)
		screen.DrawImage(coopButtonImg, buttonCoopOp)
		// Exit
	}
	if menuStage == "coop" {
		// Host
		buttonHostOp := &ebiten.DrawImageOptions{}
		buttonHostOp.GeoM.Translate(700, 450)
		screen.DrawImage(hostButtonImg, buttonHostOp)
		// Join
		buttonJoinOp := &ebiten.DrawImageOptions{}
		buttonJoinOp.GeoM.Translate(700, 550)
		screen.DrawImage(joinButtonImg, buttonJoinOp)
		// Back
		buttonBackOp := &ebiten.DrawImageOptions{}
		buttonBackOp.GeoM.Translate(700, 650)
		screen.DrawImage(backButtonImg, buttonBackOp)
	}
	if menuStage == "host" {
		if actors.CountPlayerTanks(tanks) == 2 {
			// Play button
			buttonPlayOp := &ebiten.DrawImageOptions{}
			buttonPlayOp.GeoM.Translate(700, 450)
			screen.DrawImage(playButtonImg, buttonPlayOp)
		}
		// Back
		buttonBackOp := &ebiten.DrawImageOptions{}
		buttonBackOp.GeoM.Translate(700, 650)
		screen.DrawImage(backButtonImg, buttonBackOp)
	}
	// Exit
	buttonExitOp := &ebiten.DrawImageOptions{}
	buttonExitOp.GeoM.Translate(700, 875)
	screen.DrawImage(exitButtonImg, buttonExitOp)
}

func DrawGameOverScreen(screen *ebiten.Image) {
	// Fill the background with black color
	screen.Fill(color.Black)

	// Show Game Over message
	gameOverText := "Game Over"
	textWidth := text.BoundString(mplusNormalFont, gameOverText).Max.X
	textHeight := text.BoundString(mplusNormalFont, "A").Max.Y
	textX := (screenWidth - textWidth) / 2
	textY := (screenHeight - textHeight) / 2
	text.Draw(screen, gameOverText, mplusNormalFont, textX, textY, color.White)

	gameOverText = "Press Enter to try again"
	textWidth = text.BoundString(mplusNormalFont, gameOverText).Max.X
	textHeight = text.BoundString(mplusNormalFont, "A").Max.Y
	textX = (screenWidth - textWidth) / 2
	textY = (screenHeight - textHeight) / 2
	text.Draw(screen, gameOverText, mplusNormalFont, textX, textY + 100, color.White)
}
