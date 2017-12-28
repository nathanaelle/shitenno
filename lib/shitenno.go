package shitenno

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/naoina/toml"

	backend "github.com/nathanaelle/shitenno/lib/backend"
	frontend "github.com/nathanaelle/shitenno/lib/frontend"
	syslog "github.com/nathanaelle/syslog5424"
)

type (
	stringer interface {
		String() string
	}

	Shitenno struct {
		RemoteURL    string
		SocketPrefix string
		HPKP         []string
		CertPool     string
		ClientCert   string

		Nginx   *GenericConf
		Postfix *GenericConf
		DoveCot *GenericConf

		wg *sync.WaitGroup
		db *backend.HTTPDB

		syslog *syslog.Syslog
		log    *log.Logger

		end    <-chan struct{}
		update <-chan struct{}
		mEnd   chan struct{}
	}

	GenericConf struct {
		UID    int
		GID    int
		Socket string
	}
)

const (
	IOTimeOut time.Duration = 200 * time.Millisecond
)

func SummonShitenno(confPath stringer, slog *syslog.Syslog, wg *sync.WaitGroup, end, update <-chan struct{}) (*Shitenno, error) {
	conf := &Shitenno{
		SocketPrefix: "/var/run/shitenno.",

		log:    slog.Channel(syslog.LOG_INFO).Logger(""),
		end:    end,
		update: update,
		wg:     wg,
	}

	f, err := os.Open(confPath.String())
	if err != nil {
		return nil, err
	}

	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	err = toml.Unmarshal(buf, conf)
	if err != nil {
		return nil, err
	}

	conf.db, err = backend.NewDB(conf.RemoteURL, conf.CertPool, conf.ClientCert, conf.HPKP)
	if err != nil {
		return nil, err
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

	return conf, nil
}

func (shitenno *Shitenno) End() {
	for {
		select {
		case <-shitenno.end:
			close(shitenno.mEnd)
			shitenno.log.Println("Waiting")
			shitenno.wg.Wait()
			return

		// TODO update process is quite ugly
		case <-shitenno.update:
			close(shitenno.mEnd)
			shitenno.SummonGardians()
		}
	}
}

func (shitenno *Shitenno) SummonGardians() {
	shitenno.mEnd = make(chan struct{})

	if shitenno.Nginx != nil {
		go shitenno.Summon(shitenno.Nginx, frontend.Nginx())
	}

	if shitenno.Postfix != nil {
		go shitenno.Summon(shitenno.Postfix, frontend.Postfix())
	}

	if shitenno.DoveCot != nil {
		go shitenno.Summon(shitenno.DoveCot, frontend.Dovecot())
	}
}

func (shitenno *Shitenno) Summon(c *GenericConf, handler frontend.Handler) {
	for {
		conn := createSocket(shitenno.log, shitenno.SocketPrefix+c.Socket, c.UID, c.GID, shitenno.mEnd, shitenno.wg)
		handler.Inject(shitenno.db)

		switch err := handler.Serve(conn); err {
		case nil:
			shitenno.log.Printf("Respawn %s for no reason", c.Socket)

		case io.EOF:
			return

		default:
			shitenno.log.Printf("Respawn %s : %s", c.Socket, err.Error())
		}
	}
}
