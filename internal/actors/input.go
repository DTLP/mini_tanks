package actors

// Input represents a player's button state, used to drive a Tank remotely.
type Input struct {
	HullLeft, HullRight, Forward, Back, TurretLeft, TurretRight, Fire bool
}
