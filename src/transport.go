package	main

import	(
	"fmt"
	"bytes"
	"errors"
	"strconv"
)


type	(
	Transport	interface {
		Encode([]byte) []byte
		Decode(data []byte, atEOF bool) (int, []byte, error)
	}


	NetString	struct {
	}

	DoveDict	struct {
	}
)

var (
	T_NetString	= new(NetString)
	T_DoveDict	= new(DoveDict)
)




func _next() (int,[]byte,error) {
	return 0, nil, nil
}

func _err(err error) (int,[]byte,error) {
	return 0, nil, err
}



func (_ *NetString) Encode(data []byte) []byte {
	return []byte(fmt.Sprintf("%d:%s,",len(data),data))
}


func (_ *NetString) Decode(data []byte, atEOF bool) (int, []byte, error) {
	if len(data) == 0 {
		return _next()
	}

	p_colon	:= bytes.IndexByte(data, ':')
	if p_colon < 0 {
		return _next()
	}

	size64,err := strconv.ParseInt(string(data[0:p_colon]), 10, 0)
	size:=int(size64)
	if err != nil {
		return _err(err)
	}
	p_colon++

	if len(data) < (size+p_colon+1) {
		return _next()
	}

	if data[p_colon+size] != ',' {
		return _err(errors.New(fmt.Sprintf("no , in [%s]", data[0:p_colon+size+1])))
	}

	return p_colon+size+1, data[p_colon:p_colon+size], nil
}



func (_ *DoveDict) Encode(data []byte) []byte {
	return append(data, '\n' )
}


func (_ *DoveDict) Decode(data []byte, atEOF bool) (int, []byte, error) {
	if len(data) == 0 {
		return _next()
	}

	p_nl	:= bytes.IndexByte(data, '\n')
	if p_nl < 0 {
		return _next()
	}

	return p_nl+1, data[0:p_nl], nil
}
