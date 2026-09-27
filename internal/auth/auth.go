package auth

type Mongo struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}
