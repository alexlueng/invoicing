package util

import (
	"crypto/sha1"
	"encoding/hex"
	"sort"
	"strings"
)

// CheckSignature 微信公众号签名检查
func CheckSignature(signature, timestamp, nonce, token string) bool {
	arr := []string{timestamp, nonce, token}
	// 字典序排序
	sort.Strings(arr)

	n := len(timestamp) + len(nonce) + len(token)
	var b strings.Builder
	b.Grow(n)
	for i := 0; i < len(arr); i++ {
		b.WriteString(arr[i])
	}

	return Sha1(b.String()) == signature
}

// 进行Sha1编码
func Sha1(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

type AccessTokenResponse struct {
	AccessToken string  `json:"access_token"`
	ExpiresIn   float64 `json:"expires_in"`
}

func FetchAccessToken(appID, appSecret, accessTokenFetchUrl string) (string, error) {

/*	requestLine := strings.Join([]string{accessTokenFetchUrl,
		"?grant_type=client_credential&appid=",
		appID,
		"&secret=",
		appSecret}, "")

	resp, err := http.Get(requestLine)
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Println("发送get请求获取 atoken 错误", err)
		//logUtils.GetLog().Error("发送get请求获取 atoken 错误", err)
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("发送get请求获取 atoken 读取返回body错误", err)
		//logUtils.GetLog().Error("发送get请求获取 atoken 读取返回body错误", err)
		return "", err
	}

	if bytes.Contains(body, []byte("access_token")) {
		atr := AccessTokenResponse{}
		err = json.Unmarshal(body, &atr)
		if err != nil {
			fmt.Println("发送get请求获取 atoken 返回数据json解析错误", err)
			//logUtils.GetLog().Error("发送get请求获取 atoken 返回数据json解析错误", err)
			return "", err
		}
		//return atr.AccessToken, atr.ExpiresIn, nil
	} else {
		fmt.Println("发送get请求获取 微信返回 err")
		ater := models.AccessTokenErrorResponse{}
		err = json.Unmarshal(body, &ater)
		fmt.Printf("发送get请求获取 微信返回 的错误信息 %+v\n", ater)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("%s", ater.Errmsg)
	}*/
	return "",nil
}

