package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/nathanaelle/shitenno"
	syslog "github.com/nathanaelle/syslog5424"
	types "github.com/nathanaelle/useful.types"
)

const (
	appName     string          = "shitenno"
	defaultConf types.Path      = "/etc/shitenno.conf"
	defaultPrio syslog.Priority = (syslog.LOG_DAEMON | syslog.LOG_WARNING)
)

func main() {
	var slog *syslog.Syslog

	var priority *syslog.Priority
	*priority = defaultPrio

	confPath := new(types.Path)
	*confPath = defaultConf

	numCPU := flag.Int("cpu", 1, "maximum number of logical CPU that can be executed simultaneously")
	stderr := flag.Bool("stderr", false, "optional overwrite of DevLog with stderr")
	flag.Var(priority, "priority", "optional overwrite of log priority in syslog format facility.severity")
	flag.Var(confPath, "conf", "path to the director")

	flag.Parse()

	switch {
	case *numCPU > runtime.NumCPU():
		runtime.GOMAXPROCS(runtime.NumCPU())
	case *numCPU < 1:
		runtime.GOMAXPROCS(1)
	default:
		runtime.GOMAXPROCS(*numCPU)
	}

	switch {
	case *stderr:
		co, errChan, err := (syslog.Dialer{
			FlushDelay: 100 * time.Millisecond,
		}).Dial("stdio", "stderr", syslog.T_LFENDED)

		if err != nil {
			log.Fatal(err)
		}
		slog, _ = syslog.New(co, *priority, appName)

	case !*stderr:
		co, errChan, err := (syslog.Dialer{
			FlushDelay: 500 * time.Millisecond,
		}).Dial("local", "", syslog.T_ZEROENDED)

		if err != nil {
			log.Fatal(err)
		}
		slog, _ = syslog.New(co, *priority, appName)
	}

	end, update := signalCatcher()

	shitenno, err := shitenno.SummonShitenno(confPath, slog, end, update)
	if err != nil {
		log.Fatal(err)
	}

	shitenno.SummonGardians()

	shitenno.End()
}

func signalCatcher() (<-chan struct{}, <-chan struct{}) {
	end := make(chan struct{})
	update := make(chan struct{})

	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for sig := range signalChannel {
			switch sig {
			case os.Interrupt, syscall.SIGTERM:
				close(signalChannel)
				close(end)
				close(update)

				return

			case syscall.SIGHUP:
				update <- struct{}{}
			}
		}
	}()

	return end, update
}
