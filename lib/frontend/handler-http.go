package frontend

import (
	"fmt"
	"net"
	"net/http"

	backend "github.com/nathanaelle/shitenno/lib/backend"
)

type (
	httpHandler struct {
		http.Server
		db *backend.HTTPDB
	}
)

func (h *httpHandler) Inject(db *backend.HTTPDB) {
	h.db = db
}

func (h *httpHandler) ServeHTTP(hres http.ResponseWriter, hreq *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			panic(r.(error))
		}
	}()

	query, err := backend.NewQuery("nginx", backend.NginxQuery{
		User:    hreq.Header.Get("auth-user"),
		Pass:    hreq.Header.Get("auth-pass"),
		Proto:   hreq.Header.Get("auth-proto"),
		Attempt: hreq.Header.Get("auth-login-attempt"),
		IpCli:   hreq.Header.Get("client-ip"),
	})
	if err != nil {
		hres.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	res, err := h.db.Request(query)

	if err != nil {
		hres.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	data, err := res.Nginx()
	if err != nil {
		hres.WriteHeader(http.StatusInternalServerError)
		panic(fmt.Errorf("strange Resp %+v", res))
	}

	switch res.Status {
	case "OK":
		hres.Header().Set("Auth-Status", "OK")
		hres.Header().Set("Auth-Server", data.Host)
		hres.Header().Set("Auth-Port", data.Port)

		hres.WriteHeader(http.StatusOK)

	case "KO":
		hres.Header().Set("Auth-Status", "Invalid login or password")
		hres.Header().Set("Auth-Wait", "10")

		hres.WriteHeader(http.StatusOK)

	default:
		hres.WriteHeader(http.StatusInternalServerError)
		panic(fmt.Errorf("strange Resp %+v", res))
	}

}

func (h *httpHandler) Serve(l net.Listener) error {
	h.Server.Handler = http.HandlerFunc(h.ServeHTTP)

	return h.Server.Serve(l)
}
