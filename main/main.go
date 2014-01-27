package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
)

import (
	"github.com/pteichman/go.cobe"
	"github.com/pteichman/go.cobe/console"
	"github.com/pteichman/go.cobe/ircbot"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

var (
	ircserver  = flag.String("irc.server", "", "irc server (host:port)")
	ircchannel = flag.String("irc.channels", "#cobe", "irc channels")
	ircnick    = flag.String("irc.nick", "cobe", "irc nickname")
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	args := flag.Args()
	if len(args) < 1 {
		log.Fatalln("Usage: cobe console")
	}

	b, err := cobe.OpenBrain("cobe.brain")
	if err != nil {
		log.Fatal(err)
	}

	if args[0] == "console" {
		console.RunForever(b)
	} else if args[0] == "irc-client" {
		opts := &ircbot.Options{
			*ircserver, *ircnick, []string{*ircchannel}, nil,
		}
		ircbot.RunForever(b, opts)
	}
}
