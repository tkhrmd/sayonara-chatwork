package chatwork

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"strconv"
)

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
	Utm int
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

func LoadOldChat(roomId, firstChatId int) chatList {
	resp, _ := client.Get(baseURL + "/gateway.php?cmd=load_old_chat&myid=" + myid + "&_v=1.80a&_av=4&_t=" + accessToken + "&ln=ja&room_id=" + strconv.Itoa(roomId) + "&first_chat_id=" + strconv.Itoa(firstChatId))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	result := loadOldChatResult{}
	json.Unmarshal(body, &result)
	sort.Sort(sort.Reverse(result.Result.ChatList))
	return result.Result.ChatList
}
