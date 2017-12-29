package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	sd "github.com/nathanaelle/sdialog"
	shitenno "github.com/nathanaelle/shitenno/lib"
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

		go systemdLogError(errChan)

	case !*stderr:
		co, errChan, err := (syslog.Dialer{
			FlushDelay: 500 * time.Millisecond,
		}).Dial("local", "", syslog.T_ZEROENDED)

		if err != nil {
			log.Fatal(err)
		}
		slog, _ = syslog.New(co, *priority, appName)

		go systemdLogError(errChan)
	}

	wg, end, update := signalCatcher()

	sd.Notify(sd.Status("Starting…"))
	if err := sd.Watchdog(end, wg); err != nil {
		return
	}

	shitenno, err := shitenno.SummonShitenno(confPath, slog, wg, end.Done(), update)
	if err != nil {
		log.Fatal(err)
	}
	sd.Notify(sd.MainPid(os.Getpid()), sd.Ready(), sd.Status("Waiting requests…"))

	shitenno.SummonGardians()

	shitenno.End()
}

func systemdLogError(errChan <-chan error) {
	for e := range errChan {
		sd.SD_ERR.LogError(e)
	}
}

func signalCatcher() (*sync.WaitGroup, context.Context, <-chan struct{}) {
	wg := new(sync.WaitGroup)
	end, cancel := context.WithCancel(context.Background())
	update := make(chan struct{})

	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	wg.Add(1)
	go func() {
		for sig := range signalChannel {
			switch sig {
			case os.Interrupt, syscall.SIGTERM:
				wg.Done()
				close(signalChannel)
				cancel()
				close(update)

				return

			case syscall.SIGHUP:
				update <- struct{}{}
			}
		}
	}()

	return wg, end, update
}
