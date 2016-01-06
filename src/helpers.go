package	main

import	(
	"os"
	"log"
	"reflect"
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


func exterminate(err error)  {
	var s reflect.Value

	if err == nil {
		return
	}

	s_t	:= reflect.ValueOf(err)

	for  s_t.Kind() == reflect.Ptr {
		s_t = s_t.Elem()
	}

	switch s_t.Kind() {
		case reflect.Interface:	s = s_t.Elem()
		default:		s = s_t
	}

	typeOfT := s.Type()
	pkg	:= typeOfT.PkgPath() + "/" + typeOfT.Name()

	log.Printf("\n------------------------------------\nKind : %d %d\n%s\n\n", s_t.Kind(), s.Kind(), err.Error())

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if f.CanInterface() {
			log.Printf("%s %d: %s %s = %v\n", pkg, i, typeOfT.Field(i).Name, f.Type(), f.Interface())
		} else {
			log.Printf("%s %d: %s %s = %s\n", pkg, i, typeOfT.Field(i).Name, f.Type(), f.String())
		}
	}

}
