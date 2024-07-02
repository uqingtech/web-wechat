package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"gitee.ltd/lxh/logger/log"
	"github.com/eatmoreapple/openwechat"

	"web-wechat/core"
	"web-wechat/oss"
)

// 处理文本消息
func voiceMessageHandle(ctx *openwechat.MessageContext) {
	sender, _ := ctx.Sender()
	senderUser := sender.NickName
	if ctx.IsSendByGroup() {
		// 取出消息在群里面的发送者
		senderInGroup, _ := ctx.SenderInGroup()
		senderUser = fmt.Sprintf("%v[%v]", senderInGroup.NickName, senderUser)
	}
	fileResp, err := ctx.GetVoice()
	if err != nil {
		log.Errorf("获取语音消息失败: %v", err.Error())
		return
	}
	defer fileResp.Body.Close()
	voiceByte, err := io.ReadAll(fileResp.Body)
	if err != nil {
		log.Errorf("voice读取错误: %v", err.Error())
		return
	} else {
		// 读取文件相关信息
		contentType := http.DetectContentType(voiceByte)
		// fileType := strings.Split(contentType, "/")[1]
		fileName := fmt.Sprintf("%v.%v", ctx.MsgId, "mp3")
		if user, err := ctx.Bot().GetCurrentUser(); err == nil {
			uin := user.Uin
			fileName = fmt.Sprintf("%v/%v", uin, fileName)
		}
		log.Infof("[收到新voice消息] == 发信人：%v", senderUser)

		// 上传文件
		reader2 := io.NopCloser(bytes.NewReader(voiceByte))
		flag := oss.SaveToOss(reader2, contentType, fileName)
		if flag {
			fileUrl := fmt.Sprintf("https://%v/%v/%v", core.SystemConfig.OssConfig.Endpoint, core.SystemConfig.OssConfig.BucketName, fileName)
			log.Infof("voice保存成功, voice链接: %v", fileUrl)
			ctx.Content = fileUrl
		} else {
			log.Error("voice保存失败")
		}
	}
	ctx.Next()
}
