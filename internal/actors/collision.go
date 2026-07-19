package actors

import (
	"github.com/DTLP/mini_tanks/internal/levels"

	// "fmt"
    "time"
    "math"
    "math/rand"
)

const (
	epsilon = 1e-9
)

type Vector struct {
    X float64
    Y float64
}

func init() {

    rand.Seed(time.Now().UnixNano())
}

func HandleCollision(tanks *[]Tank, levelObjects []levels.LevelBlock) {
    // Tank - Level object collision check
    for ti, _ := range *tanks {
        if hasActorCollided(&(*tanks)[ti], levelObjects) {
            // Don't let tanks get stuck when they collide with level objects
            moveActorToPreviousPosition(&(*tanks)[ti])

            // Turn enemy tanks randomly to the left or right if they collide
            // with level objects. That way they can keep moving towards the base
            if !(*tanks)[ti].IsPlayer {
                // Check if enough time has passed since the last collision
                if time.Since((*tanks)[ti].LastCollisionTime) > time.Second {
                    // Randomly turn left or right
                    if rand.Intn(2) == 0 {
                        (*tanks)[ti].Hull.Angle += 90.0
                    } else {
                        (*tanks)[ti].Hull.Angle -= 90.0
                    }

                    // Update the last collision time
                    (*tanks)[ti].LastCollisionTime = time.Now()
                }
            }
        }
    }

    // Projectile collision check
    checkProjectileCollisions(tanks, levelObjects)
}

func hasActorCollided(tank *Tank, levelObjects []levels.LevelBlock) bool {
    tankVectors := getTankCollisionVectors(tank)

    for _, object := range levelObjects {
        // Skip destroyed blocks and objects not designed to be collidable with
        if !object.Border && object.Collidable && object.Health > 0 {
            objectVectors := getObjectCollisionVectors(object)

            // Check for intersections between tank and object vectors
            if vectorsIntersect(tankVectors, objectVectors) {
                return true
            }
        }
    }

    // No collision detected
    return false
}

func getTankCollisionVectors(tank *Tank) []Vector {
    // Define tank's collision points as vectors
    vectors := []Vector{
        {tank.Hull.CollisionX1, tank.Hull.CollisionY1},
        {tank.Hull.CollisionX2, tank.Hull.CollisionY2},
        {tank.Hull.CollisionX3, tank.Hull.CollisionY3},
        {tank.Hull.CollisionX4, tank.Hull.CollisionY4},
    }

    return vectors
}

func getObjectCollisionVectors(object levels.LevelBlock) []Vector {
    // Define object's boundaries as vectors
    vectors := []Vector{
        {object.X, object.Y},
        {object.X + object.Width, object.Y},
        {object.X + object.Width, object.Y + object.Height},
        {object.X, object.Y + object.Height},
    }

    return vectors
}

func vectorsIntersect(vectors1, vectors2 []Vector) bool {
    // Check for intersections between two sets of vectors

    // Check for intersections on each axis
    for _, axis := range getAxes(vectors1) {
        if !projectionOverlap(axis, vectors1, vectors2) {
            return false
        }
    }

    for _, axis := range getAxes(vectors2) {
        if !projectionOverlap(axis, vectors1, vectors2) {
            return false
        }
    }

    return true
}

// Get the axes perpendicular to the edges of the rectangle
func getAxes(rectVectors []Vector) []Vector {
    axes := make([]Vector, len(rectVectors))

    for i, point := range rectVectors {
        nextPoint := rectVectors[(i+1)%len(rectVectors)]
        edgeVector := Vector{X: nextPoint.X - point.X, Y: nextPoint.Y - point.Y}
        // Get the perpendicular vector (normal) to the edge
        axes[i] = Vector{X: -edgeVector.Y, Y: edgeVector.X}
    }

    return axes
}

// Project vectors onto an axis and check for overlap
func projectionOverlap(axis Vector, vectors1, vectors2 []Vector) bool {
    min1, max1 := projectOntoAxis(axis, vectors1)
    min2, max2 := projectOntoAxis(axis, vectors2)

    // Check for overlap on the axis
    return (min1 <= max2 && max1 >= min2) || (min2 <= max1 && max2 >= min1)
}

// Project vectors onto an axis and return the min and max values
func projectOntoAxis(axis Vector, vectors []Vector) (float64, float64) {
    min, max := dotProduct(axis, vectors[0]), dotProduct(axis, vectors[0])

    for _, point := range vectors[1:] {
        projection := dotProduct(axis, point)
        if projection < min {
            min = projection
        }
        if projection > max {
            max = projection
        }
    }

    return min, max
}

// Dot product of two vectors
func dotProduct(v1, v2 Vector) float64 {
    return v1.X*v2.X + v1.Y*v2.Y
}

func moveActorToPreviousPosition(tank *Tank) {
    // Avoid getting tanks stuck next to level objects. Restore both the
    // previous position and the previous hull angle: a rotation alone (which
    // doesn't change X/Y) can collide with a wall, and if only X/Y were
    // restored the new angle would stick, leaving the corners overlapping
    // and the tank wedged. Recompute the collision box so the corner fields
    // reflect the safe state for the next frame.
    tank.X = tank.PrevX
    tank.Y = tank.PrevY
    tank.Hull.Angle = tank.PrevHullAngle
    updateCollisionBox(tank)
}

func checkProjectileCollisions(tanks *[]Tank, levelObjects []levels.LevelBlock) {
    for ti, t := range *tanks {
        for pi, p := range t.Projectiles {
            if hasProjectileCollidedWithObject(p.X, p.Y, p.Width, p.Height, levelObjects) {
                (*tanks)[ti].Projectiles[pi].Collided = true
            }

            if hasProjectileCollidedWithActor(p.X/gameLogicToScreenXOffset, p.Y/gameLogicToScreenYOffset, tanks, &t) {
                (*tanks)[ti].Projectiles[pi].Collided = true
            }
        }
    }
}

func hasProjectileCollidedWithObject(pX, pY, pWidth, pHeight float64, levelObjects []levels.LevelBlock) bool {
    for i, object := range levelObjects {
        if !object.Border && object.Health > 0 && object.Destructible {
            left := object.X
            right := (object.X + object.Width)
            top := object.Y
            bottom := (object.Y + object.Height)

            if pX/gameLogicToScreenXOffset+pWidth >= left && pX/gameLogicToScreenXOffset <= right &&
                pY/gameLogicToScreenYOffset+pHeight >= top && pY/gameLogicToScreenYOffset <= bottom {

                // Calculate intersection depths along X and Y axes
                dx := math.Min(right-pX/gameLogicToScreenXOffset, pX/gameLogicToScreenXOffset-left)
                dy := math.Min(bottom-pY/gameLogicToScreenYOffset, pY/gameLogicToScreenYOffset-top)

                // Determine the side of collision based on the shallower intersection depth
                if dx < dy {
                    if pX/gameLogicToScreenXOffset+pWidth/2 < left+(right-left)/2 {
                        levels.DeformBlock(&levelObjects[i], "l")
                    } else {
                        levels.DeformBlock(&levelObjects[i], "r")
                    }
                } else {
                    if pY/gameLogicToScreenYOffset+pHeight/2 < top+(bottom-top)/2 {
                        levels.DeformBlock(&levelObjects[i], "t")
                    } else {
                        levels.DeformBlock(&levelObjects[i], "b")
                    }
                }

                return true
            }
        }
    }

    return false
}

func hasProjectileCollidedWithActor(pX, pY float64, tanks *[]Tank, originatingTank *Tank) bool {
    for i := range *tanks {
        tank := &(*tanks)[i]

        // Skip the originating tank to prevent self-collision
        if tank.Name == originatingTank.Name {
            continue
        }

        // Players don't damage each other unless friendly fire is enabled.
        if originatingTank.IsPlayer && tank.IsPlayer && !FriendlyFire {
            continue
        }

        if checkCollision(pX, pY, tank.Hull.CollisionX1, tank.Hull.CollisionY1, tank.Hull.CollisionX2, tank.Hull.CollisionY2,
            tank.Hull.CollisionX3, tank.Hull.CollisionY3, tank.Hull.CollisionX4, tank.Hull.CollisionY4, tank.Hull.Angle) {
            // Collision occurred
            tank.Health -= 50
            tank.LastDamagedBy = originatingTank.Name

            return true
        }
    }

    return false
}

func checkCollision(pX, pY, x1, y1, x2, y2, x3, y3, x4, y4, tankAngle float64) bool {
    // The collision corners (x1..y4) are already in world/screen space,
    // rotated by updateCollisionBox. Testing a point against a
    // rotated rectangle only needs the two adjacent edge vectors from
    // corner 1; no re-rotation of the point is needed (the edges already
    // encode the orientation). Re-rotating by -tankAngle here would
    // double-rotate the point and hit in the wrong place.

    // Adjacent edges from corner 1: corner 2 and corner 4.
    vector1X := x2 - x1
    vector1Y := y2 - y1
    vector2X := x4 - x1
    vector2Y := y4 - y1

    // Vector from corner 1 to the projectile point.
    vectorPX := pX - x1
    vectorPY := pY - y1

    // Project the point onto each edge and check it lies within the
    // rectangle both along edge 1->2 and edge 1->4.
    dot1 := vectorPX*vector1X + vectorPY*vector1Y
    dot2 := vectorPX*vector2X + vectorPY*vector2Y

    // Give the projectile a small footprint instead of testing a single
    // dimensionless point, so glancing/fast shots register. The hull is a
    // 50px square so |edge| = 50; a 5px margin in dot-product units is
    // 5*50 = 250. This only affects projectile-vs-tank, not tank-vs-wall.
    const hitMargin = 250.0

    // Check if the point is inside the rectangle (with margin)
    return dot1 >= -hitMargin && dot1 <= vector1X*vector1X+vector1Y*vector1Y+hitMargin &&
        dot2 >= -hitMargin && dot2 <= vector2X*vector2X+vector2Y*vector2Y+hitMargin
}
