package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"fileserver/api"
	"log"

	"github.com/pborman/uuid"
)

type FileserverService struct {
	path       string
	host       string
	fileserver http.Handler
}

func NewFileserverService(host, path string) (*FileserverService, error) {
	fileserverHandler := http.FileServer(http.Dir(path))

	m := &FileserverService{path: path, fileserver: fileserverHandler, host: host}
	return m, nil
}

func (s *FileserverService) OptionUpload(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许访问所有域
	// w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Connection, User-Agent, Cookie") //header的类型
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Headers", "*") //header的类型

	return
}

func (s *FileserverService) UploadPage(w http.ResponseWriter, req *http.Request) {
	// 上传页面
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	html := `
		<html><head><meta charset="UTF-8"></head>
<form enctype="multipart/form-data" action="/v1/upload" method="POST">
	Send this file: <input name="file" type="file" />
    File Path: <input name="path" value="default" />	
	<input type="submit" value="Send File" />
	
</form>
</html>
`

	io.WriteString(w, html)
	return
}

func (s *FileserverService) PostUpload(w http.ResponseWriter, req *http.Request) {

	req.ParseForm()
	req.ParseMultipartForm(32 << 20)

	f, header, err := req.FormFile("file")
	defer req.Body.Close()

	if err != nil {
		log.Println(err)
		return
	}
	inputpath := req.FormValue("path")
	log.Println("input path:", inputpath)
	filename := header.Filename
	abspath, err := s.genFilepath(filename, inputpath)
	if err != nil {
		log.Println(err)
		Error(w, 100, err.Error())
	}
	log.Println("abspath:", abspath)

	log.Println("output file path:", abspath)
	fout, _ := os.Create(abspath)

	defer fout.Close()
	buf := make([]byte, 1024)
	for {
		_, err := f.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}
		fout.Write(buf)
	}

	fileUri := s.genFileUri(abspath[len(s.path):], req)

	resData := api.UploadRes{FileUri: fileUri}

	Sucess(w, resData, nil)
	return
}

func (s *FileserverService) PostUploadStream(w http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()
	filename := base64.StdEncoding.EncodeToString([]byte(uuid.NewRandom().String()))
	filepath, err := s.genFilepath(filename, "")
	if err != nil {
		log.Println(err)
		Error(w, 100, err.Error())
	}
	log.Println("output file path:", filepath)
	fout, _ := os.Create(filepath)

	defer fout.Close()
	_, err = io.Copy(fout, req.Body)
	if err != nil {
		log.Println("Upload Stream error:", err)
		Error(w, 1, fmt.Sprintf("Upload Stream error:%s", err))
		return
	}
	fileUri := s.genFileUri(filepath[len(s.path):], req)
	extra := map[string]string{
		"uri": fileUri,
	}
	Sucess(w, nil, extra)
	return
}

func (s *FileserverService) genFileUri(filepath string, req *http.Request) string {
	if s.host == "" {
		return "http://" + req.Host + filepath
	}
	if strings.HasPrefix(s.host, "http") {
		log.Println(s.host, filepath)
		return s.host + filepath
	} else {
		return "http://" + path.Join(s.host, filepath)
	}
}

func (s *FileserverService) validatePath(inputpath string) error {
	dir := filepath.Join(s.path, inputpath)

	if len(filepath.Dir(dir)) < len(s.path) {
		return errors.New("invalid path:" + inputpath)
	}
	return nil
}

func (s *FileserverService) ServeFile(w http.ResponseWriter, req *http.Request) {
	// if !checkAuth(req) {
	// 	w.Header().Set("WWW-Authenticate", `Basic realm="Lego File Server"`)
	// 	w.WriteHeader(401)
	// 	w.Write([]byte("401 Unauthorized\n"))
	// }
	s.fileserver.ServeHTTP(w, req)
}

func (s *FileserverService) PostMkdir(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	req.ParseMultipartForm(32 << 20)

	inputpath := req.FormValue("path")
	log.Println("input path:", inputpath)
	if err := s.validatePath(inputpath); err != nil {
		Error(w, 100, err.Error())
		return
	}
	abspath := filepath.Join(s.path, inputpath)
	log.Println("abspath:", abspath)

	err := os.MkdirAll(abspath, 0777)
	if err != nil {
		log.Println(err)
		Error(w, 1, err.Error())
		return
	}
	Sucess(w, nil, nil)
}

//todo 过滤 .. 等字符
func (s *FileserverService) DeleteFile(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	req.ParseMultipartForm(32 << 20)

	inputpath := req.FormValue("path")
	if err := s.validatePath(inputpath); err != nil {
		Error(w, 100, err.Error())
		return
	}
	force := req.FormValue("force") //是否强制删除
	log.Println("input path:", inputpath, ",force:", force)

	abspath := filepath.Join(s.path, inputpath)
	log.Println("abspath:", abspath)

	err := os.Remove(abspath)
	if err != nil {
		log.Println(err)
		if force == "true" {
			if err = os.RemoveAll(abspath); err != nil {
				Error(w, 1, err.Error())
				return
			}
		} else {
			Error(w, 1, err.Error())
			return
		}
	}
	Sucess(w, nil, nil)
}

func (s *FileserverService) GetFileList(w http.ResponseWriter, req *http.Request) {
	inputpath := req.FormValue("path")
	log.Println("input path:", inputpath)
	if err := s.validatePath(inputpath); err != nil {
		Error(w, 100, err.Error())
		return
	}
	abspath := filepath.Join(s.path, inputpath)
	log.Println("abspath:", abspath)

	fileList, err := ioutil.ReadDir(abspath)
	if err != nil {
		log.Println(err)
		Error(w, 1, err.Error())
		return
	}
	var infolist []api.FileInfo
	for _, v := range fileList {
		log.Println(v.Name())
		url := "http://" + filepath.Join(req.Host, inputpath, v.Name())
		filetype := 0
		if v.IsDir() {
			filetype = 1
		}
		info := api.FileInfo{
			Name:     v.Name(),
			Path:     inputpath,
			FilePath: filepath.Join(inputpath, v.Name()),
			Url:      url,
			Type:     filetype,
		}
		infolist = append(infolist, info)
	}
	Sucess(w, infolist, nil)
}

func (s *FileserverService) genFilepath(name string, inputpath string) (string, error) {
	if err := s.validatePath(inputpath); err != nil {
		log.Println(err)
		return "", err
	}
	var dir string
	if inputpath == "" {
		now := time.Now()
		dir = fmt.Sprintf("%s/%s/%s", s.path, now.Format("20060102"), now.Format("15"))
	} else {
		dir = filepath.Join(s.path, inputpath)
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0777)
	}
	return filepath.Join(dir, name), nil
}

func checkAuth(req *http.Request) bool {
	s := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}
	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}
	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}
	return pair[0] == "lego" && pair[1] == "legotest"
}

func Sucess(w http.ResponseWriter, data interface{}, extra interface{}) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")  //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "*") //header的类型
	w.Header().Set("Content-Type", "application/json")

	var res = api.CommonRes{
		Code:    0,
		Message: "",
		Data:    data,
		Extra:   extra,
	}
	json.NewEncoder(w).Encode(res)

	return nil
}

func Error(w http.ResponseWriter, code int, message string) error {

	w.Header().Set("Access-Control-Allow-Origin", "*")  //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "*") //header的类型
	w.Header().Set("Content-Type", "application/json")

	var res = api.CommonRes{
		Code:    code,
		Message: message,
	}
	json.NewEncoder(w).Encode(res)

	return nil
}
