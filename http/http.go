package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/julienschmidt/httprouter"
	"github.com/madskrogh/montaigne"
	"golang.org/x/net/html"
	"gopkg.in/mgo.v2/bson"
)

//SourceHandler handles incomming requests via its httprouter.Router
//and interacts with the source database through its SourceService
type SourceHandler struct {
	*httprouter.Router
	SourceService montaigne.SourceService
	Logger        *log.Logger
}

// NewSourceHandler returns a new instance of SourceHandler.
func NewSourceHandler() *SourceHandler {
	h := &SourceHandler{
		Router: httprouter.New(),
		Logger: log.New(os.Stderr, "", log.LstdFlags),
	}
	h.GET("/api/source/:title", h.Source)
	h.GET("/api/titles/", h.Titles)
	h.GET("/api/sources/", h.Sources)
	h.POST("/api/source/", h.Create)
	h.DELETE("/api/source/:title", h.Delete)
	return h
}

//Source returns a JSON source from SourceService
//based on the URL title parameter
func (h *SourceHandler) Source(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	title := p.ByName("title")
	s, err := h.SourceService.Source(title)
	if err != nil {
		Error(w, err, 404, h.Logger)
	} else if err := json.NewEncoder(w).Encode(s); err != nil {
		Error(w, err, 500, h.Logger)
	}
	w.Header().Set("Content-Type", "application/json")
}

//Sources returns all sources from SourceService
func (h *SourceHandler) Sources(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	s, err := h.SourceService.Sources()
	if len(*s) < 1 {
		w.WriteHeader(404)
		return
	} else if err != nil {
		Error(w, err, 500, h.Logger)
		return
	} else if err := json.NewEncoder(w).Encode(s); err != nil {
		Error(w, err, 500, h.Logger)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	return
}

//Titles returns all source titles from SourceServicewhich
//afterwards can be used to retrieve specific sources via Source
func (h *SourceHandler) Titles(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	t, err := h.SourceService.Titles()
	fmt.Println(*t)
	if len(*t) < 1 {
		w.WriteHeader(404)
		return
	} else if err != nil {
		Error(w, err, 500, h.Logger)
		return
	} else if err := json.NewEncoder(w).Encode(t); err != nil {
		Error(w, err, 500, h.Logger)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	return
}

//Create adds a new source. Create takes the url provided in
//the request body,gets the content and calls. ParseHTML
//which parses and returns the content within <h> and <p>
//tags. The source is stored through SourceService.
func (h *SourceHandler) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := r.ParseForm()
	if err != nil {
		Error(w, err, 500, h.Logger)
		return
	}
	//Gets url from form and gets page content
	url := r.FormValue("url")
	resp, err := http.Get(url)
	if err != nil {
		Error(w, err, 400, h.Logger)
		return
	}
	b := resp.Body
	defer b.Close()

	currentSubsection := montaigne.Subsection{}
	tempS := montaigne.Source{URL: url}
	z := html.NewTokenizer(b)
	//Calls parseHTML which returns parsed HTML from url
	content, err := parseHTML(z, &tempS, "", currentSubsection, "")
	if err != nil {
		Error(w, err, 500, h.Logger)
		return
	}
	//Calls SourceService.Create which stores the new source
	content.ID = bson.NewObjectId()
	err = h.SourceService.Create(content)
	if err != nil {
		Error(w, err, 500, h.Logger)
		return
	}
	//Marshals source to JSON and sends it back to the client
	bytes, err := json.Marshal(content)
	if err != nil {
		panic(err)
	}
	_, _ = w.Write(bytes)
	w.Header().Set("Content-Type", "application/json")
	return
}

//Delete deletes a source that mathces the title provided
//as a request url parameter
func (h *SourceHandler) Delete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	url := p.ByName("title")
	err := h.SourceService.Delete(url)
	if err != nil {
		Error(w, err, 404, h.Logger)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	return
}

//parseHTML is a recursive function that parses web content,
//retrieves content within <h> and <p> tags and stores these
//as Source and Subsection obejcts. At each iteration, the
//function calls itself with a html.Tokenizer pointing at
//a spcific token in the content. The function proceeds to
//analyze and process this token and then iterates to the
//following token. The function exits when there are no
//tokens left.
func parseHTML(z *html.Tokenizer, s *montaigne.Source, tokenType string, currentSubsection montaigne.Subsection, currentParagraph string) (*montaigne.Source, error) {
	//Next tokentype
	tt := z.Next()
	//Next token
	t := z.Token()
	//Defines and compiles regex pattern that includes all <h> tags
	r, err := regexp.Compile("h([1-9])")
	if err != nil {
		return nil, err
	}
	//Switch evaluates next token, its value and type.
	switch {
	//html.ErrorToken means that we are at the end of source.
	//Return and exit function.
	case tt == html.ErrorToken:
		s.Subsections = append(s.Subsections, currentSubsection)
		return s, nil
	//Previous tokentype is undefined (i.e. we are not between <title>, <p>
	//or <h> tags), and the current tag isn't a <title>, <p> or <h> startag.
	//Do nothing a return.
	case tokenType == "" && t.Data != "p" && t.Data != "title" && !r.MatchString(t.Data):
	//Next token is of type starttoken and is either a <p>, <title> or <h> tag.
	//Set last token to next tokens value and return.
	case tt == html.StartTagToken && (t.Data == "title" || t.Data == "p" || r.MatchString(t.Data)):
		tokenType = t.Data
	//Next token is a endtoken and is either a <p>, <title> or <h> tag.
	//Appends the current subsection to the list of paragraphs and reset
	//tokentype.
	case tt == html.EndTagToken && (t.Data == "title" || t.Data == "p" || r.MatchString(t.Data)):
		if r.MatchString(t.Data) {
			currentSubsection.Paragraphs = append(currentSubsection.Paragraphs, currentParagraph)
		}
		tokenType = ""
	//Last token was <title>. Make the next token the title of the source.
	case tokenType == "title":
		s.Title = t.Data
	//Last token was a headline. Make a new subsection, append it to
	//the subsections of the source and make the next token its
	//subtitle.
	case r.MatchString(tokenType):
		s.Subsections = append(s.Subsections, currentSubsection)
		var temp montaigne.Subsection
		temp.Subtitle = t.Data
		currentSubsection = temp
	//Last token was <p> and next token is a text token. Add this text
	//to a new, or existing paragraph of the current subsection.
	case tokenType == "p" && tt == html.TextToken:
		//No subtitle present for this paragraph. Make one up instead.
		if currentSubsection.Subtitle == "" {
			var temp montaigne.Subsection
			temp.Subtitle = "1st subsection"
			currentSubsection = temp
			//Add text to current paragraph. s
		} else {
			currentParagraph += t.Data
		}
	}
	return parseHTML(z, s, tokenType, currentSubsection, currentParagraph)
}

//Utility error function.
func Error(w http.ResponseWriter, err error, code int, logger *log.Logger) {
	// Log error.
	logger.Printf("http error: %v (code=%d)", err, code)

	// Write generic error response.
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&errorResponse{Err: err.Error()})
}

// errorResponse is a generic response for sending a error.
type errorResponse struct {
	Err string `json:"err,omitempty"`
}
