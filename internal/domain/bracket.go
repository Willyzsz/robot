package domain

type Bracket struct {
	ID         string           `json:"id"`
	CategoryID CategoryID       `json:"category_id"`
	Size       int              `json:"size"`
	TeamIDs    []TeamID         `json:"team_ids"`
	Labels     []string         `json:"labels"`
	Rounds     [][]BracketMatch `json:"rounds"`
}

type BracketMatch struct {
	Key     string   `json:"key"`
	Round   int      `json:"round"`
	Slot    int      `json:"slot"`
	TeamA   *TeamID  `json:"team_a"`
	TeamB   *TeamID  `json:"team_b"`
	Winner  *TeamID  `json:"winner"`
	MatchID *MatchID `json:"match_id"`
	Status  string   `json:"status"`
}
