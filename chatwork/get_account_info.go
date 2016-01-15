package chatwork

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"sort"
	"strconv"
)

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
	resp, _ := client.PostForm(baseURL+"/gateway.php?cmd=get_account_info&myid="+myid+"&_v=1.80a&_av=4&_t="+accessToken+"&ln=ja", url.Values{
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
