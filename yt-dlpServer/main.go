package main

import (
	"os"
	"io"
	"fmt"
	"time"
	"bytes"
	"os/exec"
	"strings"
	"net/http"
	"github.com/charmbracelet/log"
	"github.com/Supraboy981322/gomn"
	elh "github.com/Supraboy981322/ELH"
)

var (
	port int
	conf gomn.Map
	useWebUI bool
	srvErr = http.StatusInternalServerError
)

func init() {
	var ok bool
	var err error

	log.Info("initializing...") 

	conf, err = gomn.ParseFile("config.gomn")
	if err != nil {
		log.Fatalf("err parsing config:  %v", err)
	} else { log.Debug("read config file") }
	
	if logLvl, ok := conf["log level"].(string); ok {
		switch strings.ToLower(logLvl) {
		 case "debug": log.SetLevel(log.DebugLevel)
		 case "info":  log.SetLevel(log.InfoLevel)
		 case "warn":  log.SetLevel(log.WarnLevel)
		 case "error": log.SetLevel(log.ErrorLevel)
		 case "fatal": log.SetLevel(log.FatalLevel)
		 default:
			log.Fatal("invalid value for \"log level")
		}; log.Info("set log level")
	} else {
		log.Error("value of \"log level\" is not a string")
		log.Warn("defaulting to \"debug\" log level")
	}

	if port, ok = conf["port"].(int); !ok {
		log.Fatal("port is not an integer")
	} else { log.Debug("set port") }

	if useWebUI, ok = conf["use web ui"].(bool); !ok {
		log.Error("value of \"use web ui\" is not a bool")
		log.Warn("defaulting to \033[33mfalse\033[0m")
	} else { log.Debug("set web ui bool") }
	if useWebUI { log.Debug("web ui enabled")
	} else { log.Warn("web ui diabled") }

	log.Info("initialized")
}

func main() {
	log.Debug("starting web server")
	http.HandleFunc("/", webHan)
	
	log.Infof("listening on port %d", port)

	portStr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(portStr, nil))

	log.Error("uhhh... this line shouldn't've run.")
}

func webHan(w http.ResponseWriter, r *http.Request) {
	var resp string //response sent by server
	path := r.URL.Path[1:]
	switch strings.ToLower(path) {
	 case "yt-dlp", "y", "dl", "d", "ytdlp":
		 resp = "yt-dlp"
		 ytDlp(w, r)
   default:
		if useWebUI {	resp, _ = elh.Serve(w, r)
		} else { resp = "yt-dlp"; ytDlp(w, r) }
	}

	ip := r.RemoteAddr
	log.Infof("req: %s | resp: %s | ip: %s", path, resp, ip)
}

func ytDlp(w http.ResponseWriter, r *http.Request) {
	//get the url from headers,
	//  with fallback to the req body
	url := chkHeaders([]string{
			"url", "source", "src", "addr",
			"u", "address", "video", "song",
			"v",
		}, getBodyNoErr(r), r)

	//quickly return err if no url 
	if url == "" {
		http.Error(w, "no url provided", http.StatusBadRequest)
		return
	}

	if r.Header.Get("list-playlist") != "" {
		w.Header().Set("Content-Type", "text/plain")
		var out_buf bytes.Buffer
		cmd := exec.Command("yt-dlp", "-qo", "--no-warnings", "--get-id",
						"--flat-playlist", url)
		cmd.Stdout = &out_buf
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			http.Error(w, err.Error(), 500)
			log.Error(err.Error())
			return
		}
		w.Write([]byte(out_buf.String()))
		return
	}
	//let client know it's about to 
	//  recieve raw binary data
	w.Header().Set("Content-Type", "application./octet-stream")

	//get the format from headers,
	//  defaults to webm
	format := chkHeaders([]string{
			"fmt", "format", "f",
		}, "webm", r)

	filename := fmt.Sprintf("yt-dlpServer_%s", 
					time.Now().Format("2006-01-02_15-04-05"))
	if chkHeaders([]string{
		"vName", "vN", "videoName", "use-video-name"}, "", r) != "" {
		
		log.Debug("getting filename")

		cmd := exec.Command("yt-dlp", "--get-filename",
					"-qo", "\"%(title)s\"", "--no-warnings", url)

		var stdout, stderr bytes.Buffer
		cmd.Stderr = &stderr
		cmd.Stdout = &stdout

		if err := cmd.Start(); err != nil { 
			log.Error(stderr.String())
			http.Error(w, "failed to start yt-dlp", srvErr)
			return
		}
		
		if err := cmd.Wait(); err != nil {
			//err buffer to string 
			errMsg := stderr.String() 
	
			var indx int
			for _, l := range strings.Split(errMsg, "\n") {
				//remove the error type part
				//  of yt-dlp output
				indx = strings.IndexRune(l, ':')
				if indx == -1 { continue }
				errTyp := l[0:indx]
				errMsg = l[indx+1:]
				if errTyp == "ERROR" { break }
			}
	
			//remove newline
			//  (yt-dlp inserts double newline)
			errMsg = strings.ReplaceAll(errMsg, "\n", "")
	
			//send err
			http.Error(w, errMsg, srvErr)
			log.Error(errMsg)
			return
		}

		videoName := stdout.String()
		videoName = videoName[1:len(videoName)-2]
		filename = videoName
	}

	outHeader := fmt.Sprintf(
				"attachment; filename=\"%s.%s\"", filename, format)
	w.Header().Set("Content-Disposition", outHeader)

	//get quality arg from headers
	//  defaults to `bestvideo+bestaudio/best`
	quality := chkHeaders([]string{
			"quality", "qual", "q",
		}, "bestvideo+bestaudio/best", r)

  extraArgsR := chkHeaders([]string{
      "a", "args", "arg",
    }, "", r)

  extraArgs := strings.Split(extraArgsR, ";")

	//quickly return err if no url 
	if url == "" {
		http.Error(w, "no url provided", http.StatusBadRequest)
		return
	}

	//args passed to yt-dlp
	args := []string{
		url,
		"-o", "-",
		"-q",
		"--recode-video", format,
		"-f", quality,
	};args = append(args, extraArgs...)
	
	//yt-dlp cmd
	cmd := exec.Command("yt-dlp", args...)

	//create stdout buffer
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, err.Error(), srvErr)
		return
	}; defer stdout.Close()

	//a multi-buffer output of
	//  cmd stderr
	var clientMsgBuff bytes.Buffer
	errBuff := io.MultiWriter(os.Stderr, &clientMsgBuff)
	cmd.Stderr = errBuff

	//exec cmd
	if err := cmd.Start(); err != nil {
		http.Error(w, err.Error(), srvErr)
		return
	}

	//stream yt-dlp output to client
	if _, err := io.Copy(w, stdout); err != nil {
		http.Error(w, err.Error(), srvErr)
		return
	}

	if err = cmd.Wait(); err != nil {
		//err buffer to string 
		errMsg := clientMsgBuff.String()

		var indx int
		for _, l := range strings.Split(errMsg, "\n") {
			//remove the error type part
			//  of yt-dlp output
			indx = strings.IndexRune(l, ':')
			if indx == -1 { continue }
			errTyp := l[0:indx]
			errMsg = l[indx+1:]
			if errTyp == "ERROR" { break }
		}

		//remove newline
		//  (yt-dlp inserts double newline)
		errMsg = strings.ReplaceAll(errMsg, "\n", "")

		//send err
		http.Error(w, errMsg, srvErr)
		return 
	}
}
