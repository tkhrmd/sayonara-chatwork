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
	fmt.Fprintf(os.Stderr, `usage: sayonara-chatwork [option]

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
	fmt.Println(flag.NFlag())
	if flag.NArg() < 3 {
		flag.Usage()
		os.Exit(2)
	}

	err := cw.Login(*email, *password)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(2)
	}

	workingDir, _ := os.Getwd()
	logDir := path.Join(workingDir, "chatwork_log")
	if _, err := os.Stat(logDir); err != nil {
		if err := os.Mkdir(logDir, 0775); err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(2)
		}
	}

	name, _ := cw.GetRoomName(*roomId)

	filePath := path.Join(logDir, strconv.Itoa(*roomId)+"_"+name+".csv")
	file, _ := os.Create(filePath)
	defer file.Close()
	cw.Export(*roomId, file)
}
