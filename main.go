package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"time"

	"github.com/elpinal/emsh/command"
	"github.com/elpinal/emsh/position"
	"github.com/elpinal/emsh/stdin"

	"github.com/jackc/pgx/v4"
)

type InStream interface {
	EnterRaw() (func() error, error)
	GetInner() io.Reader
	io.RuneReader
}

type OutStream interface {
	io.Writer
}

type ErrStream interface {
	io.Writer
}

func main() {
	err := run(stdin.New(), os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

const database_url = "postgresql://localhost/emsh_history"

const schema = `
create table if not exists history (
    time timestamp,
    line text
)`

func run(inStream InStream, outStream OutStream, errStream ErrStream) error {
	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	// defer conn.Close(context.Background())

	_, err = conn.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	for {
		s, err := askCmd(inStream, outStream)
		if err != nil {
			fmt.Fprintf(errStream, "error: %v\n", err)
			continue
		}

		err = saveHistory(conn, s)
		if err != nil {
			fmt.Fprintf(errStream, "error: %v\n", err)
			continue
		}

		cmd := parse(inStream, outStream, s)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		err = cmd.Run()
		signal.Stop(c)
		if err != nil {
			fmt.Fprintf(errStream, "error: %v\n", err)
			continue
		}
	}

	// r, _ := winsize.Watch()
	// for {
	// 	w := <-r
	// 	fmt.Println(w)
	// }
}

func askCmd(inStream InStream, outStream OutStream) (string, error) {
	restore, err := inStream.EnterRaw()
	if err != nil {
		return "", err
	}
	defer func() {
		err = restore()
		if err != nil {
			panic(err)
		}
	}()

	p, err := getCursorPos(inStream, outStream)
	if err != nil {
		return "", err
	}
	// fmt.Fprintf(outStream, "cursor position: %v\r\n", p)

	if !p.Leftmost() {
		outStream.Write([]byte("\033[7m")) // Swap foreground and background colors.
		outStream.Write([]byte("%"))
		outStream.Write([]byte("\033[0m")) // Reset colors.
		outStream.Write([]byte("\r\n"))
	}

	fmt.Fprintf(outStream, "Σ ")

	return read(inStream, outStream)
}

func getCursorPos(inStream InStream, outStream OutStream) (*position.Point, error) {
	// CSI sequence: ask the cursor position.
	// The answer is of the form `ESC[m;nR` where `m` and `n` are natural numbers.
	fmt.Fprintf(outStream, "\033[6n")

	for {
		if r, _, err := inStream.ReadRune(); err != nil {
			return nil, err
		} else if r == '\033' {
			break
		}
	}
	if _, _, err := inStream.ReadRune(); err != nil {
		return nil, err
	}

	var buf []rune
	for { // Read until ';'.
		r, _, err := inStream.ReadRune()
		if err != nil {
			return nil, err
		}
		if r == ';' {
			break
		}
		buf = append(buf, r)
	}
	line, err := strconv.Atoi(string(buf))
	if err != nil {
		return nil, err
	}

	buf = buf[:0]
	for { // Read until 'R'.
		r, _, err := inStream.ReadRune()
		if err != nil {
			return nil, err
		}
		if r == 'R' {
			break
		}
		buf = append(buf, r)
	}
	column, err := strconv.Atoi(string(buf))
	if err != nil {
		return nil, err
	}

	return position.NewPoint(uint(line), uint(column)), nil
}

func saveHistory(conn *pgx.Conn, s string) error {
	_, err := conn.Exec(
		context.Background(),
		"INSERT INTO history VALUES ($1, $2)",
		time.Now(),
		s,
	)
	return err
}

func parse(inStream InStream, outStream OutStream, s string) command.Cmd {
	switch s {
	case "exit":
		return command.Exit{Code: 0}
	default:
		cmd := exec.Command(s)
		cmd.Stdout = outStream
		cmd.Stdin = inStream.GetInner()
		return cmd
	}
}
