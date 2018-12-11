package arxiv

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/tools/blog/atom"
)

// An EprintFeed is a feed retrieved from the arXiv API
type EprintsFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string      `xml:"title"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Link    []atom.Link `xml:"link"`
	Entry   []*Eprint   `xml:"entry"`
}

// A Category is a subject category for an e-print
type Category struct {
	Term string `xml:"term,attr"`
}

// An Author is an author of an e-print
type Author struct {
	Name string `xml:"name"`
}

// A Link is a link attached to an e-print
type Link struct {
	Rel   string `xml:"rel,attr,omitempty"`
	Href  string `xml:"href,attr"`
	Type  string `xml:"type,attr,omitempty"`
	Title string `xml:"title,attr,omitempty"`
}

// An Eprint is an e-print residing in arXiv.org
type Eprint struct {
	ID         string     `xml:"id"`
	Title      string     `xml:"title"`
	Authors    []Author   `xml:"author"`
	Links      []Link     `xml:"link"`
	Categories []Category `xml:"category"`
	Published  string     `xml:"published"`
	Updated    string     `xml:"updated"`
	Abstract   string     `xml:"summary"`
}

// EprintsService interacts with the e-print-related endpoints on arXiv's API
type EprintsService interface {
	Get(id string) (*Eprint, error)
	List(opt *EprintListOptions) ([]*Eprint, error)
}

var (
	// ErrEprintNotFound is a failure to find a specified e-print
	ErrEprintNotFound = errors.New("e-print not found")
)

// EprintListOptions specify how to retrieve a list of e-prints from the arXiv API
type EprintListOptions struct {
	Search string   `url:"search_query,omitempty"`
	IDList []string `url:"id_list,omitempty,comma"`
	QueryOptions
}

// A SearchOptions is a search query configuration
type SearchOptions struct {
	Title            string
	Author           string
	Abstract         string
	JournalReference string
	Category         string
	All              string
}

func (opt SearchOptions) String() string {
	s := ""

	if opt.All != "" {
		s = fmt.Sprintf("all:%s", opt.All)
	} else {
		m := make(map[string]string)
		m["ti"] = opt.Title
		m["au"] = opt.Author
		m["abs"] = opt.Abstract
		m["jr"] = opt.JournalReference
		m["cat"] = opt.Category

		q := []string{}
		for k, v := range m {
			if v != "" {
				q = append(q, fmt.Sprintf("%s:%s", k, v))
			}
		}

		s = strings.Join(q, " AND ")
	}

	return s
}

type eprintsService struct{ client *Client }

func (s *eprintsService) Get(id string) (*Eprint, error) {
	opt := &EprintListOptions{}
	opt.IDList = append(opt.IDList, id)

	url, err := s.client.url("query", opt)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	feed := EprintsFeed{}
	_, err = s.client.Do(req, &feed)
	if err != nil {
		return nil, err
	}

	eprints := feed.Entry
	if len(eprints) == 0 {
		return nil, ErrEprintNotFound
	}

	eprint := feed.Entry[0]

	return eprint, nil
}

func (s *eprintsService) List(opt *EprintListOptions) ([]*Eprint, error) {
	url, err := s.client.url("query", opt)
	if err != nil {
		return nil, err
	}

	fmt.Printf("URL: %s\n", url.String())

	req, err := s.client.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	feed := EprintsFeed{}
	_, err = s.client.Do(req, &feed)
	if err != nil {
		return nil, err
	}

	eprints := feed.Entry
	if len(eprints) == 0 {
		return nil, ErrEprintNotFound
	}

	return eprints, nil
}
