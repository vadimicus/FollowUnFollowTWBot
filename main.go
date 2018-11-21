package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type Creds struct {
	ConsumerKey string		`json:"consKey"`
	ConsumerSecret string	`json:"consSec"`
	Token string			`json:"token"`
	TokenSecret string		`json:"tokenSec"`
}

type ToFollowUser struct {
	Name string			`json:"name"`
	Description string	`json:"description"`
	UserId int64		`json:"user_id"`
	Weight int			`json:"weight"`

}

func main() {

	var creds Creds
	targetFollowers := make([]ToFollowUser,0)

	args := os.Args

	// target people file path -t
	// credential file path -c

	argsLen := len(args)

	if argsLen <=2 {
		fmt.Println("Not correct arguments for launch")
	} else {

		var sourceFilePath string
		var credsFilePath string

		for index, arg := range args{
			if arg == "-t"{
				sourceFilePath = args[index+1]
			}
			if arg == "-c"{
				credsFilePath = args[index+1]
			}
		}

		if credsFilePath == ""{
			fmt.Printf("please use correct arguments -t for target people, -c for credentrials")
		}

		// read the whole file at once
		b, err := ioutil.ReadFile(credsFilePath)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(b, &creds)

		fmt.Println("Got JARR from file:",creds)

		fmt.Println("SOurce FIle Path:", sourceFilePath)
		fmt.Println("Creds FIle Path:", credsFilePath)

		// read the whole file at once
		if len(sourceFilePath)>0{
			sourceF, err := ioutil.ReadFile(sourceFilePath)
			if err != nil {
				panic(err)
			}
			json.Unmarshal(sourceF, &targetFollowers)
		}

		fmt.Println("Got JARR from file len:",len(targetFollowers))

		if len(targetFollowers) > 0{
			//TODO check and fill database
		}

	}
}
