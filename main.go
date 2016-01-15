package main

import (
	"flag"
	"fmt"
	cw "github.com/maeda1991/sayonara-chatwork/chatwork"
	"os"
)

var (
	withFile = flag.Bool("f", false, "download files")
	email    = flag.String("l", "", "Email Address")
	password = flag.String("p", "", "Password")
	roomId   = flag.Int("r", 0, "Room ID")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: ")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	cw.Login(*email, *password)

	res := cw.LoadOldChat(*roomId, 0)

	fmt.Println(res)
}
