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

type Controller struct {
	db          *sql.DB
	tiingoToken string
	Server      http.Server
}

func NewController(conn *sql.DB, tt string) Controller {
	c := Controller{
		db:          conn,
		tiingoToken: tt,
	}
	c.Server = http.Server{
		Addr:    ":3000",
		Handler: c.buildRoutes(),
	}
	return c
}

func (c *Controller) buildRoutes() http.Handler {
	r := http.NewServeMux()

	r.Handle("GET /static/", http.FileServer(http.FS(assets.StaticAssets)))

	r.HandleFunc("GET /favicon.ico", MakeHandler(c.favicon))
	r.HandleFunc("GET /annual", MakeHandler(c.annual))

	r.HandleFunc("GET /net-worth/new", MakeHandler(c.newNetWorthItem))
	r.HandleFunc("POST /net-worth/new", MakeHandler(c.createNetWorthItem))
	r.HandleFunc("GET /net-worth/{id}", MakeHandler(c.netWorthItem))
	r.HandleFunc("POST /net-worth/{id}/delete", MakeHandler(c.deleteNetWorthItem))
	r.HandleFunc("POST /net-worth/{id}", MakeHandler(c.updateNetWorthItem))
	r.HandleFunc("GET /net-worth", MakeHandler(c.netWorth))

	r.HandleFunc("GET /transactions/uncategorized", MakeHandler(c.uncategorizedTransactions))
	r.HandleFunc("GET /transactions/{id}", MakeHandler(c.transaction))
	r.HandleFunc("POST /transactions/{id}/delete", MakeHandler(c.deleteTransaction))
	r.HandleFunc("POST /transactions/{id}", MakeHandler(c.updateTransaction))

	r.HandleFunc("GET /categories/new", MakeHandler(c.newCategory))
	r.HandleFunc("POST /categories/new", MakeHandler(c.createCategory))
	r.HandleFunc("GET /categories/{id}", MakeHandler(c.category))
	r.HandleFunc("POST /categories/{id}/delete", MakeHandler(c.deleteCategory))
	r.HandleFunc("POST /categories/{id}", MakeHandler(c.updateCategory))
	r.HandleFunc("GET /categories", MakeHandler(c.categories))

	r.HandleFunc("GET /trades/new", MakeHandler(c.newTrade))
	r.HandleFunc("POST /trades/new", MakeHandler(c.createTrade))
	r.HandleFunc("POST /trades/{id}/delete", MakeHandler(c.deleteTrade))
	r.HandleFunc("GET /trades/{id}", MakeHandler(c.trade))
	r.HandleFunc("POST /trades/{id}", MakeHandler(c.updateTrade))
	r.HandleFunc("GET /trades", MakeHandler(c.trades))

	// this will match everything else so handle this in home handler
	r.HandleFunc("GET /", MakeHandler(c.transactions))

	return r
}

func (c *Controller) favicon(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type Base[T any] struct {
	Data T
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

func renderTemplate[T any](w http.ResponseWriter, data Base[T], name string, files []string) error {
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

				renderTemplate(w, Base[any]{}, "layout", []string{"not-found.html", "layout.html"})
			}
		}

		duration := time.Since(start)
		log.Printf("Handled request: uri=%s method=%s duration=%s", uri, method, duration)
	}
}
