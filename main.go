package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	//"reflect"
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
)

var (
	withFile    = flag.Bool("f", false, "download files")
	email       = flag.String("e", "", "Email Address")
	password    = flag.String("p", "", "Password")
	client      = http.Client{}
	accessToken string
	myid        string
	roomDat     = map[string]RoomDatItem{}
	contactDat  = map[string]ContactDatItem{}
)

const (
	lazySelect = "0"
	loadType   = ""
	apiVersion = "4"
	version    = "1.80a"
	luguage    = "ja"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: ")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	jar, _ := cookiejar.New(nil)
	client = http.Client{
		Jar: jar,
	}

	login()
	initLoad()
	getAccountInfo()
}

func login() {
	form := url.Values{}
	form.Add("email", *email)
	form.Add("password", *password)

	resp, _ := client.PostForm("https://www.chatwork.com/login.php", form)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if regexp.MustCompile(`ACCESS_TOKEN\s+=\s+'[0-9A-Za-z]+'`).Match(body) {
	}
	accessToken = string(regexp.MustCompile(`ACCESS_TOKEN\s+=\s+'([0-9A-Za-z]+)'`).FindSubmatch(body)[1])
	myid = string(regexp.MustCompile(`myid\s+=\s+'([0-9]+)'`).FindSubmatch(body)[1])
}

type InitLoadItem struct {
	Result ResultItem
}

type ResultItem struct {
	RoomDat    map[string]RoomDatItem    `json:"room_dat"`
	ContactDat map[string]ContactDatItem `json:"contact_dat"`
}

type RoomDatItem struct {
	Mid int
	// R   int
	// Tp  int
	// C   int
	// F   int
	// T   int
	// Lt  int
	M map[string]int
}

type ContactDatItem struct {
	Aid int
	Gid int
	Nm  string
	// Cwid string
	// Onm  string
	// Dp   string
	// Sp   string
	// Fb   string
	// Tw   string
	// Ud   int
	// Av   string
	// Cv   string
	// Name string
	// Avid string
	// Mrid string
	// Rid  string
}

func initLoad() {
	query := url.Values{}
	query.Add("cmd", "init_load")
	query.Add("myid", myid)
	query.Add("_v", version)
	query.Add("_av", apiVersion)
	query.Add("_t", accessToken)
	query.Add("ln", luguage)
	query.Add("rid", lazySelect)
	query.Add("type", loadType)
	query.Add("new", "1")

	resp, _ := client.Get("https://www.chatwork.com/gateway.php?" + query.Encode())
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	result := InitLoadItem{}
	json.Unmarshal(body, &result)
	roomDat = result.Result.RoomDat
	contactDat = result.Result.ContactDat
}

type Pdata struct {
	Aid            []int `json:"aid"`
	GetPrivateData int   `json:"get_private_data"`
}

func getAccountInfo() {
	query := url.Values{}
	query.Add("cmd", "get_account_info")
	query.Add("myid", myid)
	query.Add("_v", "1.80a")
	query.Add("_av", "4")
	query.Add("_t", accessToken)
	query.Add("ln", "ja")
	query.Add("rid", "0")
	query.Add("type", "")
	query.Add("new", "1")

	mSet := map[int]bool{}
	for _, r := range roomDat {
		for id, _ := range r.M {
			i, _ := strconv.Atoi(id)
			mSet[i] = true
		}
	}
	for _, c := range contactDat {
		delete(mSet, c.Aid)
	}

	ms := []int{}
	for k, _ := range mSet {
		ms = append(ms, k)
	}
	sort.Sort(sort.IntSlice(ms))

	pdata := Pdata{
		Aid:            ms,
		GetPrivateData: 0,
	}

	pdataJson, _ := json.Marshal(pdata)

	fmt.Println(string(pdataJson))

	// form := url.Values{}
	// form.Add("pdata", pdataJson)

	// resp, _ := client.PostForm("https://www.chatwork.com/gateway.php?"+query.Encode(), form)
	// defer resp.Body.Close()

	// POSTでQuery月
	// フォームでJSON

	// resp, _ := client.Get("https://www.chatwork.com/gateway.php?" + values.Encode())
	// defer resp.Body.Close()

	// body, _ := ioutil.ReadAll(resp.Body)
	// result := InitLoadItem{}
	// json.Unmarshal(body, &result)

}
