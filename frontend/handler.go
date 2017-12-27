package frontend

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"

	backend "github.com/nathanaelle/shitenno/backend"
)

type (
	Handler interface {
		Serve(net.Listener) error
		Inject(*backend.HTTPDB)
	}

	BuffHandler struct {
		Handler   func(db *backend.HTTPDB, read *bufio.Scanner, write func([]byte)) error
		Transport Transport
		db        *backend.HTTPDB
	}

	HttpHandler struct {
		http.Server
		db *backend.HTTPDB
	}
)

func (h *HttpHandler) Inject(db *backend.HTTPDB) {
	h.db = db
}

func (h *HttpHandler) ServeHTTP(hres http.ResponseWriter, hreq *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			panic(r.(error))
		}
	}()

	ht := map[string]string{
		"auth-user":          "user",
		"auth-pass":          "pass",
		"auth-protocol":      "protocol",
		"auth-login-attempt": "attempt",
		"client-ip":          "client",
	}
	data := make(map[string]string)

	for _, h := range []string{"auth-user", "auth-pass", "auth-protocol", "auth-login-attempt", "client-ip"} {
		data[ht[h]] = hreq.Header.Get(h)
	}

	res, err := h.db.Request(&backend.Query{
		Verb:   "nginx",
		Object: data,
	})

	if err != nil {
		hres.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	switch res.Status {
	case "OK":
		switch data := res.Data.(type) {
		case map[string]interface{}:
			rData := map[string]string{
				"Auth-Status": "OK",
				"Auth-Server": data["host"].(string),
				"Auth-Port":   data["port"].(string),
			}

			for h, v := range rData {
				hres.Header().Set(h, v)
			}

			hres.WriteHeader(http.StatusOK)

		default:
			hres.WriteHeader(http.StatusInternalServerError)
			panic(fmt.Errorf("strange Resp %+v", res))
		}

	case "KO":
		switch res.Data.(type) {
		case map[string]interface{}:
			rData := map[string]string{
				"Auth-Status": "Invalid login or password",
				"Auth-Wait":   "5",
			}

			for h, v := range rData {
				hres.Header().Set(h, v)
			}

			hres.WriteHeader(http.StatusOK)

		default:
			hres.WriteHeader(http.StatusInternalServerError)
			panic(fmt.Errorf("strange Resp %+v", res))
		}

	default:
		hres.WriteHeader(http.StatusInternalServerError)
		panic(fmt.Errorf("strange Resp %+v", res))
	}

}

func (h *HttpHandler) Serve(l net.Listener) error {
	h.Server.Handler = http.HandlerFunc(h.ServeHTTP)

	return h.Server.Serve(l)
}

func (h *BuffHandler) Inject(db *backend.HTTPDB) {
	h.db = db
}

func (h *BuffHandler) Serve(l net.Listener) error {
	defer l.Close()

	for {
		fd, err := l.Accept()

		if err != nil {
			return err
		}

		go h.copeWith(fd)
	}
}

func (h *BuffHandler) copeWith(con net.Conn) {
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
