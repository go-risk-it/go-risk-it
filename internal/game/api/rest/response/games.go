package response

type Game struct {
	ID                   int64 `json:"id"`
	NumberOfParticipants int64 `json:"numberOfParticipants"`
}

type Games struct {
	Games []Game `json:"games"`
}
