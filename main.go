package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"github.com/vadimicus/FollowUnFollowTWBot/store"
	"github.com/vadimicus/FollowUnFollowTWBot/client"
	"github.com/dghubble/go-twitter/twitter"
	"time"
	"github.com/dghubble/oauth1"

	"github.com/gin-gonic/gin"

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
	route      *gin.Engine

	userStore store.UserStore

	restClient     *client.RestClient


}

type Configuration struct {
	Name              string
	Database          store.Conf
	DataBaseAddr      	string	`json:"dbAddr"`
	RestAddress       	string	`json:"restAddr"`
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

		fmt.Println("We should wait for Mongo Start 5 seconds")
		time.Sleep(5 * time.Second)

		fmt.Println("Trying initizlize Bot")

		bot, error := Init(&conf)
		if error != nil{
			fmt.Println("Init Error:", error)
		}


		//fmt.Println("Got BOT!:", bot)


		//Just twitter client initialization
		twClient := initTwitterClient(&conf)

		fmt.Println("Client Initialized Ok:", twClient)



		//TODO make it work
		//Here we are ready to start normal work of the people
		StartFollowUnFollow(twClient, bot, &conf)

	}

}


func StartFollowUnFollow(twClient *twitter.Client, bot *Bot, conf *Configuration)  {

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
			if UpdateFrame(twClient, conf){
				fmt.Println("There is too much users added. Trying to update Frame")
				time.Sleep(waitTime * time.Second)
			} else{
				fmt.Println("Update Frame Error. Gone to sleep for 3 minutes")
				time.Sleep(3 *time.Minute)
			}

		} else{
			running = false
		}
	}

	fmt.Printf("Work is done, enjoy it ;)")

	fmt.Printf("Work is done, enjoy it ;)")

}

func UpdateFrame(twClient *twitter.Client, conf *Configuration) bool{


	params := twitter.UserShowParams{ScreenName:conf.OwnAccount}
	//
	ownAcc, resp, error := twClient.Users.Show(&params)

	if resp.StatusCode == 200{
		friendsCount := float64(ownAcc.FollowersCount)

		fmt.Println("OLD FRAME :", frame)

		frame = int(friendsCount * 1.5)

		fmt.Println("NEW FRAME :", frame)
	} else{
		fmt.Printf("\nUpdate Frame Error ERR: \n Error: %s",  error)
		return false
	}

	return true

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
	if err = bot.initHttpRoutes(); err != nil {
		return nil, fmt.Errorf("Router initialization: %s", err.Error())
	}

	go bot.route.Run(conf.RestAddress)

	return bot, nil
}

func (bot *Bot) initHttpRoutes() error {
//func (bot *Bot) initHttpRoutes(conf *Configuration) error {
	router := gin.Default()
	bot.route = router
	gin.SetMode(gin.DebugMode)

	//f, err := os.OpenFile("../currencies/erc20tokens.json", os.O_RDONLY, os.FileMode(0644))
	//// f, err := os.OpenFile("/currencies/erc20tokens.json")
	//if err != nil {
	//	return err
	//}
	//
	//bs, err := ioutil.ReadAll(f)
	//if err != nil {
	//	return err
	//}
	//tokenList := store.VerifiedTokenList{}
	//_ = json.Unmarshal(bs, &tokenList)

	restClient, err := client.SetRestHandlers(
		bot.userStore,
		router,
		//conf.DonationAddresses,
		//multy.BTC,
		//multy.ETH,
		//conf.MultyVerison,
		//conf.Secretkey,
		//conf.MobileVersions,
		//tokenList,
		//conf.BrowserDefault,
		//multy.ExchangerFactory,
	)
	if err != nil {
		return err
	}
	bot.restClient = restClient

	//// socketIO server initialization. server -> mobile client
	//socketIORoute := router.Group("/socketio")
	//socketIOPool, err := client.SetSocketIOHandlers(multy.restClient, multy.BTC, multy.ETH, socketIORoute, conf.SocketioAddr, conf.NSQAddress, multy.userStore)
	//if err != nil {
	//	return err
	//}
	//multy.clientPool = socketIOPool
	//multy.ETH.WsServer = multy.clientPool.Server
	//multy.BTC.WsServer = multy.clientPool.Server

	//firebaseClient, err := client.InitFirebaseConn(&conf.Firebase, multy.route, conf.NSQAddress)
	//if err != nil {
	//	return err
	//}
	//multy.firebaseClient = firebaseClient

	return nil
}
