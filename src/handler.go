package	main


import	(
	"fmt"
	"net"
	"bytes"
	"bufio"
	"errors"
	"net/http"
	"encoding/json"
)


type	(
	Handler	interface {
		Serve(net.Listener) error
		Inject(*HTTPDB)
	}

	BuffHandler	struct{
		Handler		func(db *HTTPDB, read *bufio.Scanner, write func([]byte))
		Transport	Transport
		End		<-chan bool
		db		*HTTPDB
	}


	HttpHandler	struct{
		http.Server
		End		<-chan bool
		db		*HTTPDB
	}
)


func	(h *HttpHandler) Inject(db *HTTPDB) {
	h.db	= db
}


func	(h *HttpHandler) ServeHTTP(hres http.ResponseWriter, hreq *http.Request) {
	defer	func() {
		if r := recover(); r != nil {
			exterminate(r.(error))
		}
	}()

	ht	:= map[string]string{
		"auth-user":		"user",
		"auth-pass":		"pass",
		"auth-protocol":	"protocol",
		"auth-login-attempt":	"attempt",
		"client-ip":		"client",
	}
	data	:= make(map[string]string)

	for _,h := range []string{ "auth-user", "auth-pass", "auth-protocol", "auth-login-attempt", "client-ip" } {
		data[ht[h]] = hreq.Header.Get(h)
	}

	res,err	:= h.db.Request(&Query{
		Verb:	"nginx",
		Object:	data,
	})

	if err != nil {
		hres.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	switch res.Status {
	case "OK":
		switch	data := res.Data.(type) {
		case	map[string]interface{}:
			r_data	:= map[string]string{
				"Auth-Status" : "OK",
				"Auth-Server" : data["host"].(string),
				"Auth-Port"   : data["port"].(string),

			}

			for h,v:= range r_data {
				hres.Header().Set(h, v)
			}

			hres.WriteHeader(http.StatusOK)

		default:
			hres.WriteHeader(http.StatusInternalServerError)
			panic(errors.New(fmt.Sprintf("strange Resp %+v", res )))
		}

	case "KO":
		switch	res.Data.(type) {
		case	map[string]interface{}:
			r_data	:= map[string]string{
				"Auth-Status" : "Invalid login or password",
				"Auth-Wait"   : "5",
			}

			for h,v:= range r_data {
				hres.Header().Set(h, v)
			}

			hres.WriteHeader(http.StatusOK)

		default:
			hres.WriteHeader(http.StatusInternalServerError)
			panic(errors.New(fmt.Sprintf("strange Resp %+v", res )))
		}

	default:
		hres.WriteHeader(http.StatusInternalServerError)
		panic(errors.New(fmt.Sprintf("strange Resp %+v", res )))
	}

}


func	(h *HttpHandler) Serve(l net.Listener) error {
	h.Server.Handler = http.HandlerFunc(h.ServeHTTP)

	return h.Server.Serve(l)
}


func	(h *BuffHandler) Inject(db *HTTPDB) {
	h.db	= db
}


func	(h *BuffHandler) Serve(l net.Listener) error {
	for {
		select {
		case	<-h.End:
			return nil

		default:
			fd,err := l.Accept()

			if	err != nil {
				return err
			}

			go h.cope_with(fd)
		}
	}
}

func	(h *BuffHandler) cope_with(con net.Conn) {
	defer	func() {
		con.Close()
		if r := recover(); r != nil {
			exterminate(r.(error))
		}
	}()

	scan	:= bufio.NewScanner(con)
	scan.Split(h.Transport.Decode)

	h.Handler(h.db, scan, func(d []byte){
		con.Write(h.Transport.Encode(d))
	})
}


func postfix(db *HTTPDB, decoder *bufio.Scanner,encoder func([]byte)) {
	for decoder.Scan() {
		msg	:= bytes.SplitN(decoder.Bytes(), []byte{' '}, 2)

		res,err	:= db.Request(&Query{
			Verb:	string(msg[0]),
			Object:	string(msg[1]),
		})

		if err != nil {
			encoder([]byte("TIMEOUT error in backend"))
			panic(err)
		}

		switch res.Status {
		case "OK":
			switch	data := res.Data.(type) {
			case	string:
				encoder([]byte("OK "+data))

			default:
				encoder([]byte("TIMEOUT error in backend"))
				panic(errors.New(fmt.Sprintf("strange Resp %+v", res )))
			}

		case "KO":
			encoder([]byte("NOTFOUND "))

		default:
			encoder([]byte("TIMEOUT error in backend"))
			panic(errors.New(fmt.Sprintf("strange Resp %+v", res )))
		}
	}

	if err := decoder.Err(); err != nil {
		panic(err)
	}
}



func dovecot(db *HTTPDB, decoder *bufio.Scanner,encoder func([]byte)) {
	for decoder.Scan() {
		data	:= decoder.Bytes()

		if data[0] == 'H' {
			continue
		}
		if data[0] != 'L' {
			encoder([]byte{'F'})
			continue
		}

		msg	:= bytes.SplitN(data[1:], []byte{'/'}, 3)

		res,err	:= db.Request(&Query{
			Verb:	string(msg[1]),
			Object:	map[string]string{
				"context": string(msg[0]),
				"object":  string(msg[2]),
			},
		})

		if err != nil {
			encoder([]byte{'F'})
			panic(err)
		}

		switch res.Status {
		case "OK":
			data,err	:= json.Marshal(res.Data)
			if err != nil {
				encoder([]byte{'F'})
				panic(errors.New(fmt.Sprintf("strange Resp %+v", res )))
			}

			encoder(append([]byte{'O'},data...))

		case "KO":
			encoder([]byte{'N'})

		default:
			encoder([]byte{'F'})
			panic(errors.New(fmt.Sprintf("strange Resp %+v", res )))
		}

	}

	if err := decoder.Err(); err != nil {
		panic(err)
	}
}
