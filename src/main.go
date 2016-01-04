package main

import	(
	"os"
	"fmt"
	"log"
	"net"
	"flag"
	"runtime"
	"syscall"
	"os/signal"

	types	"github.com/nathanaelle/useful.types"

)


const	DEFAULT_CONF	types.Path	= "/etc/shitenno.conf"


func	main()  {
	conf_path	:= new(types.Path)
	*conf_path	= DEFAULT_CONF

	var	numcpu	= flag.Int("cpu", 1, "maximum number of logical CPU that can be executed simultaneously")
	flag.Var(conf_path, "conf", "path to the director" )

	flag.Parse()

	switch {
		case *numcpu >runtime.NumCPU():	runtime.GOMAXPROCS(runtime.NumCPU())
		case *numcpu <1:		runtime.GOMAXPROCS(1)
		default:			runtime.GOMAXPROCS(*numcpu)
	}

	end,_		:= SignalCatcher()
	shitenno,err	:= SummonShitenno( conf_path.String() )

	if err != nil {
		log.Fatal(err)
	}

	if shitenno.Nginx != nil {
		go shitenno.Summon(shitenno.Nginx, &HttpHandler {
			End:		end,
		})
	}

	if shitenno.Postfix != nil {
		go shitenno.Summon(shitenno.Postfix, &BuffHandler {
			End:		end,
			Transport:	T_NetString,
			Handler:	postfix,
		})
	}

	if shitenno.DoveCot != nil {
		go shitenno.Summon(shitenno.DoveCot, &BuffHandler {
			End:		end,
			Transport:	T_DoveDict,
			Handler:	dovecot,
		})
	}

	shitenno.End(end)
}


func SignalCatcher() (<-chan bool,<-chan bool)  {
	end	:= make(chan bool)
	update	:= make(chan bool)

	go func() {
		signalChannel	:= make(chan os.Signal)

		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

		defer close(signalChannel)
		defer close(update)
		defer close(end)

		for sig := range signalChannel {
			switch sig {
			case os.Interrupt, syscall.SIGTERM:
				return

			case syscall.SIGHUP:
				update <- true
			}
		}
	}()

	return end,update
}


func	create_socket(socket string, uid,gid int) net.Listener {
	conn, err := net.ListenUnix("unix",  &net.UnixAddr { socket, "unix" } )
	for err != nil {
		switch	err.(type) {
			case	*net.OpError:
				if err.(*net.OpError).Err.Error() != "bind: address already in use" {
					fmt.Printf( "%s : %s", "Listen "+socket , err )
				}

			default:
				fmt.Printf( "%s : %s", "Listen "+socket , err )
		}

		if _, r_err := os.Stat(socket); r_err != nil {
			fmt.Printf( "%s : %s", "lstat "+socket , err )
		}
		os.Remove(socket)

		conn, err = net.ListenUnix("unix",  &net.UnixAddr { socket, "unix" } )
	}
	os.Chown(socket, uid, gid)

	return	conn
}
