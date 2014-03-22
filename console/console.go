package console

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/GeertJohan/go.linenoise"
	"github.com/pteichman/go.cobe"
)

func tildeExpand(filename string) (string, error) {
	if filename[0:1] != "~" {
		return filename, nil
	}

	sep := strings.IndexRune(filename, os.PathSeparator)
	if sep < 0 {
		err := errors.New("couldn't find path separator after tilde")
		return filename, err
	}

	var u *user.User
	var err error

	username := filename[1:sep]
	if username != "" {
		u, err = user.Lookup(username)
	} else {
		u, err = user.Current()
	}

	if err != nil {
		return filename, nil
	}

	return strings.Join([]string{u.HomeDir, filename[sep:]}, ""), nil
}

func loadHistory() string {
	history, err := tildeExpand("~/.cobe_history")
	if err != nil {
		fmt.Printf("Disabling history: %s\n", err)
		return ""
	}

	err = linenoise.LoadHistory(history)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return history
}

func saveHistory(filename string) error {
	err := linenoise.SaveHistory(filename)
	if err != nil {
		fmt.Println(err)
	}

	return err
}

func RunForever(b *cobe.Cobe2Brain) {
	history := loadHistory()

	for {
		err := RunOne(b)
		if err != nil {
			break
		}
	}

	if history != "" {
		saveHistory(history)
	}
}

func RunOne(b *cobe.Cobe2Brain) error {
	line, err := linenoise.Line("> ")
	if err != nil {
		return err
	}

	if line != "" {
		linenoise.AddHistory(line)
		b.Learn(line)
	}

	fmt.Println(b.Reply(line))

	return nil
}
