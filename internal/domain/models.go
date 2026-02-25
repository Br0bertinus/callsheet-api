package domain

// Actor represents a person returned from TMDB search or lookup.
type Actor struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ProfilePath string `json:"profilePath"`
}

// Movie represents a film shared between two actors during step validation.
type Movie struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Year  int    `json:"year"`
}

// ValidateStepRequest is the body for POST /game/validate-step.
type ValidateStepRequest struct {
	CurrentActorID  int   `json:"currentActorId"`
	NextActorID     int   `json:"nextActorId"`
	MovieID         int   `json:"movieId"` // the movie the user claims connects the two actors
	VisitedActorIDs []int `json:"visitedActorIds"`
	VisitedMovieIDs []int `json:"visitedMovieIds"` // movies already used in the chain
}

// ValidateStepResponse is the response for POST /game/validate-step.
type ValidateStepResponse struct {
	Valid            bool    `json:"valid"`
	ConnectingMovies []Movie `json:"connectingMovies"` // all valid shared movies (for client hints)
}
