package frontend

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	backend "github.com/nathanaelle/shitenno/lib/backend"
)

// Dovecot handler
func Dovecot() Handler {
	return &buffHandler{
		Transport: DoveDict,
		Handler:   dovecot,
	}
}

func dovecot(db *backend.HTTPDB, decoder *bufio.Scanner, encoder func([]byte)) error {
	for decoder.Scan() {
		data := decoder.Bytes()

		if data[0] == 'H' {
			continue
		}
		if data[0] != 'L' {
			encoder([]byte{'F'})
			continue
		}

		msg := bytes.SplitN(data[1:], []byte{'/'}, 3)

		query, err := backend.NewQuery(string(msg[1]), backend.DovecotQuery{
			Namespace: string(msg[0]),
			Object:    string(msg[2]),
		})
		if err != nil {
			return err
		}

		res, err := db.Request(query)

		if err != nil {
			encoder([]byte{'F'})
			return err
		}

		switch res.Status {
		case "OK":
			data, err := json.Marshal(res.Data)
			if err != nil {
				encoder([]byte{'F'})
				return fmt.Errorf("strange Resp %+v", res)
			}

			encoder(append([]byte{'O'}, data...))

		case "KO":
			encoder([]byte{'N'})

		default:
			encoder([]byte{'F'})
			return fmt.Errorf("strange Resp %+v", res)
		}

	}

	return decoder.Err()
}
