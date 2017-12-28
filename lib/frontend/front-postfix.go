package frontend

import (
	"bufio"
	"bytes"
	"fmt"

	backend "github.com/nathanaelle/shitenno/lib/backend"
)

// Postfix handler
func Postfix() Handler {
	return &buffHandler{
		Transport: NetString,
		Handler:   postfix,
	}
}

func postfix(db *backend.HTTPDB, decoder *bufio.Scanner, encoder func([]byte)) error {
	for decoder.Scan() {
		msg := bytes.SplitN(decoder.Bytes(), []byte{' '}, 2)

		query, err := backend.NewQuery(string(msg[0]), string(msg[1]))
		if err != nil {
			return err
		}

		res, err := db.Request(query)

		if err != nil {
			encoder([]byte("TIMEOUT error in backend"))
			return err
		}

		switch res.Status {
		case "OK":
			data, err := res.Postfix()
			if err != nil {
				encoder([]byte("TIMEOUT error in backend"))
				return fmt.Errorf("strange Resp %+v", res)
			}
			encoder([]byte("OK " + data))

		case "KO":
			encoder([]byte("NOTFOUND "))

		default:
			encoder([]byte("TIMEOUT error in backend"))
			return fmt.Errorf("strange Resp %+v", res)
		}
	}

	return decoder.Err()
}
