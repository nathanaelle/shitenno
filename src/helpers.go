package	main

import	(
	"os"
	"log"
	"net"
	"syscall"
	"os/signal"
)



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


func	create_socket(l *log.Logger, socket string, uid,gid int) *net.UnixListener {
	conn, err := net.ListenUnix("unix",  &net.UnixAddr { socket, "unix" } )
	for err != nil {
		switch	err.(type) {
			case	*net.OpError:
				if err.(*net.OpError).Err.Error() != "bind: address already in use" {
					l.Printf( "Listen %s : %s", socket , err )
				}

			default:
				l.Printf( "Listen %s : %s", socket , err )
		}

		if _, r_err := os.Stat(socket); r_err != nil {
			l.Printf( "Lstat %s : %s", socket , err )
		}
		os.Remove(socket)

		conn, err = net.ListenUnix("unix",  &net.UnixAddr { socket, "unix" } )
	}
	os.Chown(socket, uid, gid)

	return	conn
}
