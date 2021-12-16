package models

type AuthenticatedUser struct {
	User string `bson:"user"`
	DB   string `bson:"db"`
}

type AuthenticatedUserRole struct {
	Role string `bson:"role"`
	DB   string `bson:"db"`
}
