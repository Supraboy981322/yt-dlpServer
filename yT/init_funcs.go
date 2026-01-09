package main

import(
	"os"
	"fmt"
	"errors"
	"slices"
	"path/filepath"
	"github.com/Supraboy981322/gomn"
)

func readConf() {
	var ok bool   //avoids weird
	var err error //golang quirk

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

	vLog("checking \"use video name\" bool in config")
	if useVideoName, ok = conf["use video name"].(bool); !ok {
		vLog("\"use video name\" in config is either not bool or not present, ignoring")
	} else { vLog("using video name as suggested file name") }

	//global extra args
	if extraArgsR, ok := conf["yt-dlp args"].([]interface{}); ok {
		for _, aR := range extraArgsR {
			if a, ok := aR.(string); !ok {
				err := errors.New("not a string")
				erorF("invalid config (\"yt-dlp args\")", err)
			} else {
				extraArgs = append(extraArgs, a)
				fmt.Println(a)
			}
		};vLog("using extra args from config")
	} else { vLog("no extra args provided in config") }

	vLog("parsed config")
}

func parseArgs() {
	//parse args (spagetti, I know)
	var taken []int //used to track if arg aready parsed
	for i, a := range args {
		//skips arg if already used
		isTak := slices.Contains(taken, i)
		if isTak { continue } //skips arg

		switch a[1:] {
		 case "L", "-playlist": mode = 1
		 case "F", "-file": mode = 2
		 case "f", "-format", "-fmt": //output file format
			if len(args)-1 >= i+1 {
				format = args[i+1]
				taken = append(taken, i+1)
			} else { invArg("have format arg but no value") }
	   case "u", "l", "-link", "-url", "-video": //url passed to yt-dlp
			if len(args)-1 >= i+1 {
				url = args[i+1]
				taken = append(taken, i+1)
			} else { invArg("have url arg but no value") } 
		 case "o", "-output": //output filename
			if len(args)-1 >= i+1 {
		  	output = args[i+1]
				taken = append(taken, i+1)
			} else { invArg("have output arg but no value") } 
		 case "a", "-extra-args", "-args": //extra yt-dlp args
		  extraArgs = append(extraArgs, args[i:]...)
			for j := i; j < len(args); j++	{
				taken = append(taken, j)
			}
		 case "s", "-server": //for over-riding server address
			if len(args)-1 >= i+1 {
				server = args[i+1]
				taken = append(taken, i+1)
			} else { invArg("have server arg but no value") } 
		 case "x", "-audio-only": extraArgs = append(extraArgs, "-x")
		 case "v", "-verbose": verbose = true //verbose log level
		 case "h", "-help": help()
		 case "q", "-qual", "-quality":
			if len(args)-1 >= i+1 {
				quality = args[i+1]
				taken = append(taken, i+1)
			} else { invArg("have quality arg, but no value") }
		 case "V", "-use-video-name":
		 	useVideoName = true
		 default:
			//check if it's the url
			if url == "" { url = a
			} else {
				err := fmt.Errorf("\"%s\" used as url, but url is already set to \"%s\"", a, url)
				invArg(err.Error())
			}
		}
	}
}
