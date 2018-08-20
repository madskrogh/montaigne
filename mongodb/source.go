package mongodb

import (
	"fmt"

	"github.com/madskrogh/montaigne"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//When creating mongodb, use schema validation to ensure only good data is accepted
//https://docs.mongodb.com/manual/core/schema-validation/

type SourceService struct {
	MongoDB *mgo.Session
}

func (ss *SourceService) Source(title string) (*montaigne.Source, error) {
	source := montaigne.Source{}
	if err := ss.MongoDB.DB("montaigne").C("sources").Find(bson.M{"title": title}).One(&source); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &source, nil
}

func (ss *SourceService) Sources() (*[]montaigne.Source, error) {
	sources := make([]montaigne.Source, 0)
	if err := ss.MongoDB.DB("montaigne").C("sources").Find(bson.M{}).All(&sources); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &sources, nil
}

func (ss *SourceService) Titles() (*[]string, error) {
	sources := []montaigne.Source{}
	if err := ss.MongoDB.DB("montaigne").C("sources").Find(bson.M{}).Select(bson.M{"_id": 0, "title": 1}).All(&sources); err != nil {
		fmt.Println(err)
		return nil, err
	}
	titles := make([]string, 0)
	for s, _ := range sources {
		titles = append(titles, sources[s].Title)
	}
	return &titles, nil
}

func (ss *SourceService) Create(source *montaigne.Source) error {
	err := ss.MongoDB.DB("montaigne").C("sources").Insert(&source)
	return err
}

func (ss *SourceService) Delete(title string) error {
	err := ss.MongoDB.DB("montaigne").C("sources").Remove(bson.M{"title": title})
	return err
}
