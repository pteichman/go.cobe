package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
)

import (
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/pteichman/go.cobe/console"
	"github.com/pteichman/go.cobe/ircbot"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

var (
	ircserver  = flag.String("irc.server", "", "irc server (host:port)")
	ircchannel = flag.String("irc.channels", "#cobe", "irc channels")
	ircnick    = flag.String("irc.nick", "cobe", "irc nickname")
)

var (
	statsdserver = flag.String("statsd.server", "", "statsd server (host:port)")
	statsdname   = flag.String("statsd.name", "cobe", "statsd name")
)

func learnFileLines(b *cobe.Cobe2Brain, path string) error {
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
			log.Fatalf("Creating cpu profile: %s", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *statsdserver != "" {
		s, err := statsd.New(*statsdserver, *statsdname)
		if err != nil {
			log.Fatalf("Initializing statsd: %s", err)
		}

		cobe.SetStatter(s)
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: cobe console")
		os.Exit(1)
	}

	b, err := cobe.OpenCobe2Brain("cobe.brain")
	if err != nil {
		log.Fatalf("Opening brain file: %s", err)
	}

	var cmd = args[0]
	switch {
	case cmd == "console":
		console.RunForever(b)
	case cmd == "ircbot" || cmd == "irc-client":
		opts := &ircbot.Options{
			Server:   *ircserver,
			Nick:     *ircnick,
			Channels: []string{*ircchannel},
		}
		ircbot.RunForever(b, opts)
	case cmd == "learn":
		for _, f := range args[1:] {
			learnFileLines(b, f)
		}
	case cmd == "del-stemmer":
		err := b.DelStemmer()
		if err != nil {
			clog.Fatal(err)
		}
	case cmd == "set-stemmer":
		if len(args) < 2 {
			clog.Fatal("Usage: set-stemmer <language>")
		}
		err := b.SetStemmer(args[1])
		if err != nil {
			clog.Fatal(err)
		}
	default:
		clog.Fatalf("Unknown command: %s", cmd)
	}
}
