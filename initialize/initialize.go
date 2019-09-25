package initialize

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/qiniu/api.v7/auth"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/conf"
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

func (qn *Qiniu) GetDownloadAddr(vod string, hlstype bool) string {

	host := qn.Host
	if hlstype {
		host = host + "?pm3u8/0" //提取私有空间hls切片
	}
	deadline := time.Now().Add(time.Second * 3600).Unix() //1小时有效期
	return storage.MakePrivateURL(authMac, host, vod, deadline)

}

//增加转码数据
func (qn *Qiniu) AddTranscode(transcode Transcode) (string, error) {

	fopAvthumb := fmt.Sprintf(transcode.VodLayout+"%s",
		storage.EncodedEntry(qn.Bucket, transcode.VodLater))
	fopVframe := fmt.Sprintf(transcode.PreviewImgLayout+"%s",
		storage.EncodedEntry(qn.Bucket, transcode.PreviewImgName))
	fopBatch := []string{fopAvthumb, fopVframe}
	fops := strings.Join(fopBatch, ";")
	force := true
	notifyURL := qn.TranscodeNotifyUrl //异步通知url
	fmt.Println(transcode.VodLater)
	return qn.storage().Pfop(qn.Bucket, transcode.VodRaw, fops, "notify", notifyURL, force)

}

//查询转码状态
//Code: 状态码0成功，1等待处理，2正在处理，3处理失败，4通知提交失败
func (qn *Qiniu) TranscodingStatus(id string) (*storage.PrefopRet, error) {

	res, err := qn.storage().Prefop(id)
	if err != nil {
		return nil, err
	}
	return &res, nil
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
func (qn *Qiniu) VerifyCallback(req *http.Request) (bool, error) {
	return authMac.VerifyCallback(req)
}

//内容审核
func (qn *Qiniu) AuditMedia(url string, ap []string) (*AuditResponse, error) {

	var param AuditParam
	param.Data.Uri = url
	param.Params.Scenes = ap
	res, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, AUDIT_IMG_URL, strings.NewReader(string(res)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", conf.CONTENT_TYPE_JSON)
	err = authMac.AddToken(auth.TokenQiniu, req)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 || response.Header.Get("X-Resp-Code") != "200" {
		return nil, errors.New("code:" + response.Status + ",response-code:" + response.Header.Get("X-Resp-Code"))
	}
	var respBody AuditResponse
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		return nil, err
	}
	return &respBody, nil
}
