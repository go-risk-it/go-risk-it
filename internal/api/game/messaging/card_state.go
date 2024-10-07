package messaging

type CardType string

const (
	Cavalry   CardType = "cavalry"
	Infantry  CardType = "infantry"
	Artillery CardType = "artillery"
	Jolly     CardType = "jolly"
)

type Card struct {
	Type   CardType `json:"type"`
	Region string   `json:"region"`
}

type CardState struct {
	Cards []Card `json:"cards"`
}
