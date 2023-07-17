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
	bx := os.Getenv("BX")
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
		token, err := funcaptcha.GetOpenAITokenWithBx(bx)
		if !strings.Contains(token, "sup=1") {
			logger.Warn("BX is expired.")
			return
		}

		if err != nil {
			logger.Error("Failed to get arkose token.")
			return
		}

		date := time.Now().Format("2006-01-02 15:04:05")
		var payload = Payload{
			Date:  date,
			Token: token,
		}
		if err := submitXyHelperToken(payload); err != nil {
			logger.Info(err.Error())
			return
		}

		totalSubmitted++
		logger.Info(fmt.Sprintf("Token is submitted at %s (%d).", date, totalSubmitted))

		sleepTime, _ := strconv.Atoi(interval)
		time.Sleep(time.Second * time.Duration(sleepTime))
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
