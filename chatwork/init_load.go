package chatwork

import (
	"encoding/json"
	"io/ioutil"
)

type initLoadResult struct {
	Result struct {
		Rooms    map[string]room    `json:"room_dat"`
		Contacts map[string]contact `json:"contact_dat"`
	}
}

type room struct {
	Mid int
	// R   int
	// Tp  int
	// C   int
	// F   int
	// T   int
	// Lt  int
	M map[string]int
}

type contact struct {
	Aid int
	Gid int
	Nm  string
	// cwid string
	// onm  string
	// dp   string
	// sp   string
	// fb   string
	// tw   string
	// ud   int
	// av   string
	// cv   string
	// name string
	// avid string
	// mrid string
	// rid  string
}

func initLoad() {
	resp, _ := client.Get(baseURL + "/gateway.php?cmd=init_load&myid=" + myid + "&_v=1.80a&_av=4&_t=" + accessToken + "&ln=ja&rid=0&type=&new=1")
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	result := initLoadResult{}
	json.Unmarshal(body, &result)
	rooms = result.Result.Rooms
	contacts = result.Result.Contacts
}
