package initialize

import "github.com/qiniu/api.v7/storage"

const (

	//图拍审核地址
	AUDIT_HOST    = "ai.qiniuapi.com"
	AUDIT_IMG_URL = "/v3/image/censor"
)

//图片审核请求内容
type AuditParam struct {
	Data struct {
		Uri string `json:"uri"`
	} `json:"data"`
	Params struct {
		Scenes []string `json:"scenes,omitempty"`
	} `json:"params,omitempty"`
}

type Qiniu struct {
	AccessKey          string
	SecretKey          string
	Bucket             string
	Expires            int
	Host               string
	HostHls            string
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

//定义转码参数
type Transcode struct {
	VodRaw           string `json:vod_raw`
	VodLater         string `json:vod_later`
	VodLayout        string `json:vod_layout`
	PreviewImgName   string `json:pre_img_name`
	PreviewImgLayout string `json:pre_img_layout`
}

//上传凭证参数
type UploadParam struct {
	storage.PutPolicy
}
