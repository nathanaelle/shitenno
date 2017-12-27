package backend

type (
	Query struct {
		Verb   string      `json:"verb"`
		Object interface{} `json:"object"`
	}
)
