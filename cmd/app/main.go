package main

func main() {

	a := App{}

	//database info
	username := "user"
	password := "password"
	database := "maindatabase"
	host := "mysql"
	port := "3306"

	// init DB
	a.Init(username, password, host, port, database)

	// listen on port :3000
	a.Run(":3000")

}
