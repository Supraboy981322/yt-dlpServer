package main

import (
	"os"
	"fmt"
	"errors"
	"strings"
)

func eror(str string, err error) {
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

func ensureConf() error {
	_, err := os.Stat(confPath)
	if errors.Is(err, os.ErrNotExist) {
		vLog("config doesn't exist; writing default")
		err := os.WriteFile(confPath, defConf(), 0666)
		if err != nil { return nil} else { vLog("no err") }
	} else if err != nil { return nil }

	vLog("ensured config")
	return nil
}

func defConf() []byte {
	c := `["server"] := "`
	var resp string
	fmt.Print("please enter your server address:  ")
	fmt.Scan(&resp)
	fmt.Printf("\n")
	vLog("setting server to"+resp)
	c = c+resp+"\"\n"
	vLog("config:")
	for _, l := range strings.Split(c, "\n") {
		vLog("\t"+l+"\033[F")
	}

	return []byte(c)
}
