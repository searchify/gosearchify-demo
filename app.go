package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"os"
	"github.com/searchify/gotank/indextank"
	"time"
	"strconv"
)

var templateDir string = ""
var myTemplates *template.Template = nil

func init() {
	templateDir = os.Getenv("MYROOT")
	if templateDir == "" {
		templateDir = "."
	}
	fmt.Printf("Using root: \"%v\"\n", templateDir)
	loadTemplates()
}

func loadTemplates() {
	f := template.FuncMap {"maxlen": Maxlen, "fmttime": FormatTime}
	var err error
	myTemplates, err = template.New("name").Funcs(f).ParseFiles(templateDir + "/web/index.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading templates from %s: %v\n", templateDir, err)
		//panic(err)
	}
}

func FormatTime(tsval interface{}) string {
	ts := tsval.(string)
	const timeFormat = "Jan 2, 2006, 3:04 pm"
	timestamp, err := strconv.Atoi(ts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Problem converting time %s: %s", ts, err)
		return "???"
	}
	tm := time.Unix(int64(timestamp), 0)
	return tm.Format(timeFormat)
}

func Maxlen(x interface{}, n int) (string) {
	item := x.(string)
	if len(item) <= n {
		return item
	}
	return item[:n]
}

var apiUrl string
var apiClient indextank.ApiClient

func main() {
	apiUrl = os.Getenv("SEARCHIFY_API_URL")
	if apiUrl == "" {
		fmt.Fprintf(os.Stderr, "You need to set your SEARCHIFY_API_URL env variable first.\n")
		fmt.Fprintf(os.Stderr, "If you're on Heroku, add the Searchify add-on - https://addons.heroku.com/searchify\n\n")
		fmt.Fprintf(os.Stderr, "Alternatively, you can explicitly set this with the following Heroku CLI command:\n")
		fmt.Fprintf(os.Stderr, "  heroku config:set SEARCHIFY_API_URL=http://:xxyyzz@myurl.api.searchify.com/\"\n")
		panic("No SEARCHIFY_API_URL")
	}
	var err error
	apiClient, err = indextank.NewApiClient(apiUrl)
	if err != nil {
		log.Fatalf("Error creating IndexTank client for API URL %s: %v\n", apiUrl, err)
		panic("Error creating client")
	}

	cwd, _ := os.Getwd()
	fmt.Printf("CWD = %s\n", cwd)
	dirName := "./web/static"

	// list static files, for debugging
	fileInfo, err := ioutil.ReadDir(dirName);
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading static files dir \"%s\": %v\n", dirName, err)
	}
	fmt.Println("Static files:")
	for _, v := range fileInfo {
		fmt.Printf("%v\n", v)
	}
	staticDir := http.Dir(dirName)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(staticDir)))

	http.HandleFunc("/search", search)
	http.HandleFunc("/", index)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

type Page struct {
	Title string
	Matches int64
	SearchTime float32
	Query string
	Results []map[string]interface{}
	DidYouMean string
	Facets map[string]map[string]int
	Homepage bool
	First int
	Last int
	NextStart int
	PrevStart int
	HasPrev bool
	HasNext bool
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	// always reparse, in development mode
	loadTemplates()

	err := myTemplates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func search(w http.ResponseWriter, req *http.Request) {
	apiClient, err := indextank.NewApiClient(apiUrl)
	if err != nil {
		log.Fatalln("Error creating client:", err)
	}

	idx := apiClient.GetIndex("enron")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	startParam := req.FormValue("start")
	start, err := strconv.Atoi(startParam)
	if err != nil || start < 0 || start > 5000 {
		start = 0
	}

	userQuery := req.FormValue("q")
	if &userQuery == nil || userQuery != "" {
		queryString := makeQuery(userQuery)
		query := indextank.QueryForString(queryString)
		query.FetchFields("subject", "text", "sender", "senderName", "timestamp", "messageid")
		query.SnippetFields("subject", "text")
		query.Start(start)
		fmt.Printf("%s searching: [%s]\n", req.RemoteAddr, userQuery)
		sr, err := idx.SearchWithQuery(query)
		if err != nil {
			fmt.Fprintf(w, "Error searching: %v\n", err)
			return
		}

		// compute pagination
		next, prev := false, false
		matches := int(sr.GetMatches())
		nextStart := start + 10
		if nextStart < matches {
			next = true
		}
		prevStart := 0
		if start >= 10 {
			prevStart = start - 10
			prev = true
		}
		first := start + 1
		last := start + 10
		if last > matches {
			last = matches
		}

		results := sr.GetResults()
		title := fmt.Sprintf("%s - Searchify using Gotank client", userQuery)
		p := &Page{
			Title: title, Query: userQuery, Results: results, Matches: sr.GetMatches(),
			SearchTime: sr.GetSearchTime(), DidYouMean: sr.GetDidYouMean(), Facets:sr.GetFacets(),
			First:first, Last:last, NextStart:nextStart, PrevStart:prevStart, HasNext:next, HasPrev:prev }
		renderTemplate(w, "index", p)
	} else {
		p := &Page{Title: "Searchify using Go IndexTank client library", Homepage: true}
		renderTemplate(w, "index", p)
	}
}

func makeQuery(s string) string {
	return fmt.Sprintf("subject:(%s)^6 OR text:(%s)", s, s)
}

func index(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("URL: %s\n", req.URL.Path)
	if req.URL.Path != "/" {
		http.Redirect(w, req, "/", 302)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	r := make([]map[string]interface{},0)
	p := &Page{Title: "Searchify using Go client library", Results: r, Homepage: true}
	renderTemplate(w, "index", p)
}
