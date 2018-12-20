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
	"strconv"

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
	DataBaseAddr      	string	`json:"dbAddr"`
	//RestAddress       string
	ConsumerKey			string	`json:"consKey"`
	ConsumerSecret		string	`json:"consSec"`
	Token				string	`json:"token"`
	TokenSecret			string	`json:"tokenSec"`
	OwnAccount			string	`json:"ownAccount"`
	Frame				int64	`json:"frame"`
}

var (
	globalOpt = Configuration{
		Name:"FollowBot",
		Database: store.Conf{
			Address:"localhost:20002",
		},
	}

	frame = 350  //default value between diff of followers and friends
)

const (
	waitTime = 87 // in seconds to respect twitter API limits 1000 follow calls per 24 hours

)


func main() {

	//Bot initialization

	//bot, error := Init(&globalOpt)
	//if error != nil{
	//	fmt.Println("Init Error:", error)
	//}

	var bot *Bot

	//Getting args
	args := os.Args

	// target people file path -t
	// credential file path -c

	argsLen := len(args)

	if argsLen <=1 {
		fmt.Println("Not correct arguments for launch")
	} else {

		var conf Configuration
		//Just getting args from the input launch
		ParseArgs(args, &conf)

		//Bot initialization

		bot, error := Init(&conf)
		if error != nil{
			fmt.Println("Init Error:", error)
		}





		//Just twitter client initialization
		twClient := initTwitterClient(&conf)

		fmt.Println("Client Initialized Ok:", twClient)

		//TODO make it work
		//MakeLikes(twClient)
		//Here we are ready to start normal work of the people
		StartFollowUnFollow(twClient, bot)

	}

}


func StartFollowUnFollow(twClient *twitter.Client, bot *Bot)  {

	var running bool
	running = true

	var allUsersCount, toFollowUsersCount, toUnFollowUsersCount int

	var usersToUnFollow []store.User

	usersToFollow, err := bot.userStore.GetUsersToFollow()
	if err != nil{
		fmt.Errorf("getToUnFollow err:", err)
	}
	usersToUnFollow,_ = bot.userStore.GetUsersToUnFollow()

	allUsers, _ := bot.userStore.GetAllUsers()

	allUsersCount = len(allUsers)
	toFollowUsersCount = len(usersToFollow)
	toUnFollowUsersCount = len(usersToUnFollow)
	usersDone ,_ := bot.userStore.GetUsersByStatus(2)


	fmt.Printf("Users stats:\n all users %d \n to follow %d \n to unfollow %d \n done: %d \n", allUsersCount, toFollowUsersCount, toUnFollowUsersCount, len(usersDone))

	var actionsHistory = []int {1,1,1} //just default values

	for running {

		//if it will be bad use it
		pendingUsers, _ := bot.userStore.GetUsersByStatus(1)

		//This need to update cause it depends on the time
		usersToUnFollow,_ = bot.userStore.GetUsersToUnFollow()


		action := ChoseAction(len(usersToFollow), len(usersToUnFollow), len(pendingUsers), actionsHistory)
		actionsHistory = append(actionsHistory, action)
		//we need to check only last 3 actions
		actionsHistory = actionsHistory[1:]

		fmt.Printf("\n Status\n to follow: %d \n to unfollow: %d \n action: %d \n action history %v", len(usersToFollow), len(usersToUnFollow), action, actionsHistory)
		if action == 1 {

			toFollowUser := usersToFollow[0]

			if Follow(&toFollowUser, twClient){
				toFollowUser.Status = 1
				toFollowUser.LastActionTime = time.Now().Unix()
				bot.userStore.Update(toFollowUser)
			} else {
				//Something went wrong
				toFollowUser.Status = 66
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
			} else {
				//something went wrong
				//mark this user with error status
				toUnfollow.Status = 666
				toUnfollow.LastActionTime = time.Now().Unix()
				bot.userStore.Update(toUnfollow)
			}
			time.Sleep(waitTime * time.Second)
			usersToUnFollow = usersToUnFollow[1:]
		} else if action == 3 {
			//TODO go to make some likes

			fmt.Println("There is too much users added. Went sleep for 3 minutes")
			time.Sleep(3 *time.Minute)
		} else{
			running = false
		}
	}

	fmt.Printf("Work is done, enjoy it ;)")

	fmt.Printf("Work is done, enjoy it ;)")

}

func MakeLikes(client *twitter.Client) bool {

	//THIS IS GETTING LAST N tweets from ethereum and trying to like them
	params := twitter.SearchTweetParams{Query:"ethereum", TweetMode:"extanded", Count:4}
	search, resp, err := client.Search.Tweets(&params)




	if resp.StatusCode != 200 {
		return false
	}


	tweets:= search.Statuses
	for _, tweet := range tweets{
		fmt.Println("TW ID: ", tweet.ID)
		fmt.Println("USER: ", tweet.User)
		fmt.Println("FULL TEXT: ", tweet.FullText)
		fmt.Println("FAVORITE COUNT: ", tweet.FavoriteCount)
		fmt.Println("TEXT: ", tweet.Text)
		fmt.Println("IS FAVORITE: ", tweet.Favorited)
	}

	fmt.Println("Search:", search)
	//fmt.Println("Resp:", resp)
	fmt.Println("Err:", err)


	return false
}


//func ChoseAction(toFCount int, toUnFCount int, followersCount int, pendingCount int) int {
func ChoseAction(toFCount int, toUnFCount int, pendingCount int, actionHistory []int) int {
	//This method should return 1 to implement follow, 2 to implement unfollow, 3 to wait human factor,  0 - work is done


	var sum int
	for _, action := range actionHistory{
		sum+=action
	}


		//this first call of the function and there is no history
	if toUnFCount == 0{
		if toFCount > 0 && pendingCount < frame{
			return 1
		} else {
			if toFCount == 0 && pendingCount == 0{
				return 0
			}
			return 3
		}
	} else {
		if sum > 5{  //[2,1,1] //[1,2,1] [2,2,1] [2,2,2] [2,1,2] [1,1,2] [1,1,1]
			if toFCount > 0 {
				return 1
			} else {
				return 2
			}
		}
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

func initTwitterClient(creds *Configuration) *twitter.Client{
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
		return false
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
		return false
	}


	return true
}

func ParseArgs(args []string, conf *Configuration){
	//var sourceFilePath string
	var confFilePath string

	//targetFollowers := make([]ToFollowUser,0)

	for index, arg := range args{
		//if arg == "-t"{
		//	sourceFilePath = args[index+1]
		//}
		if arg == "-c"{
			confFilePath = args[index+1]
		}
		//if arg == "-f"{
		//	frame, _ = strconv.Atoi(args[index+1])
		//}
	}

	if len(confFilePath) == 0{
		fmt.Errorf("please use correct arguments -t for target people, -c for credentrials")
		os.Exit(400)
	}

	//if len(sourceFilePath) != 0{
	//	// read the whole file at once
	//	if len(sourceFilePath)>0{
	//		sourceF, err := ioutil.ReadFile(sourceFilePath)
	//		if err != nil {
	//			panic(err)
	//		}
	//		json.Unmarshal(sourceF, &targetFollowers)
	//	}
	//
	//	fmt.Println("Got JARR from file len:",len(targetFollowers))
	//
	//	if len(targetFollowers) > 0{
	//		//Fill new users to database with exist check
	//		for _, followerRaw := range targetFollowers{
	//			var user store.User
	//			bot.userStore.GetUserById(followerRaw.UserId, &user)
	//			if user.UserID != followerRaw.UserId{
	//				user = store.User{Name:followerRaw.Name,UserID:followerRaw.UserId,Description:followerRaw.Description,Weight:followerRaw.Weight,Status:0,LastActionTime:time.Now().Unix() }
	//				bot.userStore.Insert(user)
	//			}
	//		}
	//	}
	//}
	// read the whole file at once
	b, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		panic(err)
	}


	json.Unmarshal(b, &conf)

	conf.Database.Address = conf.DataBaseAddr

	fmt.Println("Got JARR from file:",conf)

	fmt.Println("SOurce FIle Path:", confFilePath)
	//fmt.Println("Creds FIle Path:", credsFilePath)
	fmt.Println("Frame updated to:", conf.Frame)
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

