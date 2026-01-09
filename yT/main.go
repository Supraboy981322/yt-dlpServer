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
	"github.com/Supraboy981322/gomn"
)

var (
	mode int           //determines which mode to use
	url string         //holds url input to yt-dlp 
	verbose bool       //used to check if verbose enabled
	output string      //holds output file path (or stdout/stderr)
	server string      //holds server address 
	conf gomn.Map      //holds config 
	format string      //holds output format
	quality string     //holds output quality
	confPath string    //holds path of config
	useVideoName bool  //used to request video name as suggested filename
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
	//check if verbose enabled
	for _, a := range []string{"--verbose", "-v"} {
		if slices.Contains(args, a) { verbose = true }
	};vLog("verbose output enabled") //only prints if verbose

	readConf()
	parseArgs()

	if len(extraArgs) > 1 { extraArgs = extraArgs[1:] }
	if url == "" { invArg("need url")	}

	if mode != 1 {
		params := strings.Split(url, "?")
		if len(params) > 2 {
			params = strings.Split(params[1], "&")
			fmt.Printf("%#v\n", params)
			os.Exit(1)
			for _, p :=  range params {
				n := strings.Split(p, "=")
				if len(n) < 2 { continue }
				if strings.Contains(n[0], "list") {	mode = 1 ; break }
			}
		}
	}
}

func main() {
	switch mode {
	 case 0: dl()
	 case 1: dlFromPlaylist(url)
	 case 2: dlFromFile(url)
	}
}

func dl() {
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
		"use-video-name": "",
		"args": strings.Join(extraArgs, ";"),
	};if useVideoName { argsMap["use-video-name"] = "true" }

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
