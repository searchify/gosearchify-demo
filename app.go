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

var templateDir string = os.Getenv("MYROOT")
var myTemplates *template.Template = nil

func init() {
	if templateDir == "" {
		templateDir = "."
	}
	fmt.Printf("Using root: \"%v\"\n", templateDir)
	//templates = template.Must(template.ParseFiles(templateDir + "/web/index.html"))
	loadTemplates()
}

func loadTemplates() {
	f := template.FuncMap {"maxlen": Maxlen, "fmttime": FormatTime}
	var err error
	myTemplates, err = template.New("name").Funcs(f).ParseFiles(templateDir + "/web/index.html")
	if err != nil {
		fmt.Printf("Error reading templates from %s: %v\n", templateDir, err)
		//panic(err)
	}
}

func FormatTime(tsval interface{}) string {
	ts := tsval.(string)
	const timeFormat = "Jan 2, 2006, 3:04 pm"
	timestamp, err := strconv.Atoi(ts)
	if err != nil {
		fmt.Printf("Problem converting time %s: %s", ts, err)
		return ""
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
	//apiUrl := "http://dbajo.api.searchify.com"
	apiUrl = os.Getenv("SEARCHIFY_API_URL")
	if apiUrl == "" {
		fmt.Fprintf(os.Stderr, "You need to set your SEARCHIFY_API_URL env variable first. (\"heroku config:set SEARCHIFY_API_URL=http://...\"\n")
		fmt.Fprintf(os.Stderr, "If you're on Heroku, add the Searchify add-on\n")
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
	//http.Handle("/static2/", wrapHandler("/static2/", http.FileServer(staticDir)))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(staticDir)))


	http.HandleFunc("/search", search)
	http.HandleFunc("/debug", debug)
	http.HandleFunc("/", hello)

	port := os.Getenv("PORT")
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func wrapHandler(prefix string, h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("r.URL = %s, prefix = %s\n", r.URL.Path, prefix)
		h.ServeHTTP(w, r)
		/*
		if !strings.HasPrefix(r.URL.Path, prefix) {
			NotFound(w, r)
			return
		}
		r.URL.Path = r.URL.Path[len(prefix):]
		h.ServeHTTP(w, r)
	*/
	})
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
	NextStart int
	PrevStart int
	HasPrev bool
	HasNext bool
}

func loadPage(title string) (*Page, error) {
	//filename := title + ".txt"
	//body, err := ioutil.ReadFile(filename)
//	if err != nil {
//		return nil, err
//	}
	return &Page{Title: title}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	// always reparse, in development mode
	loadTemplates()

	err := myTemplates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
	    http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	/*
	t, err := template.ParseFiles(tmpl + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} */
}

func search(w http.ResponseWriter, req *http.Request) {
	apiUrl := "http://dbajo.api.searchify.com"
	apiClient, err := indextank.NewApiClient(apiUrl)
	if err != nil {
		log.Fatalln("Error creating client:", err)
	}

	idx := apiClient.GetIndex("enron2")
//	if !idx.Exists() {
//		fmt.Fprintf(w, "Search index is missing, something is wrong!")
//		return
//	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	startParam := req.FormValue("start")
	start, err := strconv.Atoi(startParam)
	if err != nil {
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
			fmt.Printf("Error searching: %v\n", err)
			return
		}

		// compute pagination crap
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

		results := sr.GetResults()
		tmpl := true
		if !tmpl {
			fmt.Fprintf(w, "<html><body>")
			fmt.Fprintf(w, "<h2>Got %v matches in %v seconds</h2>\n", sr.GetMatches(), sr.GetSearchTime())
			for i, r := range results {
				fmt.Fprintf(w, "<b>Result %d: %s</b><br>%s<br><br>\n", (i+1), r["subject"], r["snippet_text"])
			}
			fmt.Fprintf(w, "</body></html>")
		} else {
			title := fmt.Sprintf("%s - Searchify using Go client", userQuery)
			p := &Page{
				Title: title, Query: userQuery, Results: results, Matches: sr.GetMatches(),
				SearchTime: sr.GetSearchTime(), DidYouMean: sr.GetDidYouMean(), Facets:sr.GetFacets(),
				NextStart:nextStart, PrevStart:prevStart, HasNext:next, HasPrev:prev }
			renderTemplate(w, "index", p)
		}
	} else {
		p := &Page{Title: "Searchify using Go IndexTank client library", Homepage: true}
		renderTemplate(w, "index", p)
	}
}

func makeQuery(s string) string {
	return fmt.Sprintf("subject:(%s)^6 OR text:(%s)", s, s)
}

func hello(w http.ResponseWriter, req *http.Request) {
	//fmt.Fprintln(w, "hello, world!")
	r := make([]map[string]interface{},0)
	p := &Page{Title: "Searchify using Go client library", Results: r, Homepage: true}
	renderTemplate(w, "index", p)
}

func debug(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Debug information\n\n")
	fmt.Fprintf(w, "os.Args: %s\n", os.Args)
	fmt.Fprintf(w, "Environment:\n")
	for _, v := range os.Environ() {
		fmt.Fprintf(w, "   %s\n", v)
	}
	fmt.Fprintf(w, "Your IP: %s\n", req.RemoteAddr)
}
