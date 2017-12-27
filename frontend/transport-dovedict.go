package frontend

import "bytes"

type (
	doveDict struct {
	}
)

var (
	// DoveDict …
	DoveDict Transport = new(doveDict)
)

func (dcd *doveDict) Encode(data []byte) []byte {
	return append(data, '\n')
}

func (dcd *doveDict) Decode(data []byte, atEOF bool) (int, []byte, error) {
	if len(data) == 0 {
		return _next()
	}

	pNl := bytes.IndexByte(data, '\n')
	if pNl < 0 {
		return _next()
	}

	return pNl + 1, data[0:pNl], nil
}
