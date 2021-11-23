package main

import (
	"heroku/todos/myapp"
	"log"
	"net/http"
	"os"
)

// const portNumber = ":3000"

func main() {
	port := os.Getenv("PORT")

	mux := myapp.MakeNewHandler("./todo.db")
	defer mux.Close()

	log.Println("Started App, portNo:", port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		panic(err)
	}
}
