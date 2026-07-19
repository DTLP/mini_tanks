package game

import (
	"github.com/DTLP/mini_tanks/internal/levels"
	"github.com/DTLP/mini_tanks/internal/actors"
	"github.com/DTLP/mini_tanks/internal/scene"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth   = 1000
	ScreenHeight  = 1000
	minXCoordinates = 270
	minYCoordinates = 270
	maxXCoordinates = 4730
	maxYCoordinates = 4730
	gameLogicToScreenXOffset = 5.12
	gameLogicToScreenYOffset = 5.12
	padding      = 20
)

// Roles describe who runs the simulation for this process.
const (
	RoleSolo = iota
	RoleHost
	RoleClient
)

var (
	// role selects the simulation model for this process.
	role         = RoleSolo
	netConn      *NetConn
	lastP2Input  *actors.Input

	levelNum     = 0
	levelObjects = []levels.LevelBlock{}

	gameOver = false

	// statusMessage / statusVisible surface netplay status text. The scene
	// overlay rendering is a later step; for now they are just exposed.
	statusMessage string
	statusVisible bool
)

type Game struct {
	Tanks 		  []actors.Tank
	levelObjects  []levels.LevelBlock
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

	switch role {
	case RoleSolo:
		updateSolo(g)
	case RoleHost:
		updateHost(g)
	case RoleClient:
		updateClient(g)
	}

	return nil
}

// updateSolo runs the unchanged single-player simulation.
func updateSolo(g *Game) {
	for i := range g.Tanks {
		if g.Tanks[i].IsPlayer {
			actors.HandleMovement(&g.Tanks[i])
		}
	}

	actors.HandleCollision(&g.Tanks, levelObjects)
	actors.Update(&g.Tanks)
	actors.UpdateEnemyLogic(&g.Tanks, levelObjects)

	g.Tanks = actors.GetUpdatedTankList(g.Tanks)
	g.Tanks = actors.CheckEnemyCount(g.Tanks)

	if actors.NoEnemiesLeft(g.Tanks) && levelNum != 0 {
		levelNum += 1
		actors.ResetPlayerPositions(&g.Tanks)
		actors.ResetEnemyNamePool()
		levelObjects = levels.GetLevelObjects(levelNum)
	}

	levels.UpdateLevelObjects(levelObjects)

	if actors.NoPlayersLeft(g.Tanks) || levels.IsBaseDestroyed(levelObjects) {
		gameOver = true
	}

	if gameOver && ebiten.IsKeyPressed(ebiten.KeyEnter) {
		g.Tanks, g.levelObjects = restartGame()
	}
}

// updateHost runs the full simulation as the authoritative server. It reads
// player1 from the local keyboard, applies player2's latest remote input, then
// broadcasts a full world snapshot every frame.
func updateHost(g *Game) {
	if levelNum == 0 {
		return // still in lobby; the menu handles host_wait.
	}

	if netConn == nil || netConnClosed() {
		showDisconnected()
		maybeFallBackToSolo(g)
		return
	}

	// Player1 local input.
	for i := range g.Tanks {
		if g.Tanks[i].IsPlayer && g.Tanks[i].Player == 1 {
			actors.HandleMovement(&g.Tanks[i])
		}
	}

	// Player2 remote input (sticky between packets).
	if in := netConn.GetInput(); in != nil {
		lastP2Input = in
	}
	if lastP2Input != nil {
		for i := range g.Tanks {
			if g.Tanks[i].IsPlayer && g.Tanks[i].Player == 2 {
				actors.ApplyInput(&g.Tanks[i], lastP2Input)
			}
		}
	}

	// Host-only full simulation.
	actors.HandleCollision(&g.Tanks, levelObjects)
	actors.Update(&g.Tanks)
	actors.UpdateEnemyLogic(&g.Tanks, levelObjects)

	g.Tanks = actors.GetUpdatedTankList(g.Tanks)
	g.Tanks = actors.CheckEnemyCount(g.Tanks)

	if actors.NoEnemiesLeft(g.Tanks) && levelNum != 0 {
		levelNum += 1
		// Respawn both players on level advance so the next level starts
		// with a full team regardless of who died this level.
		if !doesPlayer1Exist(g.Tanks) {
			g.Tanks = append(g.Tanks, actors.NewTank("player1"))
		}
		if !doesPlayer2Exist(g.Tanks) {
			g.Tanks = append(g.Tanks, actors.NewTank("player2"))
		}
		actors.ResetPlayerPositions(&g.Tanks)
		actors.ResetEnemyNamePool()
		levelObjects = levels.GetLevelObjects(levelNum)
	}

	levels.UpdateLevelObjects(levelObjects)

	if actors.NoPlayersLeft(g.Tanks) || levels.IsBaseDestroyed(levelObjects) {
		gameOver = true
	}

	if gameOver && ebiten.IsKeyPressed(ebiten.KeyEnter) {
		g.Tanks, g.levelObjects = restartGame()
	}

	// Broadcast the authoritative snapshot.
	netConn.SendSnapshot(buildSnapshot(g.Tanks, levelObjects, levelNum, gameOver))
}

// updateClient performs no simulation. It sends local input to the host and
// renders the latest received snapshot.
func updateClient(g *Game) {
	if netConn == nil || netConnClosed() {
		// While at the lobby (levelNum == 0) the menu's join_connect stage
		// owns connection errors; only surface disconnects mid-game.
		if levelNum != 0 {
			showDisconnected()
			maybeFallBackToSolo(g)
		}
		return
	}

	// Send local input every frame.
	in := actors.ReadLocalInput()
	netConn.SendInput(in)

	// Pull the latest snapshot.
	if snap := netConn.GetSnapshot(); snap != nil {
		g.Tanks = reconstructTanks(snap.Tanks)
		levelObjects = reconstructBlocks(snap.Blocks)
		levelNum = snap.LevelNum
		gameOver = snap.GameOver
		// Adopt the host's kill feed: the client runs no simulation, so it
		// never calls GetUpdatedTankList and would otherwise see no entries.
		actors.TanksKilled = snap.KillFeed
	}
}

// netConnClosed reports whether the underlying connection has terminated.
func netConnClosed() bool {
	select {
	case <-netConn.Done():
		return true
	default:
		return false
	}
}

// showDisconnected surfaces a disconnect status message for a later scene
// overlay. The connection is left to be torn down via fallbackToSolo.
func showDisconnected() {
	statusMessage = "Disconnected"
	statusVisible = true
}

// maybeFallBackToSolo returns to the menu when the player presses Escape after
// a disconnect.
func maybeFallBackToSolo(g *Game) {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		g.fallbackToSolo()
	}
}

// fallbackToSolo tears down any netplay state and rebuilds the solo world.
func (g *Game) fallbackToSolo() {
	if netConn != nil {
		netConn.Close()
		netConn = nil
	}
	lastP2Input = nil
	role = RoleSolo
	menuStage = "init"
	levelNum = 0
	gameOver = false
	statusMessage = ""
	statusVisible = false
	actors.ResetCounter()
	actors.ResetEnemyNamePool()
	actors.FriendlyFire = false

	var tanks []actors.Tank
	tanks = append(tanks, actors.NewTank("player1"))
	g.Tanks = tanks
	levelObjects = levels.GetLevelObjects(levelNum)
	g.levelObjects = levelObjects
}

// localPlayerID returns the Player id this process drives: 2 for a netplay
// client, 1 for solo play and the host.
func localPlayerID() int {
	if role == RoleClient {
		return 2
	}
	return 1
}

// doesPlayer1Exist reports whether a "player1" tank is present. The host uses
// this to re-add player1 when the level advances.
func doesPlayer1Exist(tanks []actors.Tank) bool {
	for _, tank := range tanks {
		if tank.Name == "player1" {
			return true
		}
	}
	return false
}

// doesPlayer2Exist reports whether a "player2" tank is present. The host uses
// this to re-add player2 after a death mid-level.
func doesPlayer2Exist(tanks []actors.Tank) bool {
	for _, tank := range tanks {
		if tank.Name == "player2" {
			return true
		}
	}
	return false
}

func (g *Game) Draw(screen *ebiten.Image) {
	scene.DrawScreen(levelNum, screen)

	// Draw Main Menu
	if levelNum == 0 {
		scene.DrawMainMenu(g.Tanks, menuStage, scene.MenuState{JoinAddr: joinAddr, StatusMessage: statusMessage}, screen)
	}

	// Draw actors
	for i := range g.Tanks {
		scene.DrawActors(&g.Tanks[i], screen)
	}

	// Draw level objects from a viewpoint. Prefer the local player's tank,
	// then fall back to the other player's tank when the local one is dead,
	// and finally to the player base so the level stays visible even with no
	// players left. This keeps the arena drawn behind the remaining tanks.
	viewX, viewY := 2500.0, 4850.0 // player base fallback
	found := false
	for _, t := range g.Tanks {
		if t.IsPlayer && t.Player == localPlayerID() {
			viewX, viewY = t.X, t.Y
			found = true
			break
		}
	}
	if !found {
		for _, t := range g.Tanks {
			if t.IsPlayer {
				viewX, viewY = t.X, t.Y
				found = true
				break
			}
		}
	}
	scene.DrawLevel(levelObjects, viewX, viewY, screen)

	scene.DrawKillFeed(screen)

	// Check Game Over conditions
	if gameOver {
		scene.DrawGameOverScreen(screen)
	}

	// Draw the disconnect status overlay on top of the frozen world.
	// Only render mid-game: lobby connection errors are shown as red hint
	// text by the join_input/join_connect menu stages instead.
	if statusVisible && !gameOver && levelNum != 0 {
		scene.DrawStatusOverlay(screen, "Disconnected", "Press Esc to return to menu")
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {

	return ScreenWidth, ScreenHeight
}

// restartGame resets the world to level 1 so the players can try again. In
// coop (host) both players are restored; solo keeps just player1.
func restartGame() ([]actors.Tank, []levels.LevelBlock) {
	levelNum = 1
	gameOver = false
	actors.ResetCounter()
	actors.ResetEnemyNamePool()
	actors.MaxEnemies = 3

	var tanks []actors.Tank
	tanks = append(tanks, actors.NewTank("player1"))
	if role == RoleHost {
		tanks = append(tanks, actors.NewTank("player2"))
	}
	actors.ResetPlayerPositions(&tanks)
	levelObjects = levels.GetLevelObjects(levelNum)

	return tanks, levelObjects
}

// buildSnapshot projects the authoritative world state for the client.
func buildSnapshot(tanks []actors.Tank, levelObjects []levels.LevelBlock, ln int, goFlag bool) Snapshot {
	s := Snapshot{LevelNum: ln, GameOver: goFlag}
	for _, t := range tanks {
		s.Tanks = append(s.Tanks, TankSnap{
			Name: t.Name, Player: t.Player, IsPlayer: t.IsPlayer,
			X: t.X, Y: t.Y, HullAngle: t.Hull.Angle, TurretAngle: t.Turret.Angle,
			HullImage: t.Hull.Image, TurretImage: t.Turret.Image,
			Width: t.Hull.Width, Height: t.Hull.Height,
			TurretWidth: t.Turret.Width, TurretHeight: t.Turret.Height, TurretLength: t.Turret.Length,
			Health: t.Health, MaxHealth: t.MaxHealth,
			HealthBarWidth: t.HealthBarWidth, HealthBarHeight: t.HealthBarHeight,
			ReloadBarWidth: t.ReloadBarWidth, ReloadBarHeight: t.ReloadBarHeight,
			ReloadTimer: t.Turret.ReloadTimer, ReloadTime: t.Turret.ReloadTime,
			Projectiles: t.Projectiles, Explosions: t.Explosions,
			LastDamagedBy: t.LastDamagedBy,
		})
	}
	for _, b := range levelObjects {
		s.Blocks = append(s.Blocks, BlockSnap{
			X: b.X, Y: b.Y, Width: b.Width, Height: b.Height,
			Health: b.Health, Base: b.Base, Border: b.Border,
			Destructible: b.Destructible, Collidable: b.Collidable,
			ImagePath: b.Image.Path, ImageX: b.Image.X, ImageY: b.Image.Y,
			ImageWidth: b.Image.Width, ImageHeight: b.Image.Height,
			Blocks: b.Blocks,
		})
	}
	// Send the authoritative kill feed so the client renders the same entries
	// (the client runs no simulation and never calls GetUpdatedTankList).
	s.KillFeed = make([]actors.KillFeedEntry, len(actors.TanksKilled))
	copy(s.KillFeed, actors.TanksKilled)
	return s
}

// reconstructTanks rebuilds client-side tanks from a snapshot. The client only
// renders, so speeds are irrelevant; collision boxes are recomputed so the
// renderer (and any read-only checks) behave.
func reconstructTanks(snaps []TankSnap) []actors.Tank {
	var tanks []actors.Tank
	for _, s := range snaps {
		t := actors.Tank{
			X: s.X, Y: s.Y, Name: s.Name, Player: s.Player, IsPlayer: s.IsPlayer,
			Health: s.Health, MaxHealth: s.MaxHealth,
			HealthBarWidth: s.HealthBarWidth, HealthBarHeight: s.HealthBarHeight,
			ReloadBarWidth: s.ReloadBarWidth, ReloadBarHeight: s.ReloadBarHeight,
			CanMove: true,
			Hull: actors.Hull{
				Image: s.HullImage, Width: s.Width, Height: s.Height, Angle: s.HullAngle,
			},
			Turret: actors.Turret{
				Image: s.TurretImage, Width: s.TurretWidth, Height: s.TurretHeight,
				Length: s.TurretLength, Angle: s.TurretAngle,
				ReloadTimer: s.ReloadTimer, ReloadTime: s.ReloadTime,
			},
			Projectiles: s.Projectiles, Explosions: s.Explosions,
			LastDamagedBy: s.LastDamagedBy,
		}
		actors.UpdateCollisionBoxExternal(&t)
		tanks = append(tanks, t)
	}
	return tanks
}

// reconstructBlocks rebuilds client-side level objects from a snapshot.
func reconstructBlocks(snaps []BlockSnap) []levels.LevelBlock {
	var blocks []levels.LevelBlock
	for _, s := range snaps {
		blocks = append(blocks, levels.LevelBlock{
			X: s.X, Y: s.Y, Width: s.Width, Height: s.Height,
			Health: s.Health, Base: s.Base, Border: s.Border,
			Destructible: s.Destructible, Collidable: s.Collidable,
			Image: levels.Image{
				Path: s.ImagePath, X: s.ImageX, Y: s.ImageY,
				Width: s.ImageWidth, Height: s.ImageHeight,
			},
			Blocks: s.Blocks,
		})
	}
	return blocks
}
