package backend

type (
	Reply struct {
		Verb   string      `json:"verb"`
		Object interface{} `json:"object"`
		Status string      `json:"status"`
		Data   interface{} `json:"data"`
	}
)
