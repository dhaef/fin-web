package controller

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"

	"fin-web/internal/assets"
	"fin-web/internal/templates"
)

var (
	transactionsDBConn *sql.DB
	netWorthDBConn     *sql.DB
)

type Controller struct {
	Server http.Server
}

func NewController(transactionsConn *sql.DB, netWorthConn *sql.DB) Controller {
	transactionsDBConn = transactionsConn
	netWorthDBConn = netWorthConn
	return Controller{
		Server: http.Server{
			Addr:    ":3000",
			Handler: buildRoutes(),
		},
	}
}

func buildRoutes() http.Handler {
	r := http.NewServeMux()

	r.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(assets.StaticAssets))))
	// r.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("/Users/derekheafner/dev/go/fin-web/static"))))

	r.HandleFunc("GET /favicon.ico", MakeHandler(favicon))
	r.HandleFunc("GET /annual", MakeHandler(annual))
	r.HandleFunc("GET /net-worth/new", MakeHandler(newNetWorthItem))
	r.HandleFunc("POST /net-worth/new", MakeHandler(createNetWorthItem))
	r.HandleFunc("GET /net-worth/{id}", MakeHandler(netWorthItem))
	r.HandleFunc("POST /net-worth/{id}", MakeHandler(updateNetWorthItem))
	r.HandleFunc("GET /net-worth", MakeHandler(netWorth))
	r.HandleFunc("GET /transactions/uncategorized", MakeHandler(uncategorizedTransactions))
	r.HandleFunc("GET /transactions/{id}", MakeHandler(transaction))
	r.HandleFunc("POST /transactions/{id}", MakeHandler(updateTransaction))
	// this will match everything else so handle this in home handler
	r.HandleFunc("GET /", MakeHandler(transactions))

	return r
}

func favicon(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type Base struct {
	Data any
}

func buildTemplatePaths(files []string) []string {
	templatesPath := path.Join("files")

	for index, file := range files {
		files[index] = path.Join(templatesPath, file)
	}

	return files
}

func handleTemplateFiles(files []string) (*template.Template, error) {
	filesWithFullPath := buildTemplatePaths(files)

	return template.ParseFS(templates.Templates, filesWithFullPath...)
}

// func render(w http.ResponseWriter, data any, files []string) error {
// 	t, err := handleTemplateFiles(files)
// 	if err != nil {
// 		return err
// 	}
// 	return t.Execute(w, data)
// }

func renderTemplate(w http.ResponseWriter, data any, name string, files []string) error {
	t, err := handleTemplateFiles(files)
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, name, data)
}

type APIError struct {
	Status       int
	Message      string
	ResponseType string
}

func (e APIError) Error() string {
	return e.Message
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func MakeHandler(h apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// log some data
		uri := r.RequestURI
		method := r.Method
		log.Printf("Incoming request: uri=%s method=%s", uri, method)

		if err := h(w, r); err != nil {
			if e, ok := err.(APIError); ok {
				log.Println(e.Error())

				if e.ResponseType == "JSON" {
					encode(w, r, e.Status, map[string]string{"message": e.Error()})
					return
				}

				renderTemplate(w, "", "layout", []string{"not-found.html", "layout.html"})
			}
		}

		duration := time.Since(start)
		log.Printf("Handled request: uri=%s method=%s duration=%s", uri, method, duration)
	}
}
