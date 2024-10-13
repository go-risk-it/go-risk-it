package request

type CardCombination struct {
	CardIDs []int64 `json:"cardIds"`
}

type CardsMove struct {
	Combinations []CardCombination `json:"combinations"`
}
