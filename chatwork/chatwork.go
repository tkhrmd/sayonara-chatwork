package chatwork

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"
)

var (
	baseURL  string
	client   http.Client
	token    string
	myid     string
	contacts map[string]contact
	rooms    map[string]room
)

func Login(email, password string) error {
	jar, _ := cookiejar.New(nil)
	client = http.Client{Jar: jar}
	form := url.Values{}
	form.Add("email", email)
	form.Add("password", password)
	resp, _ := client.PostForm("https://www.chatwork.com/login.php", form)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if !regexp.MustCompile(`ACCESS_TOKEN\s+=\s+'[0-9A-Za-z]+'`).Match(body) {
		return fmt.Errorf("login failed")
	}
	token = string(regexp.MustCompile(`ACCESS_TOKEN\s+=\s+'([0-9A-Za-z]+)'`).FindSubmatch(body)[1])
	myid = string(regexp.MustCompile(`myid\s+=\s+'([0-9]+)'`).FindSubmatch(body)[1])
	if regexp.MustCompile(`PLAN_NAME\s+=\s+'KDDI ChatWork'`).Match(body) {
		baseURL = "https://kcw.kddi.ne.jp"
		initLoad()
	} else {
		baseURL = "https://www.chatwork.com"
		initLoad()
		getAccountInfo()
	}
	return nil
}

type initLoadResult struct {
	Result struct {
		Rooms    map[string]room    `json:"room_dat"`
		Contacts map[string]contact `json:"contact_dat"`
	}
}

type room struct {
	N string
	M map[string]int
}

type contact struct {
	Aid int
	Nm  string
	Rid int
}

func initLoad() {
	resp, _ := client.Get(baseURL + "/gateway.php?cmd=init_load&myid=" + myid + "&_v=1.80a&_av=4&_t=" + token + "&ln=ja&rid=0&type=&new=1")
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	result := initLoadResult{}
	json.Unmarshal(body, &result)
	rooms = result.Result.Rooms
	contacts = result.Result.Contacts
}

type pdata struct {
	Aid            []int `json:"aid"`
	GetPrivateData int   `json:"get_private_data"`
}

type getAccountInfoResult struct {
	Result struct {
		Accounts map[string]contact `json:"account_dat"`
	}
}

func getAccountInfo() {
	aidMap := map[int]bool{}
	for _, v := range rooms {
		for k, _ := range v.M {
			aid, _ := strconv.Atoi(k)
			aidMap[aid] = true
		}
	}
	for _, v := range contacts {
		delete(aidMap, v.Aid)
	}
	aidSet := []int{}
	for k, _ := range aidMap {
		aidSet = append(aidSet, k)
	}
	sort.Sort(sort.IntSlice(aidSet))
	pdata, _ := json.Marshal(pdata{
		Aid:            aidSet,
		GetPrivateData: 0,
	})
	resp, _ := client.PostForm(baseURL+"/gateway.php?cmd=get_account_info&myid="+myid+"&_v=1.80a&_av=4&_t="+token+"&ln=ja", url.Values{
		"pdata": {string(pdata)},
	})
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	result := getAccountInfoResult{}
	json.Unmarshal(body, &result)
	for k, v := range result.Result.Accounts {
		contacts[k] = v
	}
}

type loadOldChatResult struct {
	Result struct {
		ChatList chatList `json:"chat_list"`
	}
}

type chatList []chat

type chat struct {
	Aid int
	Id  int
	Msg string
	Tm  int
}

func (l chatList) Len() int {
	return len(l)
}

func (l chatList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l chatList) Less(i, j int) bool {
	return l[i].Id < l[j].Id
}

func loadOldChat(roomId, firstChatId int) chatList {
	resp, _ := client.Get(baseURL + "/gateway.php?cmd=load_old_chat&myid=" + myid + "&_v=1.80a&_av=4&_t=" + token + "&ln=ja&room_id=" + strconv.Itoa(roomId) + "&first_chat_id=" + strconv.Itoa(firstChatId))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	result := loadOldChatResult{}
	json.Unmarshal(body, &result)
	return result.Result.ChatList
}

func GetRoomName(rid int) string {
	room, _ := rooms[strconv.Itoa(rid)]
	if len(room.N) != 0 {
		return room.N
	}
	for _, v := range contacts {
		if v.Rid == rid {
			return v.Nm
		}
	}
	return ""
}

func Export(rid int, file *os.File) {
	writer := csv.NewWriter(file)
	firstChatId := 0
	for {
		chatList := loadOldChat(rid, firstChatId)
		sort.Sort(sort.Reverse(chatList))
		name := ""
		for _, v := range chatList {
			name = contacts[strconv.Itoa(v.Aid)].Nm
			if len(name) == 0 {
				name = strconv.Itoa(v.Aid)
			}
			writer.Write([]string{
				time.Unix(int64(v.Tm), 0).Format("2006-01-02 15:04:05"),
				name,
				v.Msg,
			})
			firstChatId = v.Id
		}
		writer.Flush()
		if len(chatList) < 40 {
			break
		}
	}
}
