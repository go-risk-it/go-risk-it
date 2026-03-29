package game

type PhaseType string

const (
	Cards     PhaseType = "cards"
	Deploy    PhaseType = "deploy"
	Attack    PhaseType = "attack"
	Conquer   PhaseType = "conquer"
	Reinforce PhaseType = "reinforce"
)
