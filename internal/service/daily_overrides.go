package service

// dailyOverrides maps a game-day date string ("YYYY-MM-DD", Pacific time) to a
// fixed {startActorID, targetActorID} pair that supersedes the PRNG-derived pair.
//
// Use this for planned editorial overrides — e.g. featuring two actors who
// are in the news, tying the challenge to a film anniversary, etc.
//
// Note: date keys are in America/Los_Angeles (Pacific) time, matching the
// rollover logic in DailyChallenge. The game day begins at 01:00 Pacific.
//
// For an immediate same-day override without redeploying, set the environment
// variable instead (see DailyChallenge in game.go):
//
//	DAILY_CHALLENGE_OVERRIDE=<startActorID>,<targetActorID>
var dailyOverrides = map[string][2]int{
	// Example — uncomment and adjust to activate:
	// "2026-04-01": {31, 287}, // Tom Hanks → Brad Pitt (April Fools special)
}
