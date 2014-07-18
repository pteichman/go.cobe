//
// Implements a minimalistic REST HTTP server.
//
// When started with learning enabled will learn from input text
// otherwise will be in read only mode.
//
// Currently the only endpoint avaiable is '/reply'. To get a reply
// POST plaintext to this endpoint. Replies are just plaintext, too.
//
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
	"github.com/pteichman/go.cobe"
)

var (
	brain      *cobe.Cobe2Brain
	clog            = logging.MustGetLogger("cobe.server")
	isLearning bool = false
)

// Starts http server and listens on given address.
func runForever(addr string) {
	clog.Debug("Starting cobe REST HTTP server listening on %s.", addr)

	http.HandleFunc("/reply", replyHandler)
	http.HandleFunc("/", notFoundHandler)

	http.ListenAndServe(addr, nil)
}

// Reply handler. POST questions in plaintext to this endpoint
// and get an answer in plaintext back.
func replyHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	in := strings.TrimSpace(string(body))
	out := brain.Reply(in)

	fmt.Fprint(w, out)

	if isLearning {
		brain.Learn(in)
	}
}

// Handles 404s. Does not return any text in body as this
// can be mistaken to be a valid answer.
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func main() {
	app := cli.NewApp()
	app.Name = "cobe-server"
	app.Usage = "minimal REST HTTP server"
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "learn, L", Usage: "enables learning from input"},
		cli.StringFlag{Name: "brain, b", Value: "cobe.brain", Usage: "name of brain to use"},
	}
	app.Action = func(c *cli.Context) {
		if len(c.Args()) < 1 {
			clog.Fatal("Usage: server <host:[port]>")
		}
		var err error
		if brain, err = cobe.OpenCobe2Brain(c.String("brain")); err != nil {
			clog.Fatal("%s", err)
		}
		isLearning = c.Bool("learn")

		runForever(c.Args()[0])
	}
	app.Run(os.Args)
}
