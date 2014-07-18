package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
	"github.com/pteichman/go.cobe"
)

var (
	brain *cobe.Cobe2Brain
	clog  = logging.MustGetLogger("cobe.tool")
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
	app := cli.NewApp()
	app.Name = "cobe-tool"
	app.Usage = "cobe administration tool"

	app.Before = func(c *cli.Context) error {
		var err error
		brain, err = cobe.OpenCobe2Brain(c.String("brain"))
		if err != nil {
			clog.Fatal(err)
		}
		return nil
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "brain, b", Value: "cobe.brain", Usage: "name of brain to use"},
	}

	app.Commands = []cli.Command{
		{
			Name:  "learn",
			Usage: "trains the cobe brain from given file/s",
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					clog.Fatal("Usage: learn <file> [<file>]")
				}
				for _, f := range c.Args() {
					learnFileLines(brain, f)
				}
			},
		},
		{
			Name:  "set-stemmer",
			Usage: "sets stemmer language (i.e. 'english') and updates stems from tokens",
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					clog.Fatal("Usage: set-stemmer <language>")
				}
				if err := brain.SetStemmer(c.Args()[0]); err != nil {
					clog.Fatal(err)
				}
			},
		},
		{
			Name:  "del-stemmer",
			Usage: "removes stemmer and all generated stems",
			Action: func(c *cli.Context) {
				if err := brain.DelStemmer(); err != nil {
					clog.Fatal(err)
				}
			},
		},
	}

	app.Run(os.Args)
}
