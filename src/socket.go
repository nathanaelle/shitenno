package	main

import	(
	"os"
	"log"
	"net"
	"time"
)


type	(

	Listener struct {
		*net.UnixListener
		end	<-chan struct{}
	}

	EOConn struct {
	}


	Conn	struct {
		net.Conn
		end	<-chan struct{}
	}

)

func (_ *EOConn) Error() string {
	return "end of connection"
}

func (_ *EOConn) Timeout() bool {
	return true
}

func (_ *EOConn) Temporary() bool {
	return false
}


func	create_socket(l *log.Logger, socket string, uid,gid int, end <-chan struct{}) *Listener {
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

	return	&Listener{ conn, end }
}


func (lst *Listener)Accept() (net.Conn,error) {
	for {
		select {
		case	<-lst.end:
			return nil,new(EOConn)

		default:
			lst.SetDeadline(time.Now().Add(LISTEN_EXPIRE))
			fd,err := lst.UnixListener.Accept()
			switch	{
			case	err == nil:
				return &Conn{ fd, lst.end }, nil

			default:
				if nerr,ok := err.(net.Error); !ok || !nerr.Timeout() {
					return nil,err
				}
			}
		}
	}
}


func (conn Conn) Read(b []byte) (n int, err error) {
	n1 := 0

	for {
		select {
		case	<-conn.end:
			return 0,new(EOConn)

		default:
			conn.SetReadDeadline(time.Now().Add(LISTEN_EXPIRE))
			n1,err = conn.Conn.Read(b[n:])
			n+=n1
			if err == nil || n == len(b) {
				return n,nil
			}

			if nerr,ok := err.(net.Error); !ok || !nerr.Timeout() {
				return
			}
		}
	}
}
