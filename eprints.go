package arxiv

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/tools/blog/atom"
)

// An EprintsFeed is a feed retrieved from the arXiv API
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
	Published  time.Time  `xml:"published"`
	Updated    time.Time  `xml:"updated"`
	Abstract   string     `xml:"summary"`
}

func (e *Eprint) String() string {
	auths := []string{}
	for i := 0; i < len(e.Authors); i++ {
		auths = append(auths, e.Authors[i].Name)
	}
	s := fmt.Sprintf("Title: %s\n\nAuthors: %s\n\nAbstract: %s", e.Title, strings.Join(auths, ", "), e.Abstract)
	return s
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

	req, err := s.client.newRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	feed := EprintsFeed{}
	_, err = s.client.do(req, &feed)
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

	req, err := s.client.newRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	feed := EprintsFeed{}
	_, err = s.client.do(req, &feed)
	if err != nil {
		return nil, err
	}

	eprints := feed.Entry
	if len(eprints) == 0 {
		return nil, ErrEprintNotFound
	}

	return eprints, nil
}

// Subjects holds available e-print subjects
var Subjects = map[string]string{
	"physics": "Physics",
	"math":    "Mathematics",
	"cs":      "Computer Science",
	"q-bio":   "Quantitative Biology",
	"q-fin":   "Quantitative Finance",
	"stat":    "Statistics",
	"eess":    "Electrical Engineering and Systems Science",
	"econ":    "Economics",
}

// Subcategories holds e-print subcategories by subject
var Subcategories = map[string][]string{
	"physics": []string{
		"acc-ph", "ao-ph", "app-ph", "atm-clus", "atom-ph", "bio-ph", "chem-ph",
		"class-ph", "comp-ph", "data-an", "ed-ph", "flu-dyn", "gen-ph", "geo-ph",
		"hist-ph", "ins-det", "med-ph", "optics", "plasm-ph", "pop-ph", "soc-ph",
		"space-ph",
	},
	"math": []string{
		"AC", "AG", "AP", "AT", "CA", "CO", "CT", "CV", "DG", "DS", "FA", "GM",
		"GN", "GR", "GT", "HO", "IT", "KT", "LO", "MG", "MP", "NA", "NT", "OA",
		"OC", "PR", "QA", "RA", "RT", "SG", "SP", "ST",
	},
	"cs": []string{
		"AI", "AR", "CC", "CE", "CG", "CL", "CR", "CV", "CY", "DB", "DC", "DL",
		"DM", "DS", "ET", "FL", "GL", "GR", "GT", "HC", "IR", "IT", "LG", "LO",
		"MA", "MM", "MS", "NA", "NE", "NI", "OH", "OS", "PF", "PL", "RO", "SC",
		"SD", "SE", "SI", "SY",
	},
	"q-bio": []string{
		"BM", "CB", "GN", "MN", "NC", "OT", "PE", "QM", "SC", "TO",
	},
	"q-fin": []string{
		"CP", "EC", "GN", "MF", "PM", "PR", "RM", "ST", "TR",
	},
	"stat": []string{
		"AP", "CO", "ME", "ML", "OT", "TH",
	},
	"eess": []string{
		"AS", "IV", "SP",
	},
	"econ": []string{
		"EM",
	},
}
