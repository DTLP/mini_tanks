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
            if !(*tanks)[ti].Player {
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
    // Avoid getting tanks stuck next to level objects
    tank.Hull.X = tank.Hull.PrevX
    tank.Hull.Y = tank.Hull.PrevY
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

        if checkCollision(pX, pY, tank.Hull.CollisionX1, tank.Hull.CollisionY1, tank.Hull.CollisionX2, tank.Hull.CollisionY2,
            tank.Hull.CollisionX3, tank.Hull.CollisionY3, tank.Hull.CollisionX4, tank.Hull.CollisionY4, tank.Hull.Angle) {
            // Collision occurred
            tank.Health -= 50

            return true
        }
    }

    return false
}

func checkCollision(pX, pY, x1, y1, x2, y2, x3, y3, x4, y4, tankAngle float64) bool {
    rotatedPX := math.Cos(-tankAngle)*(pX-x1) - math.Sin(-tankAngle)*(pY-y1) + x1
    rotatedPY := math.Sin(-tankAngle)*(pX-x1) + math.Cos(-tankAngle)*(pY-y1) + y1

    // Calculate vectors from point 1 to the other corners of the rectangle
    vector1X := x2 - x1
    vector1Y := y2 - y1
    vector2X := x3 - x1
    vector2Y := y3 - y1

    // Calculate vectors from point 1 to the rotated projectile point
    vectorPX := rotatedPX - x1
    vectorPY := rotatedPY - y1

    // Calculate dot products
    dot1 := vectorPX*vector1X + vectorPY*vector1Y
    dot2 := vectorPX*vector2X + vectorPY*vector2Y

    // Check if the point is inside the rectangle
    return dot1 >= 0 && dot1 <= vector1X*vector1X+vector1Y*vector1Y &&
        dot2 >= 0 && dot2 <= vector2X*vector2X+vector2Y*vector2Y
}
