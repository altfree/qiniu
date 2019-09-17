package initialize

import "github.com/qiniu/api.v7/storage"

const (

	//图拍审核地址
	AUDIT_HOST    = "http://ai.qiniuapi.com"
	AUDIT_IMG_URL = "/v3/image/censor"
)

//图片审核请求内容
type AuditParam struct {
	Data struct {
		Uri string `json:"uri"`
	} `json:"data"`
	Params struct {
		Scenes string `json:scenes`
	} `json:"params"`
}

type Qiniu struct {
	AccessKey          string
	SecretKey          string
	Bucket             string
	Expires            int64
	TranscodeNotifyUrl string
	UploadNotoifyUrl   string
}

// 自定义七牛返回body

type NotifyBody struct {
	Key    string `json:key`
	Hash   string `json:hash`
	Fsize  int    `json:fsize`
	Bucket string `json:bucket`
	Name   string `json:name`
}

type UploadParam struct {
	storage.PutPolicy
}
