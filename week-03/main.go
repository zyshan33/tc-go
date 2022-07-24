package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

func signalInf() error {
	ch := make(chan os.Signal)
	signal.Notify(ch)
	s := <-ch
	switch s {
	case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
		return errors.New("exit by signal")
	default:
		fmt.Println("signal：", s)
	}
	return nil
}

func main() {
	group, errCtx := errgroup.WithContext(context.Background())

	var s myServer
	se := http.Server{
		Handler: s,
		Addr:    ":8080",
	}

	http.Handle("/", s)
	group.Go(func() error {
		defer fmt.Println("g1 return")
		return se.ListenAndServe()
	})

	group.Go(func() error {
		select {
		case <-errCtx.Done():
			return se.Shutdown(errCtx)
		}
		return nil
	})

	group.Go(func() error {
		err := signalInf()
		if err != nil {
			return err
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		fmt.Println("all dead:", err)
	}
}

type myServer struct{}

func (server myServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "halo！")
}
