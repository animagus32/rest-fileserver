package api

type UploadRes struct {
	FileUri string `json:"file_uri"`
}

// {"code":0,
//  "message":"",
//  "result":null,
//  "extra":{"uri":"http://p40xv5kt4.bkt.clouddn.com/faceimg-2018-02-22T11-40-55-_qigKvLTzpUz7bWq3nZ3tw=="}
// }

type CommonRes struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Extra   interface{} `json:"extra,omitempty"`
}

type FileInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	FilePath string `json:"file_path"`
	Url      string `json:"url"`
	Type     int    `json:"type"` //0 文件，1目录
}
