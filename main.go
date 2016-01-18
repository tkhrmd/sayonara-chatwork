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
	email    = flag.String("l", "", "Email Address")
	password = flag.String("p", "", "Password")
	roomId   = flag.Int("r", 0, "Room ID")
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage: `)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	cw.Login(*email, *password)
	workingDir, _ := os.Getwd()
	logDir := path.Join(workingDir, "chatwork_log")
	if _, err := os.Stat(logDir); err != nil {
		if err := os.Mkdir(logDir, 0775); err != nil {
			os.Exit(2)
		}
	}

	name := cw.GetRoomName(*roomId)
	filePath := path.Join(logDir, strconv.Itoa(*roomId)+"_"+name+".csv")
	file, _ := os.Create(filePath)
	cw.Export(*roomId, file)
	defer file.Close()
}
