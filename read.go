package main

import (
	"fmt"
	"io"
)

func read(inStream InStream, outStream OutStream) (string, error) {
	buf := make([]rune, 0, 8)
loop:
	for {
		r, _, err := inStream.ReadRune()
		if err == io.EOF {
			continue
		}
		if err != nil {
			return "", err
		}

		switch {
		case r == '\r':
			fmt.Fprintf(outStream, "\r\n")
			break loop
		case r == CTRL_H:
			if len(buf) > 0 {
				fmt.Fprintf(outStream, "\033[1D")
				fmt.Fprintf(outStream, "\033[0K")
				buf = buf[:len(buf)-1]
			}
		case CTRL_A <= r && r <= CTRL_Z:
			// TODO
		default:
			fmt.Fprintf(outStream, "%c", r)
			buf = append(buf, r)
		}
	}

	return string(buf), nil
}
