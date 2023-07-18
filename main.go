package main

//goland:noinspection GoSnakeCaseUsage
import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/linweiyuan/funcaptcha"
	"github.com/linweiyuan/go-logger/logger"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

var client tls_client.HttpClient

//goland:noinspection GoUnhandledErrorResult
func init() {
	client, _ = tls_client.NewHttpClient(tls_client.NewNoopLogger(), []tls_client.HttpClientOption{
		tls_client.WithCookieJar(tls_client.NewCookieJar()),
		tls_client.WithClientProfile(tls_client.Okhttp4Android13),
	}...)

	proxy := os.Getenv("PROXY")
	if proxy != "" {
		client.SetProxy(proxy)
	}

	funcaptcha.SetTLSClient(client)
}

type Payload struct {
	Date  string `json:"date"`
	Token string `json:"token"`
}

//goland:noinspection GoUnhandledErrorResult
func main() {
	ctx := gctx.New()
	// bx := os.Getenv("BX")
	bx := g.Cfg().MustGetWithEnv(ctx, "BX").String()

	if bx == "" {
		logger.Error("Please set BX.")
		return
	}

	interval := os.Getenv("INTERVAL")
	if interval == "" {
		interval = "3"
	}

	// fix 403
	funcaptcha.GetOpenAITokenWithBx(bx)

	totalSubmitted := 0
	for {
		sleepTime, _ := strconv.Atoi(interval)
		time.Sleep(time.Second * time.Duration(sleepTime))

		token, err := funcaptcha.GetOpenAITokenWithBx(bx)
		if err != nil {
			logger.Error("Failed to get arkose token, please try again later.")
			continue

		}

		if !strings.Contains(token, "sup=1") {
			logger.Warn("BX is expired.")
			continue
		}

		date := time.Now().Format("2006-01-02 15:04:05")
		var payload = Payload{
			Date:  date,
			Token: token,
		}
		if err := submitXyHelperToken(payload); err != nil {
			logger.Info(err.Error())
			continue
		}

		totalSubmitted++
		logger.Info(fmt.Sprintf("Token is submitted at %s (%d).", date, totalSubmitted))

	}
}

//goland:noinspection GoUnhandledErrorResult,GoErrorStringFormat
func submitXyHelperToken(payload Payload) error {
	jsonBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "https://chatarkose.xyhelper.cn/pushtoken", bytes.NewReader(jsonBytes))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to submit arkose token with status code: %d", resp.StatusCode)
	}

	return nil
}
