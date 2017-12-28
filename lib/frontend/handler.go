package frontend

import (
	"net"

	backend "github.com/nathanaelle/shitenno/lib/backend"
)

type (
	Handler interface {
		Serve(net.Listener) error
		Inject(*backend.HTTPDB)
	}
)
