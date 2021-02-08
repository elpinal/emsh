package stdin

import (
	"bufio"
	"io"
	"os"
	"syscall"

	"github.com/pkg/term/termios"
	"golang.org/x/sys/unix"
)

type Stdin struct {
	*bufio.Reader
}

func New() Stdin {
	return Stdin{Reader: bufio.NewReader(os.Stdin)}
}

func (_ Stdin) GetInner() io.Reader {
	return os.Stdin
}

func setState(state *unix.Termios) error {
	return termios.Tcsetattr(uintptr(syscall.Stdin), termios.TCSANOW, state)
}

func (in Stdin) EnterRaw() (func() error, error) {
	var state unix.Termios
	err := termios.Tcgetattr(uintptr(syscall.Stdin), &state)
	if err != nil {
		return nil, err
	}
	prevState := state

	termios.Cfmakeraw(&state)
	state.Cc[unix.VMIN] = 0
	state.Cc[unix.VTIME] = 5 // 500 milliseconds.

	err = setState(&state)
	if err != nil {
		return nil, err
	}

	return func() error { return setState(&prevState) }, nil
}
