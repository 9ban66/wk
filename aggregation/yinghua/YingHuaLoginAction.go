package yinghua

import (
	"errors"
	"fmt"
	"os"
	"strings"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
	yinghuaApi "github.com/yatori-dev/yatori-go-core/api/yinghua"
	"github.com/yatori-dev/yatori-go-core/utils"
	"github.com/yatori-dev/yatori-go-core/utils/log"
)

var errYingHuaBadCaptcha = errors.New("yinghua bad captcha")

// YingHuaLoginAction 登录API聚合整理
// {"refresh_code":1,"status":false,"msg":"账号密码不正确"}
// {"_code": 1, "status": false,"msg": "账号登录超时，请重新登录", "result": {}}
func YingHuaLoginAction(cache *yinghuaApi.YingHuaUserCache) error {
	const maxLoginAttempts = 5
	for attempt := 1; attempt <= maxLoginAttempts; attempt++ {
		path, _ := cache.VerificationCodeApi(3)
		if path == "" {
			return errors.New("无法正常获取英华验证码，请检查平台地址是否正确")
		}
		img, _ := utils.ReadImg(path)
		codeResult := ddddocr.SemiOCRVerification(img, ort.NewShape(1, 18))
		utils.DeleteFile(path)
		if strings.TrimSpace(codeResult) == "" || strings.EqualFold(strings.TrimSpace(codeResult), "stub") {
			return errors.New("英华登录失败：ddddocr 未正确启用，请确认已移除本地 stub replace，并使用 CGO_ENABLED=1 构建")
		}
		cache.SetVerCode(codeResult)

		jsonStr, err := cache.LoginApi(3, nil)
		if err != nil {
			return fmt.Errorf("英华登录请求失败: %w", err)
		}
		log.Print(log.DEBUG, "["+cache.Account+"] "+"LoginAction---"+jsonStr)
		if err := applyYingHuaLoginResponse(cache, jsonStr); errors.Is(err, errYingHuaBadCaptcha) {
			continue
		} else {
			return err
		}
	}
	return errors.New("英华登录失败：验证码识别多次失败或平台响应异常")
}

func applyYingHuaLoginResponse(cache *yinghuaApi.YingHuaUserCache, jsonStr string) error {
	msg := gojsonq.New().JSONString(jsonStr).Find("msg")
	if msg == "验证码有误！" {
		return errYingHuaBadCaptcha
	}
	if strings.Contains(jsonStr, ">选择学校<") {
		return fmt.Errorf("请填写正确的英华登录后 URL，首页地址和登录后地址可能不一样")
	}
	redirect := gojsonq.New().JSONString(jsonStr).Find("redirect")
	if redirect == nil {
		if msg == nil {
			return errors.New("英华登录响应异常，未返回登录跳转地址")
		}
		return errors.New(fmt.Sprint(msg))
	}
	redirectStr, ok := redirect.(string)
	if !ok || !strings.Contains(redirectStr, "token=") || !strings.Contains(redirectStr, "&sign=") {
		return errors.New("英华登录响应异常，登录跳转地址缺少 token 或 sign")
	}
	cache.SetToken(strings.Split(strings.Split(redirectStr, "token=")[1], "&")[0])
	cache.SetSign(strings.Split(redirectStr, "&sign=")[1])
	return nil
}

// LoginTimeoutAfreshAction 超时重登逻辑
func LoginTimeoutAfreshAction(cache *yinghuaApi.YingHuaUserCache, backJson string) {
	//未超时则直接return
	if !strings.Contains(backJson, "账号登录超时，请重新登录") {
		return
	}
	log.Print(log.INFO, "["+cache.Account+"] ", log.BoldRed, "检测到登录超时，正在进行重新登录逻辑...")
	err := YingHuaLoginAction(cache)
	if err != nil {
		log.Print(log.INFO, "["+cache.Account+"] ", log.BoldRed, "超时重登失败")
		os.Exit(0)
	}
	log.Print(log.INFO, "["+cache.Account+"] ", log.BoldGreen, "超时重登成功")
}
