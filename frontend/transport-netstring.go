package frontend

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
)

type (
	netString struct {
	}
)

var (
	// NetString implementation : https://en.wikipedia.org/wiki/Netstring
	NetString Transport = new(netString)
)

func (ns *netString) Encode(data []byte) []byte {
	lendata := len(data)
	loglen := math.Floor(math.Log10(float64(len(data))))
	ret := make([]byte, 0, 3+int(loglen))
	ret = strconv.AppendInt(ret, int64(lendata), 10)
	ret = append(ret, ':')
	ret = append(ret, data...)
	ret = append(ret, ',')

	return ret // []byte(fmt.Sprintf("%d:%s,", len(data), data))
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
