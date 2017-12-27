package shitenno

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/naoina/toml"

	backend "github.com/nathanaelle/shitenno/backend"
	syslog "github.com/nathanaelle/syslog5424"
	types "github.com/nathanaelle/useful.types"
)

type (
	Shitenno struct {
		SyslogDest string
		DevLog     *types.Path
		Priority   *syslog.Priority

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

		end    <-chan bool
		update <-chan bool
		mEnd   chan struct{}
	}

	GenericConf struct {
		UID    int
		GID    int
		Socket string
	}
)

const (
	IOTimeOut   time.Duration   = 200 * time.Millisecond
	AppName     string          = "shitenno"
	DefaultConf types.Path      = "/etc/shitenno.conf"
	DefaultPrio syslog.Priority = (syslog.LOG_DAEMON | syslog.LOG_WARNING)
)

func SummonShitenno() *Shitenno {
	var priority *syslog.Priority

	confPath := new(types.Path)
	*confPath = DefaultConf

	numCPU := flag.Int("cpu", 1, "maximum number of logical CPU that can be executed simultaneously")
	stderr := flag.Bool("stderr", false, "optional overwrite of DevLog with stderr")
	optPrio := flag.String("priority", "", "optional overwrite of log priority in syslog format facility.severity")
	flag.Var(confPath, "conf", "path to the director")

	flag.Parse()

	// TODO flag knows nothing about nil `value Value`
	// TODO so this empty string is an ugly way to detect the "I don't want any default value"
	if *optPrio != "" {
		priority = new(syslog.Priority)
		err := priority.Set(*optPrio)
		if err != nil {
			log.Fatal(err)
		}
	}

	switch {
	case *numCPU > runtime.NumCPU():
		runtime.GOMAXPROCS(runtime.NumCPU())
	case *numCPU < 1:
		runtime.GOMAXPROCS(1)
	default:
		runtime.GOMAXPROCS(*numCPU)
	}

	conf := &Shitenno{
		SocketPrefix: "/var/run/shitenno.",

		wg: new(sync.WaitGroup),
	}

	f, err := os.Open(confPath.String())
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	err = toml.Unmarshal(buf, conf)
	if err != nil {
		log.Fatal(err)
	}

	conf.db, err = backend.NewDB(conf.RemoteURL, conf.CertPool, conf.ClientCert, conf.HPKP)
	if err != nil {
		log.Fatal(err)
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

	switch {
	case priority != nil:
		conf.Priority = priority

	case priority == nil && conf.Priority == nil:
		conf.Priority = new(syslog.Priority)
		*conf.Priority = DefaultPrio
	}

	switch {
	case *stderr:
		co, errChan, err := (syslog.Dialer{
			FlushDelay: 100 * time.Millisecond,
		}).Dial("stdio", "stderr", syslog.T_LFENDED)

		if err != nil {
			log.Fatal(err)
		}
		conf.syslog, _ = syslog.New(co, *conf.Priority, AppName)

	case !*stderr && conf.DevLog != nil:
		co, errChan, err := (syslog.Dialer{
			FlushDelay: 500 * time.Millisecond,
		}).Dial("local", conf.DevLog.String(), syslog.T_ZEROENDED)

		if err != nil {
			log.Fatal(err)
		}
		conf.syslog, _ = syslog.New(co, *conf.Priority, AppName)

	case !*stderr && conf.DevLog == nil:
		co, errChan, err := (syslog.Dialer{
			FlushDelay: 500 * time.Millisecond,
		}).Dial("local", "", syslog.T_ZEROENDED)

		if err != nil {
			log.Fatal(err)
		}
		conf.syslog, _ = syslog.New(co, *conf.Priority, AppName)
	}

	conf.log = conf.syslog.Channel(syslog.LOG_INFO).Logger("")

	conf.end, conf.update = SignalCatcher()

	return conf
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
		go shitenno.Summon(shitenno.Nginx, &HttpHandler{})
	}

	if shitenno.Postfix != nil {
		go shitenno.Summon(shitenno.Postfix, &BuffHandler{
			Transport: T_NetString,
			Handler:   postfix,
		})
	}

	if shitenno.DoveCot != nil {
		go shitenno.Summon(shitenno.DoveCot, &BuffHandler{
			Transport: T_DoveDict,
			Handler:   dovecot,
		})
	}
}

func (shitenno *Shitenno) Summon(c *GenericConf, handler Handler) {
	for {
		conn := create_socket(shitenno.log, shitenno.SocketPrefix+c.Socket, c.UID, c.GID, shitenno.mEnd, shitenno.wg)
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
