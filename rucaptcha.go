package rucaptcha

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const MaxRetriesCaptcha = 7
const CaptchaNotReady = "CAPCHA_NOT_READY"

type RucaptchaResponse struct {
	Status  int32
	Request string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

/**
 * Send captcha base64 to rucaptcha
 */
func SendBase64(base64image string) string {
	rucaptchaKey := os.Getenv("RUCAPTCHA_KEY")
	contentType := "application/x-www-form-urlencoded"
	url := "http://rucaptcha.com/in.php"
	dataStr := fmt.Sprintf(`{"method":"base64","key":"%s","body":"%s","json":"1"}`, rucaptchaKey, base64image)
	data := []byte(dataStr)

	var rucaptchaResp RucaptchaResponse

	resp, err := http.Post(url, contentType, bytes.NewReader(data))

	if err != nil {
		log.Fatal(err)
		return ""
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if err = json.Unmarshal(body, &rucaptchaResp); err != nil {
		log.Fatal(err)
		return ""
	}

	fmt.Println(rucaptchaResp.Request)

	return rucaptchaResp.Request
}

func Retreive(id string, captcha *string) error {
	return retry(MaxRetriesCaptcha, time.Second, func() error {
		rucaptchaKey := os.Getenv("RUCAPTCHA_KEY")
		url := fmt.Sprintf("http://rucaptcha.com/res.php?key=%s&action=get&id=%s&json=1", rucaptchaKey, id)

		var rucaptchaResp RucaptchaResponse
		resp, err := http.Get(url)

		if err != nil {
			log.Fatal(err)
			return err
		}

		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)

		if err = json.Unmarshal(body, &rucaptchaResp); err != nil {
			log.Fatal(err)
			return err
		}

		if rucaptchaResp.Request == CaptchaNotReady {
			return errors.New("Captcha is not ready")
		}

		*captcha = rucaptchaResp.Request

		return nil
	})
}

// https://upgear.io/blog/simple-golang-retry-function/
func retry(attempts int, sleep time.Duration, f func() error) error {
	if err := f(); err != nil {
		if s, ok := err.(stop); ok {
			// Return the original error for later checking
			return s.error
		}

		if attempts--; attempts > 0 {
			// Add some randomness to prevent creating a Thundering Herd
			jitter := time.Duration(rand.Int63n(int64(sleep)))
			sleep = sleep + jitter/2

			time.Sleep(sleep)
			return retry(attempts, 2*sleep, f)
		}
		return err
	}

	return nil
}

type stop struct {
	error
}
