package lib

import (
	"bytes"
	"fmt"
	"strconv"
)

type (
	// Transport …
	Transport interface {
		Encode([]byte) []byte
		Decode(data []byte, atEOF bool) (int, []byte, error)
	}

	netString struct {
	}

	doveDict struct {
	}
)

var (
	// NetString …
	NetString Transport = new(netString)

	// DoveDict …
	DoveDict Transport = new(doveDict)
)

func _next() (int, []byte, error) {
	return 0, nil, nil
}

func _err(err error) (int, []byte, error) {
	return 0, nil, err
}

func (ns *netString) Encode(data []byte) []byte {
	return []byte(fmt.Sprintf("%d:%s,", len(data), data))
}

func (ns *netString) Decode(data []byte, atEOF bool) (int, []byte, error) {
	if len(data) == 0 {
		return _next()
	}

	pColon := bytes.IndexByte(data, ':')
	if pColon < 0 {
		return _next()
	}

	size64, err := strconv.ParseInt(string(data[0:pColon]), 10, 0)
	size := int(size64)
	if err != nil {
		return _err(err)
	}
	pColon++

	if len(data) < (size + pColon + 1) {
		return _next()
	}

	if data[pColon+size] != ',' {
		return _err(fmt.Errorf("no , in [%s]", data[0:pColon+size+1]))
	}

	return pColon + size + 1, data[pColon : pColon+size], nil
}

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
