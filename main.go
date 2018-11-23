package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"github.com/vadimicus/FollowUnFollowTWBot/store"
	"github.com/dghubble/go-twitter/twitter"
	"time"
	"github.com/dghubble/oauth1"
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
			Address:"localhost:20002",
		},
	}

	//bot Bot
)

const (
	waitTime = 87 // in seconds
	frame = 200
)


func main() {

	var twClient *twitter.Client
	var creds Creds
	targetFollowers := make([]ToFollowUser,0)

	//Bot initialization
	bot, error := Init(&globalOpt)
	if error != nil{
		fmt.Println("Init Error:", error)
	}

	fmt.Println("Got Bot:", bot)

	//Parsing args
	args := os.Args

	// target people file path -t
	// credential file path -c

	argsLen := len(args)

	if argsLen <=1 {
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

		if len(credsFilePath) == 0{
			fmt.Errorf("please use correct arguments -t for target people, -c for credentrials")
			os.Exit(400)
		}

		if len(sourceFilePath) != 0{
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
				//Fill new users to database with exist check
				for _, followerRaw := range targetFollowers{
					var user store.User
					bot.userStore.GetUserById(followerRaw.UserId, &user)
					if user.UserID != followerRaw.UserId{
						user = store.User{Name:followerRaw.Name,UserID:followerRaw.UserId,Description:followerRaw.Description,Weight:followerRaw.Weight,Status:0,LastActionTime:time.Now().Unix() }
						bot.userStore.Insert(user)
					}
				}
			}
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



		//Here we are ready to start normal work of the people
		twClient = initTwitterClient(&creds)

		fmt.Println("Client Initialized Ok:", twClient)
		//fmt.Printf("TWITTER CLIENT pointer:", twClient)

		//
		//config := oauth1.NewConfig(creds.ConsumerKey, creds.ConsumerSecret)
		//token := oauth1.NewToken(creds.Token, creds.TokenSecret)
		//// http.Client will automatically authorize Requests
		//httpClient := config.Client(oauth1.NoContext, token)
		//
		//// twitter client
		//client := twitter.NewClient(httpClient)
		//
		//fmt.Printf("TWITTER CLIENT pointer:", client)
		//time.Sleep(10 * time.Second)
		//fmt.Println("Client Initialized Ok:", client)

		//all, err := bot.userStore.GetAllUsers()
		//fmt.Printf("ALL: %d , err:%s \n", len(all), err)
		//usersToFollow,err := bot.userStore.GetUsersToFollow()
		//fmt.Printf("USEROT F: %d , err:%s \n", len(usersToFollow), err)


		//StartFollowUnFollow(&twClient)
		//ALL IN ONE SHIT CODE


		var running bool
		running = true

		var allUsersCount, toFollowUsersCount, toUnFollowUsersCount, usersDoneCount int64



		usersToFollow, err := bot.userStore.GetUsersToFollow()
		if err != nil{
			fmt.Errorf("getToUnFollow err:", err)
		}
		usersToUnFollow,err := bot.userStore.GetUsersToUnFollow()

		allUsers, _ := bot.userStore.GetAllUsers()

		allUsersCount = int64(len(allUsers))
		toFollowUsersCount = int64(len(usersToFollow))
		toUnFollowUsersCount = int64(len(usersToUnFollow))
		usersDoneCount = allUsersCount - toFollowUsersCount - toUnFollowUsersCount


		fmt.Printf("Wokr need to be done for %d persents\n to follow %d \n to unfollow %d \n done: %d \n", (usersDoneCount/allUsersCount)*100, toFollowUsersCount, toUnFollowUsersCount, usersDoneCount)

		var followers int

		for running {

			//toUnfollow, _ := bot.userStore.GetUsersToUnFollow()
			//toFollow, _ := bot.userStore.GetUsersToFollow()
			pending, _ := bot.userStore.GetUsersByStatus(1)

			action := ChoseAction(len(usersToFollow), len(usersToUnFollow), followers, len(pending) )
			fmt.Printf("\n Status\n to follow: %d \n to unfollow: %d \n action: %d", len(usersToFollow), len(usersToUnFollow), action)
			if action == 1 {

				toFollowUser := usersToFollow[0]

				if Follow(&toFollowUser, twClient){
					toFollowUser.Status = 1
					toFollowUser.LastActionTime = time.Now().Unix()
					bot.userStore.Update(toFollowUser)
				}
				usersToFollow = usersToFollow[1:]

				time.Sleep(waitTime * time.Second)

			} else if action == 2{
				//toUnFollowArr,_ := bot.userStore.GetUsersToUnFollow()
				toUnfollow := usersToUnFollow[0]

				if UnFollow(&toUnfollow, twClient){
					toUnfollow.Status = 2
					toUnfollow.LastActionTime = time.Now().Unix()
					bot.userStore.Update(toUnfollow)
				}
				time.Sleep(waitTime * time.Second)
				usersToUnFollow = usersToUnFollow[1:]
			} else if action == 3 {
				//TODO implement wait time
			} else{
				running = false
			}
		}

		fmt.Printf("Work is done, enjoy it ;)")


	}

}

func StartFollowUnFollow(client *twitter.Client)  {

	//var stop bool
	//stop = false
	//
	//var allUsersCount, toFollowUsersCount, toUnFollowUsersCount, usersDoneCount int64
	//
	//
	//
	//usersToFollow,_ := bot.userStore.GetUsersToUnFollow()
	//usersToUnFollow,_ := bot.userStore.GetUsersToUnFollow()
	//
	//allUsers, _ := bot.userStore.GetAllUsers()
	//
	//allUsersCount = int64(len(allUsers))
	//toFollowUsersCount = int64(len(usersToFollow))
	//toUnFollowUsersCount = int64(len(usersToUnFollow))
	//usersDoneCount = allUsersCount - toFollowUsersCount - toUnFollowUsersCount
	//
	//
	//fmt.Printf("Wokr done for %d persents\n to follow %d \n to unfollow %d \n done: %d \n", (usersDoneCount/allUsersCount)*100, toFollowUsersCount, toUnFollowUsersCount, usersDoneCount)
	//
	//
	//for stop != true {
	//	action := ChoseAction()
	//
	//	if action == 1 {
	//
	//		toFollowArr,_ := bot.userStore.GetUsersToFollow()
	//		toFollow := toFollowArr[0]
	//		fmt.Print("To Follow: %d \n", len(toFollowArr))
	//		if Follow(&toFollow, client){
	//			toFollow.Status = 1
	//			toFollow.LastActionTime = time.Now().Unix()
	//			bot.userStore.Update(toFollow)
	//		}
	//		time.Sleep(waitTime * time.Second)
	//
	//	} else if action == 2{
	//		toUnFollowArr,_ := bot.userStore.GetUsersToUnFollow()
	//		toUnfollow := toUnFollowArr[0]
	//		fmt.Print("To UnFollow: %d \n", len(toUnFollowArr))
	//		if UnFollow(&toUnfollow, client){
	//			toUnfollow.Status = 2
	//			toUnfollow.LastActionTime = time.Now().Unix()
	//			bot.userStore.Update(toUnfollow)
	//		}
	//		time.Sleep(waitTime * time.Second)
	//
	//	} else{
	//		stop = true
	//	}
	//}

	fmt.Printf("Work is done, enjoy it ;)")

}

func ChoseAction(toFCount int, toUnFCount int, followersCount int, pendingCount int) int {
	//This method should return 1 to implement follow, 2 to implement unfollow, 3 to wait human factor,  0 - work is done

	if followersCount!= 0 && pendingCount/followersCount >= 2{
		if toUnFCount == 0{
			return 3
		} else {
			return 2
		}
	}


	if toUnFCount == 0{
		if toFCount > 0{
			return 1
		} else {
			return 0
		}
	} else if toUnFCount <= frame{
		if toFCount > 0 {
			return 1
		} else {
			return 2
		}
	}  else if toUnFCount > frame {
		return 2
	}
	return 0
}

func test(bot *Bot)  {

	//inDB,err := bot.userStore.GetAllUsers()
	allCleanPeople, err := bot.userStore.GetUsersToFollow()
	fmt.Println("Err:", err)
	toFollowArr := allCleanPeople[:20]

	for _, toFollow := range toFollowArr{
		//TODO make subscribtion
		toFollow.Status=1
		toFollow.LastActionTime = time.Now().Unix()

		bot.userStore.Update(toFollow)
	}

	fmt.Println("Went sleep")
	time.Sleep(30 * time.Second)

	fmt.Println("Getting people to unsubscribe")

	toUnFollow, err := bot.userStore.GetUsersToUnFollow()

	fmt.Println("People to unsubscribe 30 secs waited:", len(toUnFollow))

	fmt.Println("Went sleep")
	time.Sleep(30 * time.Second)

	fmt.Println("Getting people to unsubscribe")

	toUnFollowLast, err := bot.userStore.GetUsersToUnFollow()

	fmt.Println("People to unsubscribe 1 min :", len(toUnFollowLast))

	var unsubscibedIDS []int64

	for _, unsub := range toUnFollowLast{
		unsub.Status = 2
		unsub.LastActionTime = time.Now().Unix()
		bot.userStore.Update(unsub)
		unsubscibedIDS = append(unsubscibedIDS, unsub.UserID)
	}


	for _, check := range unsubscibedIDS{
		fmt.Println("GETTING UNSUBSCRIBERS FROM DB BY ID: ", check)
		var user store.User

		bot.userStore.GetUserById(check, &user)

		fmt.Println("GOT USER:", user)
	}
}

func initTwitterClient(creds *Creds) *twitter.Client{
	config := oauth1.NewConfig(creds.ConsumerKey, creds.ConsumerSecret)
	token := oauth1.NewToken(creds.Token, creds.TokenSecret)
	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// twitter client
	return twitter.NewClient(httpClient)
}


func UnFollow(user *store.User, client *twitter.Client) bool {

	//This is code to unFollow people

	params := twitter.FriendshipDestroyParams{ScreenName:user.Name,UserID:user.UserID}

	_, resp, err:=client.Friendships.Destroy(&params)

	if resp.StatusCode == 200{
		fmt.Printf("\nUnFollowed: %s id: %d \n", user.Name, user.UserID)
	} else{
		fmt.Printf("\nUnFollowed ERR: %s id: %d \n Error: %s", user.Name, user.UserID, err)
	}
	return true
}

func Follow(user *store.User, client *twitter.Client) bool {


	follow := true
	params := twitter.FriendshipCreateParams{UserID:user.UserID,ScreenName:user.Name,Follow:&follow}

	_, resp, err:=client.Friendships.Create(&params)

	if resp.StatusCode == 200{
		fmt.Printf("\nFollowed: %s id: %d \n", user.Name, user.UserID)
	} else{
		fmt.Printf("\nFollowed ERR: %s id: %d \n Error: %s", user.Name, user.UserID, err)
	}


	return true
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

