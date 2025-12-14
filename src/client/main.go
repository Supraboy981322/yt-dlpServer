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
	url string         //holds url input to yt-dlp 
	verbose bool       //used to check if verbose enabled
	output string      //holds output file path (or stdout/stderr)
	server string      //holds server address 
	conf gomn.Map      //holds config 
	format string      //holds output format
	confPath string    //holds path of config
	args = os.Args[1:] //cmd args
	extraArgs []string //holds extra args passed to yt-dlp
	projName = "yt-dls"//used for config path
)

func init() {
	var ok bool   //avoids weird
	var err error //golang quirk

	//check if verbose enabled
	for _, a := range []string{"--verbose", "-v"} {
		if slices.Contains(args, a) { verbose = true }
	};vLog("verbose output enabled") //only prints if verbose

	//get home dir for config path 
	vLog("getting home dir")
	homeDir, err := os.UserHomeDir()
	if err != nil { erorF("failed to get home dir", err) }
	vLog("got home dir")

	//build config path
	{ vLog("building config path")
		confDir := homeDir //start with home dir
		p := []string{".config", "Supraboy981322", projName}
		for _, d := range p { //add dir to path
			confDir = filepath.Join(confDir, d)
		}//set config path to dir plus config filename
		confPath = filepath.Join(confDir, "config.gomn")
	}; vLog("built config path")

	//make sure the config exists
	{ vLog("ensuring config exists")
		if err := ensureConf(); err != nil {
			vLog("encountered err...")
			erorF("failed to ensure config exists", err)
		} else { vLog("no err") }
	}

	//read the config
	vLog("reading config")
	conf, err = gomn.ParseFile(confPath)
	if err != nil { erorF("failed to read config", err) }
	vLog("read config")
	
	vLog("parsing config")

	//set the server address 
	vLog("getting server address from config")
	if server, ok = conf["server"].(string); !ok {
		vLog("hit err...")
		err = errors.New("not a string")
		erorF("invalid server address", err)
	};vLog("got server address from config")

	vLog("parsed config")

	//parse args
	var taken []int //used to track if arg aready parsed
	for i, a := range args {
		//skips arg if already used
		isTak := slices.Contains(taken, i)
		if isTak { continue } //skips arg

		switch a[1:] {
		 case "f", "-format", "-fmt": //output file format
			format = args[i+1]
			taken = append(taken, i+1)
	   case "u", "l", "-link", "-url": //url passed to yt-dlp
			url = args[i+1]
			taken = append(taken, i+1)
		 case "o", "-output": //output filename
		  output = args[i+1]
			taken = append(taken, i+1)
		 case "a", "-extra-args", "-args": //extra yt-dlp args
		  extraArgs = args[i:]
			for j := i; j < len(args); j++	{
				taken = append(taken, j)
			}
		 case "s", "-server": //for over-riding server address
			server = args[i+1]
			taken = append(taken, i+1)
		 case "v", "-verbose": verbose = true //verbose log level
		 default: //invalid arg
			fmt.Fprintf(os.Stderr, "invalid arg:  %s", a)
		}
	}
}

func main() {
	vLog("server: "+server)
	erorF("todo", errors.New("finish the client"))
}
