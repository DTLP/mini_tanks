package main

import (
    "github.com/DTLP/mini_tanks/internal/game"

    "github.com/hajimehoshi/ebiten/v2"	
)


func main() {
    ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
    ebiten.SetWindowTitle("Mini Tanks")

    game := game.NewGame()
    if err := ebiten.RunGame(game); err != nil {
        panic(err)
    }
}
