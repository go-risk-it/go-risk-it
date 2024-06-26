package dto

type region struct {
	ExternalReference string `json:"id"`
	Name              string `json:"name"`
	Continent         string `json:"continent"`
}

type continent struct {
	ExternalReference string `json:"id"`
	Name              string `json:"name"`
	BonusTroops       int    `json:"bonusTroops"`
}

type border struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type Board struct {
	Regions    []region    `json:"layers"`
	Continents []continent `json:"continents"`
	Borders    []border    `json:"links"`
}
