package main

import (
	"os"
	"fmt"
	"errors"
	"strings"
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

