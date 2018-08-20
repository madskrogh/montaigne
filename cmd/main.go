package main

import (
	las "net/http"

	"github.com/madskrogh/montaigne/http"
	"github.com/madskrogh/montaigne/mongodb"

	mgo "gopkg.in/mgo.v2"
)

func main() {

	session, err := mgo.Dial("mongodb://localhost")
	if err != nil {
		panic(err)
	}

	h := http.NewSourceHandler()
	h.SourceService = &mongodb.SourceService{MongoDB: session}

	las.ListenAndServe(":8080", h)

	//implement authentication
}
