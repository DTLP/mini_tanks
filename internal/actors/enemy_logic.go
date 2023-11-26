package actors

import (
	"github.com/DTLP/mini_tanks/internal/levels"

	// "fmt"
    "math"
)

const (
	turretAngleTolerance = 1.0

	playerBaseX = 2500
	playerBaseY = 4850
	engagePlayerBaseDistance = 2000.0
)

var (
	angleToBase = 0.0
)

type Point struct {
    X, Y float64
}

func init() {
	//
}

func UpdateEnemyLogic(tanks *[]Tank, levelObjects []levels.LevelBlock) {
    var players []Tank

    // Separate players from non-players
    for i := range *tanks {
        t := &(*tanks)[i]
        if t.IsPlayer {
            players = append(players, *t)
        }
    }

    for i := range *tanks {
        t := &(*tanks)[i]
        if !t.IsPlayer {
            // Calculate angle to player's base
            angleToBase = calculateAngle(t, playerBaseX, playerBaseY)
            
            // Move towards player's base
            moveTank(t, angleToBase)

			// Shoot player base if tank is close to it
			if isCloseToPlayerBase(t, playerBaseX, playerBaseY) {
				angleToBase := calculateAngle(t, playerBaseX, playerBaseY)
				// Aim at player base
				aimAtTarget(t, angleToBase)
				if math.Abs(t.Turret.Angle-angleToBase) <= turretAngleTolerance &&
							t.Turret.ReloadTimer == 0 {
					// shoot(t)
				}
			}

            // Shoot player tanks if can see them
            if isPlayerInSight(t, players, levelObjects) {
                // Calculate angle to player
                angleToPlayer := calculateAngle(t, players[0].X, players[0].Y)
				// Aim at player
				aimAtTarget(t, angleToPlayer)
				if math.Abs(t.Turret.Angle-angleToPlayer) <= turretAngleTolerance &&
							t.Turret.ReloadTimer == 0 {
					// shoot(t)
				}
            }
        }
    }
}

func calculateAngle(t *Tank, targetX, targetY float64) float64 {
    // Calculate angle between tank and target
    deltaX := t.X - targetX
    deltaY := t.Y - targetY
    angle := math.Atan2(deltaY, deltaX) * 180 / math.Pi - 180

    return angle
}

func isCloseToPlayerBase(t *Tank, targetX, targetY float64) bool {
    // Calculate distance between tank and player base
	deltaX := targetX - t.X
    deltaY := targetY - t.Y
    distance := math.Sqrt(deltaX*deltaX + deltaY*deltaY)

    // Check if the distance is less than the threshold
    return distance < engagePlayerBaseDistance
}

func moveTank(tank *Tank, targetAngle float64) {
	// Adjust tank's angle towards the target angle
	angularDifference := targetAngle - tank.Hull.Angle

	// Normalize the angle difference to be between -180 and 180 degrees
	if angularDifference > 180 {
		angularDifference -= 360
	} else if angularDifference < -180 {
		angularDifference += 360
	}

	// Set the rotation direction based on the sign of the angle difference
	rotationDirection := 0.0
	if angularDifference > 0 {
		rotationDirection = 1.0
	} else if angularDifference < 0 {
		rotationDirection = -1.0
	}

	// Adjust the tank's angle based on the rotation speed
	rotationAmount := tank.Hull.RotationSpeed * rotationDirection
	tank.Hull.Angle += rotationAmount

	// Move the tank forward
	tank.PrevX = tank.X
	tank.PrevY = tank.Y
	tank.X += tank.Hull.Speed * math.Cos(-tank.Hull.Angle*math.Pi/180.0)
	tank.Y += tank.Hull.Speed * math.Sin(tank.Hull.Angle*math.Pi/180.0)

	// Update collision box
	updateCollisionBox(tank)
}


func isPlayerInSight(t *Tank, players []Tank, levelObjects []levels.LevelBlock) bool {
	for _, player := range players {
		if !isObstacleBetween(t, player, levelObjects) {
			return true
		}
	}

	// No player in sight
	return false
}

func isObstacleBetween(t *Tank, player Tank, levelObjects []levels.LevelBlock) bool {
    for _, levelObject := range levelObjects {
        for _, obstacle := range levelObject.Blocks {
            for _, line := range obstacle.Walls {
                if doLinesIntersect(t, player, line) {
                    return true
                }
            }
        }
    }

    // No obstacles between source and target
    return false
}

func doLinesIntersect(t *Tank, player Tank, line levels.Line) bool {
    // Define the four points of the two lines
    p1 := Point{X: line.X1, Y: line.Y1}
    q1 := Point{X: line.X2, Y: line.Y2}
    p2 := Point{X: t.X, Y: t.Y}
    q2 := Point{X: player.X, Y: player.Y}

    // Find the orientations for each triplet of points
    o1 := orientation(p1, q1, p2)
    o2 := orientation(p1, q1, q2)
    o3 := orientation(p2, q2, p1)
    o4 := orientation(p2, q2, q1)

    // Check for general case and special cases where the points are collinear
    if (o1 != o2) && (o3 != o4) {
        return true
    }

    // Check for special cases where the points are collinear and lie on the segments
    if (o1 == 0) && onSegment(p1, p2, q1) {
        return true
    }
    if (o2 == 0) && onSegment(p1, q2, q1) {
        return true
    }
    if (o3 == 0) && onSegment(p2, p1, q2) {
        return true
    }
    if (o4 == 0) && onSegment(p2, q1, q2) {
        return true
    }

    // No intersection
    return false
}

// orientation calculates the orientation of three points (p, q, r)
func orientation(p, q, r Point) int {
    val := (q.Y-p.Y)*(r.X-q.X) - (q.X-p.X)*(r.Y-q.Y)

    if val == 0 {
        return 0 // Collinear
    }
    if val > 0 {
        return 1 // Clockwise orientation
    }
    return 2 // Counterclockwise orientation
}

// onSegment checks if point q lies on segment pr
func onSegment(p, q, r Point) bool {
    return (q.X <= math.Max(p.X, r.X)) && (q.X >= math.Min(p.X, r.X)) &&
        (q.Y <= math.Max(p.Y, r.Y)) && (q.Y >= math.Min(p.Y, r.Y))
}

func aimAtTarget(t *Tank, angle float64) {
	if t.Turret.Angle > angle {
		t.Turret.Angle -= t.Turret.RotationSpeed
	}
	if t.Turret.Angle < angle {
		t.Turret.Angle += t.Turret.RotationSpeed
	}
}