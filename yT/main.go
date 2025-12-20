package main

import (
	"os"
	"io"
	"fmt"
	"time"
	"errors"
	"slices"
	"strings"
	"net/http"
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
	quality string     //holds output quality
	confPath string    //holds path of config
	args = os.Args[1:] //cmd args
	extraArgs []string //holds extra args passed to yt-dlp
	projName = "yt-dlpServer"//used for config path
)

type (
	//progress writer
	progWr struct {
		Down uint64
	}
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

	//parse args (spagetti, I know)
	var taken []int //used to track if arg aready parsed
	for i, a := range args {
		//skips arg if already used
		isTak := slices.Contains(taken, i)
		if isTak { continue } //skips arg

		switch a[1:] {
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
		 default:
			//check if it's the url
			if url == "" { url = a
			} else {
				err := fmt.Errorf("%s used as url, but url is already set to %s", a, url)
				invArg(err.Error())
			}
		}
	}

	if url == "" { invArg("need url")	}
}

func main() {
	//print server domain being used
	fmt.Printf("using server:  %s\n", server)

	//construct request
	req, err := http.NewRequest("GET", server, nil)
	if err != nil { erorF("failed to create request", err) }

	//handle audio-only output 
	if format == "" && slices.Contains(extraArgs, "-x") && quality == "" {
		format = "webm"
		quality = "bestaudio"
	}

	//map of header values and their keys 
	argsMap := map[string]string {
		"fmt": format,
		"qual": quality,
		"url": url,
		"args": strings.Join(extraArgs, ";"),
	}

	//range over said map of headers
	for header, val := range argsMap {
		req.Header.Set(header, val)
		if val != "" { //only print if set
			fmt.Printf("%s:  %s\n", header, val)
		}
	};fmt.Printf("\n") //start newline for activity spinner
	
	//channel to send quit msg
	quitProg := make(chan bool)
	go func(){ //activity spinner
		progIcn := []rune{'⠻','⠽','⠾','⠷','⠯','⠟',}
		for i := 0;; i++ {
			//reset index to 0
			if i >= len(progIcn) { i = 0 }
			select { //handle channel comms.
       case <- quitProg:
				//move cursor up one line and
				//  clear it before returning
				fmt.Printf("\033[A\033[2K\033[0m")
				return
	     default:
				//ansii code to manipulate cursor and use color
				fmt.Printf("\033[A\033[2K\033[1;34m %s\033[0;1m "+
							"Making request...\033[0m\n", string(progIcn[i]))
				//wait 100 milliseconds (looks nicer)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	//actually do request
	client := &http.Client{} //create client
	resp, err := client.Do(req) //make request
	if err != nil { erorF("\033[Aerr making request", err) }
	defer resp.Body.Close() //keep body open

	//check for server err
	if resp.StatusCode != http.StatusOK {
		//assume server only sent err 
		//  in body (as opposed to binary)
		bod, err := io.ReadAll(resp.Body)
		if err != nil { eror("\033[Aerr reading response body", err) }

		//if no err was sent, use status code
		if bod == nil {
			bod = []byte(fmt.Sprintf("(%d): %s", resp.StatusCode, resp.Status))
		}

		//print err
		err = errors.New(string(bod))
		erorF("\033[Aserver reported bad status code", err)
	} else { vLog("response status: "+resp.Status) }

	if output == "--" { /* TODO: print binary to stdout */
		erorF("\033[ATODO:", errors.New("output to stdout"))
	} else if output == "" { //if no output arg
		//get the suggested file name
		oR := resp.Header.Get("Content-Disposition")
		oR = strings.Split(oR, ";")[1]
		oR = strings.Split(oR, "=")[1]
		//set output to suggested filename
		output = oR[1:len(oR)-1]
	}

	//stop activity spinner
	quitProg <- true

	//print output filename
	fmt.Printf("\033[2K\033[1moutputing to:  \033[0;35m%s\033[0m\n\n", output)

	//create progress writer
	pw := &progWr{}

	//create output file
	out, err := os.Create(output)
	if err != nil { erorF("failed to create output file", err) }
	defer out.Close() //hold output file open

	reader := io.TeeReader(resp.Body, pw)

	_, err = io.Copy(out, reader)
	if err != nil { erorF("err streaming to output", err) }

	//print in a human-readable format (i.e. not just bytes)
	fmt.Printf("\n\033[1;32mdone. total: \033[1;34m%s\033[0m\n",
					bytesToHumanReadable(pw.Down))
}

func (pw *progWr) Write(p []byte) (n int, err error) {
	n = len(p) //number of bytes
	pw.Down += uint64(n) //add to amnt downloaded
	dlSize := bytesToHumanReadable(pw.Down) //convert to a human-readable format

	//print it in a way that looks appealing
	fmt.Printf("\033[A;2K\033[1m\rdownloaded \033[1;34m%s\033[0;1m ....\033[0m\n", dlSize)
	return n, nil //return num bytes
}

func bytesToHumanReadable(b uint64) string {
	//if stupidly large, don't even process it
	if b >= 9223372036854775807 {
		return fmt.Sprintf("greater than 9.22 EB")
	}

	//find best measurement
	s := float64(b)
	for _, f := range []string{"B", "KB", "MB", "GB", "TB", "PB"} {
		if s < 1000.0 { return fmt.Sprintf("%.2f %s", s, f)
		} else { s = s / 1000.0 }
	}

	//otherwise, it's in exabytes
	return fmt.Sprintf("%.2f EB", s)
}
