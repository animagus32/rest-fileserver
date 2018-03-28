package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"log"

	"fileserver/service"

	"github.com/gorilla/mux"
)

type Config struct {
	MaxProcs   int    `json:"max_procs"`
	BindHost   string `json:"bind_host"`
	Host       string `json:"host"`
	DebugLevel int    `json:"debug_level"`
	FilePath   string `json:"file_path"`
}

const appName = "file-server"

var conf Config

func init() {
	cfgpath := flag.String("f", "server.conf", "server config")
	// load default config

	f, err := os.Open(*cfgpath)
	if err != nil {
		log.Panic(err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(data, &conf)
	if err != nil {
		log.Panic(err)
	}

	log.Println("config:", conf)

	if _, err := os.Stat(conf.FilePath); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(conf.FilePath, 0777); err != nil {
				log.Panic(err)
			}
		} else {
			log.Panic(err)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(conf.MaxProcs)

	svc, err := service.NewFileserverService(conf.Host, conf.FilePath)
	if err != nil {
		log.Fatal("failed to create file server service instance:", err)
	}

	var router = mux.NewRouter()

	router.HandleFunc("/v1/upload", svc.OptionUpload).Methods("OPTIONS")
	router.HandleFunc("/upload", svc.UploadPage).Methods("GET")
	router.HandleFunc("/v1/upload", svc.PostUpload).Methods("POST")
	router.HandleFunc("/v1/mkdir", svc.PostMkdir).Methods("POST")
	router.HandleFunc("/v1/file/list", svc.GetFileList).Methods("GET")
	router.HandleFunc("/v1/file", svc.DeleteFile).Methods("DELETE")

	router.HandleFunc("/v1/upload/stream", svc.PostUploadStream).Methods("POST")

	router.PathPrefix("/").HandlerFunc(svc.ServeFile)

	log.Printf("Starting %s..., listen on %s", appName, conf.BindHost)
	log.Fatal("http.ListenAndServe:", http.ListenAndServe(conf.BindHost, router))
	log.Printf(appName + " stopped, process exit")
}
