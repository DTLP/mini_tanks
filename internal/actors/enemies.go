package actors

import (
    // "fmt"
	// "math/rand"
)

var (
	enemies         []Tank
	MaxEnemies      = 3
	TotalEnemiesNum = 10
    killedEnemies   = 0
)

// Define a list of enemy names
var enemyNames = []string{"Albert", "Allen", "Bert", "Bob",
						"Cecil", "Clarence", "Elliot", "Elmer",
						"Ernie", "Eugene", "Fergus", "Ferris",
						"Frank", "Frasier", "Fred", "George",
						"Graham", "Harvey",  "Irwin", "Larry",
						"Lester", "Marvin", "Neil", "Niles",
						"Oliver", "Opie",  "Ryan", "Toby",
						"Ulric", "Ulysses", "Uri", "Waldo",
						"Wally", "Walt", "Wesley", "Yanni",
						"Yogi", "Yuri"}
var levelEnemyNames []string

func CheckEnemyCount(tanks []Tank) []Tank {
	// Check if there are enough enemy tanks on the screen
	count := CountNonPlayerTanks(tanks)

    if count < MaxEnemies && killedEnemies+count < TotalEnemiesNum {
		// Spawn more if needed
        return AddEnemy(tanks)
    }

    return tanks
}

func CountNonPlayerTanks(tanks []Tank) int {
	count := 0

	for _, tank := range tanks {
		if !tank.IsPlayer {
			count++
		}
	}

	return count
}

func AddEnemy(tanks []Tank) []Tank {
    // Append a new enemy to the slice
    newEnemy := NewTank(getUniqueEnemyName())

    updateCollisionBox(&newEnemy)

    tanks = append(tanks, newEnemy)

    return tanks
}

func getUniqueEnemyName() string {
    if len(levelEnemyNames) == 0 {
		// Repopulate LevelEnemyNames with the original names if empty
		levelEnemyNames = append(levelEnemyNames, enemyNames...)
    }
	name := levelEnemyNames[0]
	levelEnemyNames = levelEnemyNames[1:]

	return name
}

func NoEnemiesLeft(tanks []Tank) bool {
    if CountNonPlayerTanks(tanks) == 0 && killedEnemies == TotalEnemiesNum {
		// Reset killed enemy count for next level
		killedEnemies = 0

        return true
    }

    return false
}

func ResetCounter() {
    killedEnemies = 0
}