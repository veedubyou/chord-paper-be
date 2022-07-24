package userstorage

const (
	idKey = "id"
)

type dbUser struct {
	ID    string `dynamo:"id"`
	Name  string `dynamo:"username"`
	Email string `dynamo:"email"`
}
