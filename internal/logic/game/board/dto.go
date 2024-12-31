package board

type RegionDto struct {
	ExternalReference string `json:"id"`
	Name              string `json:"name"`
	Continent         string `json:"continent"`
}

type ContinentDto struct {
	ExternalReference string `json:"id"`
	Name              string `json:"name"`
	BonusTroops       int    `json:"bonusTroops"`
}

type BorderDto struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type BoardDto struct {
	Regions    []RegionDto    `json:"layers"`
	Continents []ContinentDto `json:"continents"`
	Borders    []BorderDto    `json:"links"`
}
