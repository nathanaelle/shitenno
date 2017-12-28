package frontend

import (
	"bufio"
	"io"
	"net"

	backend "github.com/nathanaelle/shitenno/lib/backend"
)

type (
	buffHandler struct {
		Handler   func(db *backend.HTTPDB, read *bufio.Scanner, write func([]byte)) error
		Transport Transport
		db        *backend.HTTPDB
	}
)

func (h *buffHandler) Inject(db *backend.HTTPDB) {
	h.db = db
}

func (h *buffHandler) Serve(l net.Listener) error {
	defer l.Close()

	for {
		fd, err := l.Accept()

		if err != nil {
			return err
		}

		go h.copeWith(fd)
	}
}

func (h *buffHandler) copeWith(con net.Conn) {
	defer func() {
		con.Close()
		if r := recover(); r != nil {
			panic(r.(error))
		}
	}()

	scan := bufio.NewScanner(con)
	scan.Split(h.Transport.Decode)

	err := h.Handler(h.db, scan, func(d []byte) {
		con.Write(h.Transport.Encode(d))
	})

	switch err {
	case nil:
		return

	case io.EOF:
		return

	default:
		panic(err)
	}
}
