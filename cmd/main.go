package main

import (
	las "net/http"

	mgo "gopkg.in/mgo.v2"

	"github.com/madskrogh/montaigne/http"
	"github.com/madskrogh/montaigne/mongodb"
)

func main() {

	//Initiates new MongoDB session
	session, err := mgo.Dial("mongodb://localhost")
	if err != nil {
		panic(err)
	}

	//Creates a new SourceHandler
	h := http.NewSourceHandler()

	//Creates a new SourceService and stores
	//it in the SourceHandler
	h.SourceService = &mongodb.SourceService{MongoDB: session}

	las.ListenAndServe(":8080", h)

	//implement authentication
}
