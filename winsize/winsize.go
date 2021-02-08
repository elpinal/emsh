package winsize

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

type Window struct {
	Width  int
	Height int
}

func Get() (Window, error) {
	w, h, err := terminal.Get(syscall.Stdin)
	if err != nil {
		return Window{}, err
	}
	return Window{Width: w, Height: h}, nil
}

func Watch() (<-chan Window, <-chan error) {
	r := make(chan Window)
	errch := make(chan error)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGWINCH)
	go func() {
		for {
			<-c
			w, err := Get()
			if err != nil {
				errch <- err
			}
			r <- w
		}
	}()
	return r, errch
}
