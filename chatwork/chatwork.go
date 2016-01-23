package chatwork

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
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
	client      http.Client
	baseURL     string
	token       string
	myid        string
	contactList map[string]contact
	roomList    map[string]room
)

func Login(email, password string) error {
	jar, _ := cookiejar.New(nil)
	client = http.Client{Jar: jar}
	form := url.Values{}
	form.Add("email", email)
	form.Add("password", password)
	resp, err := client.PostForm("https://www.chatwork.com/login.php", form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if !regexp.MustCompile(`ACCESS_TOKEN\s+=\s+'[0-9A-Za-z]+'`).Match(body) {
		return fmt.Errorf("login failed")
	}
	token = string(regexp.MustCompile(`ACCESS_TOKEN\s+=\s+'([0-9A-Za-z]+)'`).FindSubmatch(body)[1])
	myid = string(regexp.MustCompile(`myid\s+=\s+'([0-9]+)'`).FindSubmatch(body)[1])
	// change the end point by either KDDI ChatWork or ChatWork
	if regexp.MustCompile(`PLAN_NAME\s+=\s+'KDDI ChatWork'`).Match(body) {
		baseURL = "https://kcw.kddi.ne.jp"
		initLoad()
	} else {
		baseURL = "https://www.chatwork.com"
		initLoad()
		pause()
		getAccountInfo()
	}
	pause()
	return nil
}

type initLoadResult struct {
	Result struct {
		RoomList    map[string]room    `json:"room_dat"`
		ContactList map[string]contact `json:"contact_dat"`
	}
}

type room struct {
	Name    string         `json:"n"`
	Members map[string]int `json:"m"`
}

type contact struct {
	AccountId int    `json:"aid"`
	Name      string `json:"nm"`
	RoomId    int    `json:"rid"`
}

// get the chat list and contact list
func initLoad() error {
	resp, err := client.Get(baseURL + "/gateway.php?cmd=init_load&myid=" + myid + "&_v=1.80a&_av=4&_t=" + token + "&ln=ja&rid=0&type=&new=1")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	result := initLoadResult{}
	json.Unmarshal(body, &result)
	roomList = result.Result.RoomList
	contactList = result.Result.ContactList
	return nil
}

type privateData struct {
	AccountIds     []int `json:"aid"`
	GetPrivateData int   `json:"get_private_data"`
}

type getAccountInfoResult struct {
	Result struct {
		Accounts map[string]contact `json:"account_dat"`
	}
}

// get member's account data that haven't added contact list
func getAccountInfo() error {
	aidMap := map[int]bool{}
	for _, room := range roomList {
		for aidStr, _ := range room.Members {
			aid, _ := strconv.Atoi(aidStr)
			aidMap[aid] = true
		}
	}
	for _, contact := range contactList {
		delete(aidMap, contact.AccountId)
	}
	aidSet := []int{}
	for aid, _ := range aidMap {
		aidSet = append(aidSet, aid)
	}
	sort.Sort(sort.IntSlice(aidSet))
	pdata, _ := json.Marshal(privateData{
		AccountIds:     aidSet,
		GetPrivateData: 0,
	})
	form := url.Values{}
	form.Add("pdata", string(pdata))
	resp, err := client.PostForm(baseURL+"/gateway.php?cmd=get_account_info&myid="+myid+"&_v=1.80a&_av=4&_t="+token+"&ln=ja", form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	result := getAccountInfoResult{}
	json.Unmarshal(body, &result)
	for aid, account := range result.Result.Accounts {
		contactList[aid] = account
	}
	return nil
}

type loadOldChatResult struct {
	Result struct {
		ChatList chatList `json:"chat_list"`
	}
}

type chatList []chat

type chat struct {
	AccountId int    `json:"aid"`
	Id        int    `json:"id"`
	Message   string `json:"msg"`
	Time      int    `json:"tm"`
}

func (cl chatList) Len() int {
	return len(cl)
}

func (cl chatList) Swap(i, j int) {
	cl[i], cl[j] = cl[j], cl[i]
}

func (cl chatList) Less(i, j int) bool {
	return cl[i].Id < cl[j].Id
}

func loadOldChat(roomId, firstChatId int) (chatList, error) {
	resp, err := client.Get(baseURL + "/gateway.php?cmd=load_old_chat&myid=" + myid + "&_v=1.80a&_av=4&_t=" + token + "&ln=ja&room_id=" + strconv.Itoa(roomId) + "&first_chat_id=" + strconv.Itoa(firstChatId))
	if err != nil {
		return chatList{}, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	result := loadOldChatResult{}
	json.Unmarshal(body, &result)
	return result.Result.ChatList, nil
}

func GetRoomName(roomId int) (string, error) {
	room, ok := roomList[strconv.Itoa(roomId)]
	if !ok {
		return "", fmt.Errorf("you are not a member of the room(" + strconv.Itoa(roomId) + ")")
	}
	if len(room.Name) != 0 {
		return room.Name, nil
	}
	for _, contact := range contactList {
		if contact.RoomId == roomId {
			return contact.Name, nil
		}
	}
	return "", fmt.Errorf("no room name")
}

func Export(roomId int, file *os.File) error {
	writer := csv.NewWriter(file)
	firstChatId := 0
	for {
		chatList, err := loadOldChat(roomId, firstChatId)
		if err != nil {
			return err
		}
		// order by chat.Id desc
		sort.Sort(sort.Reverse(chatList))
		for _, chat := range chatList {
			name := contactList[strconv.Itoa(chat.AccountId)].Name
			if len(name) == 0 {
				name = strconv.Itoa(chat.AccountId)
			}
			// 0123456789,2006-01-02 15:04:05,NAME,MESSAGE
			writer.Write([]string{
				strconv.Itoa(chat.Id),
				time.Unix(int64(chat.Time), 0).Format("2006-01-02 15:04:05"),
				name,
				chat.Message,
			})
			firstChatId = chat.Id
		}
		writer.Flush()
		if len(chatList) < 40 {
			break
		}
		pause()
	}
	return nil
}

func pause() {
	rand.Seed(time.Now().UnixNano())
	time.Sleep(time.Duration(rand.Int63n(2000)+1000) * time.Millisecond)
}
