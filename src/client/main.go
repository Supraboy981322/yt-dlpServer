package main

import (
	"os"
	"fmt"
	"errors"
	"slices"
//	"net/http"
	"path/filepath"
	"github.com/Supraboy981322/gomn"
)

var (
	url string
	verbose bool
	output string
	server string
	conf gomn.Map
	format string
	confPath string
	args = os.Args[1:]
	extraArgs []string
	projName = "yt-dls"
)

func init() {
	var ok bool   //avoids weird
	var err error //golang quirk

	for _, a := range []string{"--verbose", "-v"} {
		if slices.Contains(args, a) { verbose = true }
	};vLog("verbose output enabled")

	vLog("getting home dir")
	homeDir, err := os.UserHomeDir()
	if err != nil { erorF("failed to get home dir", err) }
	vLog("got home dir")

	{ vLog("building config path")
		confDir := homeDir
		p := []string{".config", "Supraboy981322", projName}
		for _, d := range p {
			confDir = filepath.Join(confDir, d)
		}
		confPath = filepath.Join(confDir, "config.gomn")
	}; vLog("built config path")

	{ vLog("ensuring config exists")
		if err := ensureConf(); err != nil {
			vLog("encountered err...")
			erorF("failed to ensure config exists", err)
		} else { vLog("no err") }
	}

	vLog("reading config")
	conf, err = gomn.ParseFile(confPath)
	if err != nil { erorF("failed to read config", err) }
	vLog("read config")
	
	vLog("parsing config")

	vLog("getting server address from config")
	if server, ok = conf["server"].(string); !ok {
		vLog("hit err...")
		err = errors.New("not a string")
		erorF("invalid server address", err)
	};vLog("got server address from config")

	vLog("parsed config")

	var taken []int
	for i, a := range args {
		isTak := slices.Contains(taken, i)
		if isTak { continue } //skips arg

		switch a[1:] {
		 case "f", "-format", "-fmt":
			format = args[i+1]
			taken = append(taken, i+1)
	   case "u", "l", "-link", "-url":
			url = args[i+1]
			taken = append(taken, i+1)
		 case "o", "-output":
		  output = args[i+1]
			taken = append(taken, i+1)
		 case "a", "-extra-args", "-args":
		  extraArgs = args[i:]
			for j := i; j < len(args); j++	{
				taken = append(taken, j)
			}
		 case "s", "-server":
			server = args[i+1]
			taken = append(taken, i+1)
		 case "v", "-verbose": verbose = true
		 default:
			fmt.Fprintf(os.Stderr, "invalid arg:  %s", a)
		}
	}
}

func main() {
	vLog("server: "+server)
	erorF("todo", errors.New("finish the client"))
}
