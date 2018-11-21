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
	GetUserById(user_id bson.M, user *User)
	GetUserByName(name bson.M, user *User)
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

func (mStore *MongoUserStore) GetUserById(user_id bson.M, user *User) {
	mStore.usersData.Find(user_id).One(user)
	return // why?
}

func (mStore *MongoUserStore) GetUserByName(name bson.M, user *User) {
	mStore.usersData.Find(name).One(user)
	return // why?
}

func (mStore *MongoUserStore) Update(sel, update bson.M) error {
	return mStore.usersData.Update(sel, update)
}

func (mStore *MongoUserStore) Insert(user User) error {
	return mStore.usersData.Insert(user)
}

func (mStore *MongoUserStore) Close() error {
	mStore.session.Close()
	return nil
}







type User struct {
	UserID  string   `bson:"user_id"`  // User uqnique identifier
	Name    string   `bson:"user_name"`
	Description string		`bson:"description"`
	Weight		int			`bson:"weight"`
	LastActionTime int64	`bson:"last_action_time"`
	Status			int		`bson:"status"`

}