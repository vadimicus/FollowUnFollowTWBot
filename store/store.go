package store

import (

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	GetUserByName(name string, user *User)
	Update(sel, update bson.M) error
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

func (mStore *MongoUserStore) Update(sel, update bson.M) error {
	return mStore.usersData.Update(sel, update)
}

func (mStore *MongoUserStore) GetAllUsers()([]User, error) {
	allUsers := []User{}
	err:= mStore.usersData.Find(nil).All(&allUsers)
	return allUsers, err
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