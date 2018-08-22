package mongodb

import (
	"fmt"

	"github.com/madskrogh/montaigne"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//SourceService is an implementation of the montaigne.SourceService interface
type SourceService struct {
	MongoDB *mgo.Session
}

//Source returns the source that matches the given title
func (ss *SourceService) Source(title string) (*montaigne.Source, error) {
	source := montaigne.Source{}
	if err := ss.MongoDB.DB("montaigne").C("sources").Find(bson.M{"title": title}).One(&source); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &source, nil
}

//Sources return all sources
func (ss *SourceService) Sources() (*[]montaigne.Source, error) {
	sources := make([]montaigne.Source, 0)
	if err := ss.MongoDB.DB("montaigne").C("sources").Find(bson.M{}).All(&sources); err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &sources, nil
}

//Titles projects all source titles
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

//Create inserts a new source into the collection
func (ss *SourceService) Create(source *montaigne.Source) error {
	err := ss.MongoDB.DB("montaigne").C("sources").Insert(&source)
	return err
}

//Delete removes the first source that matches on title
func (ss *SourceService) Delete(title string) error {
	err := ss.MongoDB.DB("montaigne").C("sources").Remove(bson.M{"title": title})
	return err
}
