package frontend

import (
	"bufio"
	"bytes"
	"fmt"

	backend "github.com/nathanaelle/shitenno/backend"
)

// Postfix handler
func Postfix() Handler {
	return &BuffHandler{
		Transport: NetString,
		Handler:   postfix,
	}
}

func postfix(db *backend.HTTPDB, decoder *bufio.Scanner, encoder func([]byte)) error {
	for decoder.Scan() {
		msg := bytes.SplitN(decoder.Bytes(), []byte{' '}, 2)

		res, err := db.Request(&backend.Query{
			Verb:   string(msg[0]),
			Object: string(msg[1]),
		})

		if err != nil {
			encoder([]byte("TIMEOUT error in backend"))
			return err
		}

		switch res.Status {
		case "OK":
			switch data := res.Data.(type) {
			case string:
				encoder([]byte("OK " + data))

			default:
				encoder([]byte("TIMEOUT error in backend"))
				return fmt.Errorf("strange Resp %+v", res)
			}

		case "KO":
			encoder([]byte("NOTFOUND "))

		default:
			encoder([]byte("TIMEOUT error in backend"))
			return fmt.Errorf("strange Resp %+v", res)
		}
	}

	return decoder.Err()
}
