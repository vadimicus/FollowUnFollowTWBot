package store

import (

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
	"strconv"
)

const (
	TableUsers		= "UsersCollection"
)

type Conf struct {
	Address string
	DBUsers	string

	//RestoreState
	DBRestoreState string
	TableState     string
}

type UserStore interface {
	GetUserById(user_id int64, user *User)
	GetAllUsers()([]User, error)
	GetUsersToFollow()([]User, error)
	GetUserByName(name string, user *User)
	Update(user User) error
	Insert(user User) error
	Close() error
}

type MongoUserStore struct {
	config *Conf
	session *mgo.Session
	usersData *mgo.Collection

	RestoreState *mgo.Collection

}

func InitUserStore(conf Conf) (UserStore, error){
	uStore := &MongoUserStore{config: &conf,}

	addr := []string{"localhost:27017"}

	mongoDBDial := &mgo.DialInfo{
		Addrs:    addr,
		//Username: conf.Username,
		//Password: conf.Password,
	}

	session, err := mgo.DialWithInfo(mongoDBDial)
	if err != nil {
		return nil, err
	}


	uStore.session = session
	uStore.usersData = uStore.session.DB(conf.DBUsers).C(TableUsers)

	uStore.RestoreState = uStore.session.DB(conf.DBRestoreState).C(conf.TableState)

	return uStore, nil
}

func (mStore *MongoUserStore) GetUserById(user_id int64, user *User) {
	query := bson.M{"user_id": user_id}
	mStore.usersData.Find(query).One(user)
	return // why?
}

func (mStore *MongoUserStore) GetUserByName(name string, user *User) {
	query := bson.M{"user_name": name}
	mStore.usersData.Find(query).One(user)
	return // why?
}

func (mStore *MongoUserStore) Update(user User) error {
	sel := bson.M{"user_id":user.UserID}
	//update := bson.M{"$set": bson.M{"wallets.status": WalletStatusDeleted,},}
	update := bson.M{"$set": bson.M{"last_action_time": user.LastActionTime, "status": user.Status},}

	return mStore.usersData.Update(sel, update)
}

func (mStore *MongoUserStore) GetAllUsers()([]User, error) {
	allUsers := []User{}
	err:= mStore.usersData.Find(nil).All(&allUsers)
	return allUsers, err
}

func (mStore *MongoUserStore) GetUsersToFollow()([]User, error){
	users := []User{}
	query := bson.M{"status": 0}
	err := mStore.usersData.Find(query).All(&users)
	return users, err
}

func (mStore *MongoUserStore) GetUsersToUnFollow()([]User, error){
	users := []User{}
	var dateToUnffolow int64

	//dateToUnffolow = time.Now().Unix() - int64(3*24*time.Hour)
	dateToUnffolow = time.Now().Unix() - int64(1 * time.Minute)

	query := bson.M{"status": 1, "last_action_time":bson.M{ "$lt": dateToUnffolow}}
	err := mStore.usersData.Find(query).All(&users)


	return users, err
}

func (mStore *MongoUserStore) Insert(user User) error {
	return mStore.usersData.Insert(user)
}

func (mStore *MongoUserStore) Close() error {
	mStore.session.Close()
	return nil
}







type User struct {
	UserID  int64   `bson:"user_id"`  // User uqnique identifier
	Name    string   `bson:"user_name"`
	Description string		`bson:"description"`
	Weight		int			`bson:"weight"`
	LastActionTime int64	`bson:"last_action_time"`
	Status			int		`bson:"status"`

}