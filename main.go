package main

import (
	"fmt"
	"io/ioutil"
        "time"
)

func main() {
	for {
		fmt.Println("crawler: loop started")
		dat, err := ioutil.ReadFile("/etc/ezgliding/conf.d/crawler.conf")
		fmt.Printf("%v\n", string(dat))
		if err != nil {
			fmt.Println("crawler: failed to read conf file")
		}
		dat, err = ioutil.ReadFile("/etc/ezgliding/secret.d/crawler.conf")
		if err != nil {
			fmt.Println("crawler: failed to read secrets file")
		}
		fmt.Printf("%v\n", string(dat))
		time.Sleep(10 * time.Second)
	}
}
