package	main

import	(
	"os"
	"log"
	"net"
	"time"
	"sync"
)


type	(

	Listener struct {
		*net.UnixListener
		end	<-chan struct{}
		wg	*sync.WaitGroup
	}

	EOConn struct {
	}


	Conn	struct {
		net.Conn
		end	<-chan struct{}
		wg	*sync.WaitGroup
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


func	create_socket(l *log.Logger, socket string, uid,gid int, end <-chan struct{},wg *sync.WaitGroup) *Listener {
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

	wg.Add(1)
	return	&Listener{ conn, end, wg }
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
				lst.wg.Add(1)
				return &Conn{ fd, lst.end, lst.wg }, nil

			default:
				if nerr,ok := err.(net.Error); !ok || !nerr.Timeout() {
					return nil,err
				}
			}
		}
	}
}

func (lst *Listener)Close() (err error) {
	err = lst.UnixListener.Close()
	lst.wg.Done()
	return
}



func (conn *Conn) Read(b []byte) (n int, err error) {
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
				conn.SetReadDeadline(time.Now().Add(time.Hour))
				return n,nil
			}

			if nerr,ok := err.(net.Error); !ok || !nerr.Timeout() {
				return
			}
		}
	}
}

func (conn *Conn)Close() (err error) {
	err = conn.Conn.Close()
	conn.wg.Done()
	return
}
