package montaigne

import "gopkg.in/mgo.v2/bson"

type Source struct {
	ID          bson.ObjectId `json:"id" bson:"_id"`
	URL         string        `json:"url" bson:"url"`
	Title       string        `json:"title" bson:"title"`
	Subsections []Subsection  `json:"subsections" bson:"subsections"`
}

type Subsection struct {
	Subtitle   string   `json:"subtitle" bson:"subtitle"`
	Paragraphs []string `json:"paragraphs" bson:"paragraphs"`
}

type SourceService interface {
	Source(title string) (*Source, error)
	Sources() (*[]Source, error)
	Titles() (*[]string, error)
	Create(source *Source) error
	Delete(title string) error
}
