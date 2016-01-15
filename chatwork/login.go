package chatwork

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

var (
	baseURL     string
	client      http.Client
	accessToken string
	myid        string
	contacts    map[string]contact
	rooms       map[string]room
)

func Login(email, password string) error {
	jar, _ := cookiejar.New(nil)
	client = http.Client{Jar: jar}
	resp, _ := client.PostForm("https://www.chatwork.com/login.php", url.Values{
		"email":    {email},
		"password": {password},
	})
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if !regexp.MustCompile(`ACCESS_TOKEN\s+=\s+'[0-9A-Za-z]+'`).Match(body) {
		return fmt.Errorf("login failed.")
	}
	accessToken = string(regexp.MustCompile(`ACCESS_TOKEN\s+=\s+'([0-9A-Za-z]+)'`).FindSubmatch(body)[1])
	myid = string(regexp.MustCompile(`myid\s+=\s+'([0-9]+)'`).FindSubmatch(body)[1])
	isKDDI := regexp.MustCompile(`PLAN_NAME\s+=\s+'KDDI ChatWork'`).Match(body)
	if isKDDI {
		baseURL = "https://kcw.kddi.ne.jp"
	} else {
		baseURL = "https://www.chatwork.com"
	}
	initLoad()
	if !isKDDI {
		getAccountInfo()
	}
	return nil
}
