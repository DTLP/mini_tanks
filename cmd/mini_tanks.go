package main

import (
    "github.com/DTLP/mini_tanks/internal/game"

    "github.com/hajimehoshi/ebiten/v2"	
)


func main() {
    ebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
    ebiten.SetWindowTitle("Mini Tanks")

    // Keep Update running even while the window is unfocused so menu buttons
    // (e.g. the host_wait Back/Cancel button) still respond to mouse clicks.
    // This is the default on desktop, but set it explicitly for safety.
    ebiten.SetRunnableOnUnfocused(true)

    game := game.NewGame()
    if err := ebiten.RunGame(game); err != nil {
        panic(err)
    }
}
