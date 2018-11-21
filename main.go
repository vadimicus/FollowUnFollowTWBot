package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"github.com/vadimicus/FollowUnFollowTWBot/store"
	"time"
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

type Bot struct {
	config     *Configuration
	//route      *gin.Engine

	userStore store.UserStore

	//restClient     *client.RestClient


}

type Configuration struct {
	Name              string
	Database          store.Conf
	//SocketioAddr      string
	//RestAddress       string

}

var (
	globalOpt = Configuration{
		Name:"FollowBot",
		Database: store.Conf{
			Address:"localhost:27017",
		},
	}
)


func main() {

	var creds Creds
	targetFollowers := make([]ToFollowUser,0)

	bot, error := Init(&globalOpt)
	if error != nil{
		fmt.Println("Init Error:", error)
	}

	fmt.Println("Got Bot:", bot)
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

			//Test fill database

			for _, followerRaw := range targetFollowers{

				var user store.User

				bot.userStore.GetUserById(followerRaw.UserId, &user)

				if user.UserID != followerRaw.UserId{
					user = store.User{Name:followerRaw.Name,UserID:followerRaw.UserId,Description:followerRaw.Description,Weight:followerRaw.Weight,Status:0,LastActionTime:time.Now().Unix() }

					bot.userStore.Insert(user)
				}

			}


		}

		inDB,err := bot.userStore.GetAllUsers()

		fmt.Println("Goted users from DB LEN:", len(inDB))
		fmt.Println("Goted users from DB ERR:", err)

	}

}

//func Init(conf *Configuration) (*Multy, error) {
func Init(conf *Configuration) (*Bot, error) {
	bot := &Bot{
		config: conf,
	}
	// DB initialization
	userStore, err := store.InitUserStore(conf.Database)
	if err != nil {
		return nil, fmt.Errorf("DB initialization: %s on port %s", err.Error(), conf.Database.Address)
	}
	bot.userStore = userStore
	fmt.Println("UserStore initialization done on %s √", conf.Database)


	//users data set

	//// REST handlers
	//if err = multy.initHttpRoutes(conf); err != nil {
	//	return nil, fmt.Errorf("Router initialization: %s", err.Error())
	//}
	return bot, nil
}

