package actors

import (
	// "fmt"
    "time"
    "math"
    "math/rand"
)

type KillFeedEntry struct {
    TankName string
    TimeAdded time.Time
}

var TanksKilled []KillFeedEntry

type ProgressBar struct {
    X, Y, Width, FilledWidth, Height int
}

var (
    progressBar ProgressBar
    progressBarFilled ProgressBar
    // Define the offset of the progress bar from the tank
    barOffsetX = 0
    barOffsetY = 10
    barWidth   = 300
    barHeight  = 5
)

type Tanks struct{
    Tanks   []Tank
}

type Tank struct {
    X               float64
    Y               float64
    PrevX           float64
    PrevY           float64
    MaxHealth       int
    Health          int
    HealthBarWidth  int
    HealthBarHeight int
	Name			string
    IsPlayer        bool
    Player          int
    CanMove         bool

    Hull            Hull
    Turret          Turret
    Projectiles     []Projectile
    Explosions      []Explosion

    ReloadBarWidth  int    
    ReloadBarHeight int

    LastCollisionTime time.Time
}

type Hull struct {
    PrevX           float64
    PrevY           float64
    Width           float64
    Height          float64
    Angle           float64
    Speed           float64
    ReverseSpeed    float64
    RotationSpeed   float64
    Image           string

    CollisionX1     float64
    CollisionY1     float64
    CollisionX2     float64
    CollisionY2     float64
    CollisionX3     float64
    CollisionY3     float64
    CollisionX4     float64
    CollisionY4     float64
}

type Turret struct {
    Width           float64
    Height          float64
    Length          float64
    Angle           float64
    RotationSpeed   float64
    Image           string

    ProjectileSpeed float64
    ReloadTime      float64
    ReloadTimer     float64

    ProgressBar     ProgressBar
}

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

func NewTank(name string) Tank {
    // Default preset
    tank := Tank{
        X:               1850,
        Y:               4730,
        MaxHealth:       200,
        Health:          200,
        HealthBarWidth:  50,
        HealthBarHeight: 5,
        ReloadBarWidth:  50,
        ReloadBarHeight: 5,
		Name:			 name,
		IsPlayer:        false,
        Player:          0,
        CanMove:         true,

		Hull: Hull{
            Width:           50,
            Height:          50,
            Angle:           -90.0,
            Speed:           20,
            ReverseSpeed:    10.0,
            RotationSpeed:   2.0,
            Image: "resources/green_tank_hull.png",
        },
        Turret: Turret{
            Width:           50,
            Height:          50,
            Angle:           -90.0,
            RotationSpeed:   2.0,
            ProjectileSpeed: 150.0,
			// ProjectileSpeed: 10.0,
            ReloadTime:      100.0,
			// ReloadTime:      1.0,
            ReloadTimer:     0.0,
            Image: "resources/green_tank_turret.png",
        },
    }

    switch name {
    case "player1":
		tank.IsPlayer = true
        tank.Player   = 1
		tank.X        = 1850 // Spawn next to the base
		tank.Y        = 4730
		tank.X   = tank.X
		tank.Y   = tank.Y
	case "player2":
		tank.IsPlayer = true
        tank.Player   = 2
		tank.X        = 2200 // Spawn next to the base
		tank.Hull.Image = "resources/brown_tank_hull.png"
		tank.Turret.Image = "resources/brown_tank_turret.png"
    default:
        rand.Seed(time.Now().UnixNano())
        spawnOptions := [][]int{{270, 270}, {2500, 270}, {4800, 270}}
        randomSpawn := spawnOptions[rand.Intn(len(spawnOptions))]

        tank.X = float64(randomSpawn[0])
        tank.Y = float64(randomSpawn[1])
        tank.MaxHealth            = 50
        tank.Health               = 50
        tank.Hull.Speed           = 5
        tank.Hull.Angle           = 90.0
        tank.Hull.Image = "resources/grey_tank_hull.png"
        tank.Turret.Angle         = 90.0
        tank.Turret.RotationSpeed = 1.0
        tank.Turret.ReloadTime    = 200.0
        tank.Turret.Image = "resources/grey_tank_turret.png"
        tank.Name = name
    }

	updateCollisionBox(&tank)

    return tank
}

func addProjectile(t *Tank) []Projectile {
    newProjectile := Projectile{
        X:         t.X,
        Y:         t.Y,
        VelocityX: t.Turret.ProjectileSpeed,
        VelocityY: t.Turret.ProjectileSpeed,
        Angle:     t.Turret.Angle,
        Width:     1.0,
        Height:    5.0,
        Collided:  false,
    }

    // Only one projectile per tank
    t.Projectiles = []Projectile{newProjectile}

    return t.Projectiles
}

func updateProjectiles(t *Tank) {
    // Update the position of projectiles
    for i := range t.Projectiles {
        if t.Projectiles[i].Collided {
            // Create an explion at impact coordinates and remove projectile
            addExplosion(t.Projectiles[i].X, t.Projectiles[i].Y, t)
            t.Projectiles = removeProjectile(t.Projectiles, i)

            continue
        }

        angleRadians := t.Projectiles[i].Angle * math.Pi / 180.0 // Convert degrees to radians
        t.Projectiles[i].X += t.Projectiles[i].VelocityX * math.Cos(angleRadians)
        t.Projectiles[i].Y += t.Projectiles[i].VelocityY * math.Sin(angleRadians)
    }
}

func removeProjectile(projectiles []Projectile, index int) []Projectile {
    // Ensure the index is within bounds
    if index < 0 || index >= len(projectiles) {
        return projectiles
    }

    // Create a new slice without the projectile at the specified index
    return append(projectiles[:index], projectiles[index+1:]...)

}

func updateExplosions(t *Tank) {
    for i := range t.Explosions {
        // Increment the Frame value or perform any other updates
        t.Explosions[i].Frame++

        // Check if the explosion has reached its maximum frame count
        if t.Explosions[i].Frame >= 16 {
            // Remove the explosion from the slice
            t.Explosions = append(t.Explosions[:i], t.Explosions[i+1:]...)
        }
    }
}

func addExplosion(x, y float64, t *Tank) []Explosion {
    newExplosion := Explosion{
        X:        x,
        Y:        y,
        Frame:    -1,
    }
    t.Explosions = append(t.Explosions, newExplosion)

    return t.Explosions

}

func Update(tanks *[]Tank) {
    for i := range *tanks {
        t := &(*tanks)[i]

        updateGunReloadTimer(t)

        updateProjectiles(t)

        updateExplosions(t)
    }
}

func updateGunReloadTimer(t *Tank) {
    // Update turret reload progress bar
    if t.Turret.ReloadTimer > 0 {
        t.Turret.ReloadTimer--

        // Calculate the percentage of reload progress
        progressPercentage := float64(t.Turret.ReloadTimer) / float64(t.Turret.ReloadTime)

        // Calculate the width of the filled portion
        filledWidth := int(float64(barWidth) * progressPercentage)

        // Update the progress bar's position and size
        progressBar.X = int(t.X - float64(barWidth)/2)
        progressBar.Y = int(t.Y + t.Hull.Height/2 + float64(barOffsetY))
        progressBar.Width = barWidth        
        progressBar.Height = barHeight
        progressBar.FilledWidth = filledWidth
    }
}

func GetUpdatedTankList(tanks []Tank) []Tank {
    var aliveTanks []Tank

    for _, tank := range tanks {
        if tankIsAlive(tank) {
            aliveTanks = append(aliveTanks, tank)
        }
    }

    return aliveTanks
}

func tankIsAlive(tank Tank) bool {
    if tank.Health == 0 {
        // Append tank to the Kill Feed
        entry := KillFeedEntry{
            TankName: tank.Name,
            TimeAdded: time.Now(),
        }
        TanksKilled = append(TanksKilled, entry)

        // Increase killed enemy count
        if !tank.IsPlayer {
            killedEnemies += 1
        }

        return false
    }

    return true
}


func shoot(t *Tank) {
    t.Projectiles = addProjectile(t)

    // Start the reload timer
    t.Turret.ReloadTimer = t.Turret.ReloadTime

    // Create an instance of ProgressBar and set its values
    progressBar := ProgressBar{
        X:           int(t.X / gameLogicToScreenXOffset),
        Y:           int(t.Y / gameLogicToScreenYOffset) + 50,
        Width:       50,
        FilledWidth: 0,
        Height:      5,
    }
    t.Turret.ProgressBar = progressBar
}



type TankUpdate struct {
    Type        string   `json:"Type"`
    X        	float64  `json:"X"`
    Y		  	float64  `json:"Y"`
    Name        string   `json:"Name"`
    HullAngle		float64  `json:"HullAngle"`
    HullCollisionX1 float64  `json:"X1"`
    HullCollisionY1 float64  `json:"Y1"`
    HullCollisionX2 float64  `json:"X2"`
    HullCollisionY2 float64  `json:"Y2"`
    HullCollisionX3 float64  `json:"X3"`
    HullCollisionY3 float64  `json:"Y3"`
    HullCollisionX4 float64  `json:"X4"`
    HullCollisionY4 float64  `json:"Y4"`
    TurretAngle     float64  `json:"TurretAngle"`
    Projectiles     []Projectile `json:"Projectiles"`
    Explosions      []Explosion  `json:"Explosions"`
}
