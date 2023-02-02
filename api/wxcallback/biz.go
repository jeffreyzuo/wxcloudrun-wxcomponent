package wxcallback

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/errno"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/log"

	"github.com/WeixinCloud/wxcloudrun-wxcomponent/db/dao"
	"github.com/WeixinCloud/wxcloudrun-wxcomponent/db/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type wxCallbackBizRecord struct {
	CreateTime   int64  `json:"CreateTime"`
	ToUserName   string `json:"ToUserName"`
	FromUserName string `json:"FromUserName"`
	Content      string `json:"Content"`
	MsgType      string `json:"MsgType"`
	Event        string `json:"Event"`
}

func bizHandler(c *gin.Context) {
	// 记录到数据库
	body, _ := ioutil.ReadAll(c.Request.Body)
	var json wxCallbackBizRecord
	if err := binding.JSON.BindBody(body, &json); err != nil {
		c.JSON(http.StatusOK, errno.ErrInvalidParam.WithData(err.Error()))
		return
	}
	r := model.WxCallbackBizRecord{
		CreateTime:  time.Unix(json.CreateTime, 0),
		ReceiveTime: time.Now(),
		Appid:       c.Param("appid"),
		ToUserName:  json.ToUserName,
		MsgType:     json.MsgType,
		Event:       json.Event,
		PostBody:    string(body),
	}
	if json.CreateTime == 0 {
		r.CreateTime = time.Unix(1, 0)
	}
	if err := dao.AddBizCallBackRecord(&r); err != nil {
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}

	// 转发到用户配置的地址
	proxyOpen, err := proxyCallbackMsg("", json.MsgType, json.Event, string(body), c)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusOK, errno.ErrSystemError.WithData(err.Error()))
		return
	}
	if !proxyOpen {
		c.String(http.StatusOK, fmt.Sprint(c.Writer))
	}
	//文本消息回复测试
	log.Info("Proxy over.")
	if json.MsgType == "Text" {
		log.Info("Text message response")
		ct := time.Unix(1, 0)
		c.String(http.StatusOK, fmt.Sprintf("<xml>\n  <ToUserName><![CDATA[%s]]></ToUserName>\n  <FromUserName><![CDATA[%s]]></FromUserName>\n  <CreateTime>%d</CreateTime>\n  <MsgType><![CDATA[text]]></MsgType>\n  <Content><![CDATA[%s]]></Content>\n</xml>", json.FromUserName, json.ToUserName, ct, json.Content))
	}
}
