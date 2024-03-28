package request

type CreateGame struct {
	Players []string `json:"players"`
}
