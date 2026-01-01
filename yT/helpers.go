package main

import (
	"io"
	"os"
	"fmt"
	"time"
	"errors"
	"strings"
	"net/http"
	"path/filepath"
)

func eror(str string, err error) {
	fmt.Printf("\033[2K")
	str = fmt.Sprintf("\033[1;30;41m%s\033[0m", str)
	err = fmt.Errorf("    \033[1;31m%v\033[0m", err)
	fmt.Fprintf(os.Stderr, "%s\n%v\n", str, err)
}

func erorF(str string, err error) {
	eror(str, err)
	os.Exit(1)
}

func vLog(str string) {
	if verbose {
		fmt.Println(str)
	}
}

func invArg(str string) {
	vLog("found err...")
	err := errors.New(str)
	erorF("invalid arg", err)
}

//func to make sure config is ok
func ensureConf() error {
	//check file
	_, err := os.Stat(confPath)
	//check if it exists 
	if errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(filepath.Dir(confPath), 0766)
		//generate it if not
		vLog("config doesn't exist; writing default")
		err := os.WriteFile(confPath, defConf(), 0666)
		if err != nil { return err } else { vLog("no err") }
    fmt.Println("setup complete")
    fmt.Print("\n\trun the command again to use it\n")
    fmt.Println("\tfor usage, see: '\033[33myT\033[36m --help\033[0m'\n")
    os.Exit(0)
	} else if err != nil { return err }

	//return no err 
	vLog("ensured config")
	return nil
}

//func that builds config 
func defConf() []byte {
	fmt.Println("config not found")
	fmt.Printf("  creating config ")
	fmt.Println("(you will prompted for settings)")
	//first line of config
	c := `["server"] := "`
	
	//var to hold user input
	var input string

	//prompt user for server address
	fmt.Print("\033[1;33mplease enter your server address:  \033[0m")
	fmt.Scan(&input) //read terminal input
	fmt.Printf("\n") //print newline

	vLog("setting server to"+input)

	//add to config
	c = c+input+"\"\n"

	//print config if verbose 
	vLog("config:")
	for _, l := range strings.Split(c, "\n") {
		vLog("\t"+l+"\033[F")
	}

	fmt.Println("created config\n")

	//return config as byte slice
	return []byte(c)
}
	
func help() {
	lines := []string{
		"yt-dlp server (working name) --> help",
		"  -h, --help",
		"    this screen",
		"  -x, --audio-only",
		"    only download audio",
		"  -v, --verbose",
		"    show verbose output",
		"  -f, --format, --fmt",
		"    file format",
		"  -o, --output",
		"    output file name (extension is appended to end, so don't put it here)",
		"  -s, --server",
		"    over-ride server address",
		"  -a, --args, --extra-args",
		"    additional args passed to yt-dlp on the server",
		"  -V, --use-video-title",
		"    request server to suggest the video's title as the filename",
		"  -u, -l, --url, --link, --video",
		"    video url",
	}
	for _, l := range lines {
		fmt.Println(l)
	}
	os.Exit(0)
}

func dlFromPlaylist(url string) {
	fmt.Println("")
	//construct request
	req, err := http.NewRequest("GET", server, nil)
	if err != nil { erorF("failed to create request", err) }
	req.Header.Set("list-playlist", "true")
	req.Header.Set("url", url)

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

	//assume server only sent list 
	//  in body (as opposed to binary)
	bod, err := io.ReadAll(resp.Body)
	if err != nil { eror("\033[Aerr reading response body", err) }

	if resp.StatusCode != http.StatusOK {
		//if no err was sent, use status code
		if bod == nil {
			bod = []byte(fmt.Sprintf("(%d): %s", resp.StatusCode, resp.Status))
		}
		//print err
		err = errors.New(string(bod))
		erorF("\033[Aserver reported bad status code", err)
	}

	quitProg<-true
	listR := string(bod)
	list := strings.Split(listR, "\n")
	if list[len(list)-1] == "" { list = list[:len(list)-1] }
	for i, v := range list {
		fmt.Printf("\n\n\033[1;35mstarting\033[0m \033[38;2;255;165;0m%d\033[0m"+
			" \033[1;35mof\033[0m \033[38;2;155;165;0m%d\033[0m\n", i+1, len(list))
		url = v ;	dl()
	}
}
