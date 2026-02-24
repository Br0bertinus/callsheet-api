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
	CurrentActorID int   `json:"currentActorId"`
	NextActorID    int   `json:"nextActorId"`
	VisitedActorIDs []int `json:"visitedActorIds"`
}

// ValidateStepResponse is the response for POST /game/validate-step.
type ValidateStepResponse struct {
	Valid            bool    `json:"valid"`
	ConnectingMovies []Movie `json:"connectingMovies"`
}
