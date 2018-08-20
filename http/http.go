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

type SourceHandler struct {
	*httprouter.Router
	SourceService montaigne.SourceService
	Logger        *log.Logger
}

// NewDialHandler returns a new instance of DialHandler.
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

func (h *SourceHandler) Source(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	title := p.ByName("title")
	fmt.Println(title)
	s, err := h.SourceService.Source(title)
	if err != nil {
		Error(w, err, 404, h.Logger)
	} else if err := json.NewEncoder(w).Encode(s); err != nil {
		Error(w, err, 500, h.Logger)
	}
	w.Header().Set("Content-Type", "application/json")
}

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

func (h *SourceHandler) Create(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	err := r.ParseForm()
	if err != nil {
		Error(w, err, 500, h.Logger)
		return
	}
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
	content, err := parseHTML(z, &tempS, "", currentSubsection, "")
	if err != nil {
		Error(w, err, 500, h.Logger)
		return
	}
	content.ID = bson.NewObjectId()

	err = h.SourceService.Create(content)
	if err != nil {
		Error(w, err, 500, h.Logger)
		return
	}
	bytes, err := json.Marshal(content)
	if err != nil {
		panic(err)
	}
	_, _ = w.Write(bytes)
	w.Header().Set("Content-Type", "application/json")
	return
}

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

func parseHTML(z *html.Tokenizer, s *montaigne.Source, tokenType string, currentSubsection montaigne.Subsection, currentParagraph string) (*montaigne.Source, error) {
	//Tokentype
	tt := z.Next()
	//Token
	t := z.Token()
	r, err := regexp.Compile("h([1-9])")
	if err != nil {
		return nil, err
	}
	switch {
	case tt == html.ErrorToken:
		s.Subsections = append(s.Subsections, currentSubsection)
		return s, nil
	case tokenType == "" && t.Data != "p" && t.Data != "title" && !r.MatchString(t.Data):
	case tt == html.StartTagToken && (t.Data == "title" || t.Data == "p" || r.MatchString(t.Data)):
		tokenType = t.Data
	case tt == html.EndTagToken && (t.Data == "title" || t.Data == "p" || r.MatchString(t.Data)):
		if r.MatchString(t.Data) {
			currentSubsection.Paragraphs = append(currentSubsection.Paragraphs, currentParagraph)
		}
		tokenType = ""
	case tokenType == "title":
		s.Title = t.Data
	case r.MatchString(tokenType):
		s.Subsections = append(s.Subsections, currentSubsection)
		var temp montaigne.Subsection
		temp.Subtitle = t.Data
		currentSubsection = temp
	case tokenType == "p" && tt == html.TextToken:
		if currentSubsection.Subtitle == "" {
			var temp montaigne.Subsection
			temp.Subtitle = "1st subsection"
			currentSubsection = temp
		} else {
			currentParagraph += t.Data
		}
	}
	return parseHTML(z, s, tokenType, currentSubsection, currentParagraph)
}

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
