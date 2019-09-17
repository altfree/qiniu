package initialize

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/qiniu/api.v7/auth"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

var authMac *auth.Credentials

//初始化七牛sdk
func (qn *Qiniu) NewQi(AccessKey, SecretKey string) {
	qn.AccessKey, qn.SecretKey = AccessKey, SecretKey
	authMac = qbox.NewMac(qn.AccessKey, qn.SecretKey)
}

//GetUploadEvi获取上传凭证
//通过设置上传凭证的参数，可以实现文件在上传完成后对视频实现转码

func (qn *Qiniu) GetUploadEvi(putPolicy UploadParam) string {

	if putPolicy.Scope == "" {
		putPolicy.Scope = qn.Bucket
	}
	//转码通知url
	if putPolicy.PersistentNotifyURL == "" {
		putPolicy.PersistentNotifyURL = qn.TranscodeNotifyUrl
	}
	//上传成功后通知url
	if putPolicy.CallbackURL == "" {
		putPolicy.CallbackURL = qn.UploadNotoifyUrl
	}

	return putPolicy.UploadToken(authMac)
}

//GetDownloadAddr获取文件下载地址
//vod加转码参数实现视频流实时转码，目前仅支持hls格式

func (qn *Qiniu) GetDownloadAddr(domain, vod string) string {

	deadline := time.Now().Add(time.Second * 3600).Unix() //1小时有效期
	return storage.MakePrivateURL(authMac, domain, vod, deadline)

}

//增加转码数据
func (qn *Qiniu) AddTranscode(rawvod, latervod, layout, img string) (string, error) {

	fopAvthumb := fmt.Sprintf(layout+"%s",
		storage.EncodedEntry(qn.Bucket, latervod))
	fopVframe := fmt.Sprintf("vframe/jpg/offset/10|saveas/%s",
		storage.EncodedEntry(qn.Bucket, img))
	fopBatch := []string{fopAvthumb, fopVframe}
	fops := strings.Join(fopBatch, ";")
	force := true
	notifyURL := qn.TranscodeNotifyUrl //异步通知url
	return qn.storage().Pfop(qn.Bucket, rawvod, fops, "notify", notifyURL, force)

}

//查询转码状态
func (qn *Qiniu) TranscodingStatus(id string) (string, error) {

	ret, err := qn.storage().Prefop(id)
	if err != nil {
		return "", err
	}
	fmt.Println(ret.String())
	return ret.String(), nil
}

//地域信息
func (qn *Qiniu) storage() *storage.OperationManager {

	cfg := storage.Config{
		UseHTTPS: false,
	}
	cfg.Zone = &storage.ZoneHuabei
	return storage.NewOperationManager(authMac, &cfg)

}

//验证请求是否来自七牛
func (qn *Qiniu) VerifyCallback(mac *auth.Credentials, req *http.Request) (bool, error) {
	return mac.VerifyCallback(req)
}

//内容审核
func (qn *Qiniu) AuditMedia(url string, kind string) {

	var param AuditParam
	param.Data.Uri = url
	param.Params.Scenes = kind
	res, err := json.Marshal(param)
	if err != nil {
		panic(err)
	}
	data := "POST" + " " + AUDIT_IMG_URL + "\nHost: " + AUDIT_HOST + "\nContent-Type: " + "application/json" + "\n\n" + string(res)

	token := authMac.Sign([]byte(data))
	client := &http.Client{}
	reque, err := http.NewRequest("POST", AUDIT_HOST+AUDIT_IMG_URL, strings.NewReader(string(res)))
	if err != nil {
		panic(err)
	}
	token = "Qiniu " + token
	fmt.Println(data)
	reque.Header.Add("Content-Type", "application/json")
	reque.Header.Add("Authorization", token)
	response, err := client.Do(reque)
	if err != nil {
		panic(err)
	}
	response.Body.Close()
	v, _ := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(response.Status)
	fmt.Println(v)
	// http.Post(AUDIT_HOST+AUDIT_IMG_URL, "application/json", strings.NewReader(string(res)))
	// return
}
