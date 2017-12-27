package frontend

type (
	// Transport …
	Transport interface {
		Encode([]byte) []byte
		Decode(data []byte, atEOF bool) (int, []byte, error)
	}
)

func _next() (int, []byte, error) {
	return 0, nil, nil
}

func _err(err error) (int, []byte, error) {
	return 0, nil, err
}
