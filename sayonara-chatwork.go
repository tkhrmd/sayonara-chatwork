package main

import (
	"flag"
	"fmt"
	cw "github.com/maeda1991/sayonara-chatwork/chatwork"
	"os"
	"path"
	"strconv"
)

var (
	email    = flag.String("l", "", "login email address")
	password = flag.String("p", "", "login password")
	roomId   = flag.Int("r", 0, "chat room id")
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage: sayonara-chatwork [options]

  -l    login email address
  -p    login password
  -r    chat room id
          8digit number in the URL. as below
          https://www.chatwork.com/#!ridXXXXXXXX
                                        ¯¯¯¯¯¯¯¯
`)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NFlag() < 3 {
		flag.Usage()
		os.Exit(1)
	}

	err := cw.Login(*email, *password)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	roomName, err := cw.GetRoomName(*roomId)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	wd, _ := os.Getwd()
	dir := path.Join(wd, "chatwork_log")
	if _, err := os.Stat(dir); err != nil {
		if err := os.Mkdir(dir, 0775); err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	fileName := path.Join(dir, strconv.Itoa(*roomId)+"_"+roomName+".csv")
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer file.Close()

	err = cw.Export(*roomId, file)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		file.Close()
		os.Exit(1)
	}
}
