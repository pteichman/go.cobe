package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
)

import (
	"github.com/op/go-logging"
	"github.com/pteichman/go.cobe"
	"github.com/pteichman/go.cobe/console"
	"github.com/pteichman/go.cobe/ircbot"
)

var clog = logging.MustGetLogger("cobe.main")

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

var (
	ircserver  = flag.String("irc.server", "", "irc server (host:port)")
	ircchannel = flag.String("irc.channels", "#cobe", "irc channels")
	ircnick    = flag.String("irc.nick", "cobe", "irc nickname")
)

func learnFileLines(b *cobe.Brain, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	s := bufio.NewScanner(bufio.NewReader(f))
	for s.Scan() {
		fmt.Println(s.Text())
		b.Learn(s.Text())
	}

	return nil
}

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			clog.Fatal("%s", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: cobe console")
		os.Exit(1)
	}

	b, err := cobe.OpenBrain("cobe.brain")
	if err != nil {
		clog.Fatal("%s", err)
	}

	switch cmd := args[0]; cmd {
	case "console":
		console.RunForever(b)
	case "irc-client":
		opts := &ircbot.Options{
			*ircserver, *ircnick, []string{*ircchannel}, nil,
		}
		ircbot.RunForever(b, opts)
	case "learn":
		for _, f := range args[1:] {
			learnFileLines(b, f)
		}
	default:
		clog.Fatalf("Unknown command: %s", cmd)
	}
}
