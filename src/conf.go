package main

import	(
	"os"
	"fmt"
	"sync"
	"io/ioutil"

	"github.com/naoina/toml"
)


type	(
	Shitenno	struct {
		DevLog		string
		RemoteURL	string
		SocketPrefix	string
		HPKP		[]string
		CertPool	string
		ClientCert	string

		Nginx	*GenericConf
		Postfix	*GenericConf
		DoveCot	*GenericConf

		wg	*sync.WaitGroup
		db	*HTTPDB
	}

	GenericConf struct {
		UID	int
		GID	int
		Socket	string
	}
)


func SummonShitenno(file string) (*Shitenno,error) {
	conf := &Shitenno {
		DevLog:		"/dev/log",
		SocketPrefix:	"/var/run/shitenno.",
		wg:		new(sync.WaitGroup),
	}

	f,_	:= os.Open(file)
	defer f.Close()

	buf,_	:= ioutil.ReadAll(f)
	err	:= toml.Unmarshal(buf, conf)
	if err != nil {
		return	nil, err
	}

	conf.db,err	= NewDB( conf.RemoteURL, conf.CertPool, conf.ClientCert, conf.HPKP )
	if err != nil {
		return	nil, err
	}

	if conf.Nginx != nil && conf.Nginx.Socket == "" {
		conf.Nginx.Socket = "nginx"
	}

	if conf.Postfix != nil && conf.Postfix.Socket == "" {
		conf.Postfix.Socket = "postfix"
	}

	if conf.DoveCot != nil && conf.DoveCot.Socket == "" {
		conf.DoveCot.Socket = "dovecot"
	}

	return	conf, nil
}


func	(s *Shitenno) End(end <-chan bool)  {
	select {
	case <-end:
		fmt.Println("Waiting")
		s.wg.Wait()
	}
}

func	(s *Shitenno) Summon(c *GenericConf, handler Handler) {
	s.wg.Add(1)
	defer	s.wg.Done()

	conn	:= create_socket(s.SocketPrefix + c.Socket, c.UID, c.GID)
	defer	conn.Close()

	handler.Inject(s.db)

	err	:= handler.Serve(conn)
	if err != nil {
		panic(err)
	}
}
