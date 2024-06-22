package board

type Region struct {
	ExternalReference string `json:"id"`
	Name              string `json:"name"`
	Continent         string `json:"continent"`
}
type Continent struct {
	ExternalReference string `json:"id"`
	Name              string `json:"name"`
	BonusTroops       int    `json:"bonusTroops"`
}

type Border struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type Board struct {
	Regions    []Region    `json:"layers"`
	Continents []Continent `json:"continents"`
	Borders    []Border    `json:"links"`
}
