package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname string) {
	connectionString := fmt.Sprintf("%s:%s@/%s", user, password, dbname)

	fmt.Println(connectionString)

	var err error
	a.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
	a.initUrlRoutes()
}
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("APP_PORT")), a.Router))
}

func (a *App) initializeRoutes() {
	// a.Router.HandleFunc("/products", a.getProducts).Methods("GET")
	// a.Router.HandleFunc("/product", a.createProduct).Methods("POST")
	// a.Router.HandleFunc("/product/{id:[0-9]+}", a.getProduct).Methods("GET")
	// a.Router.HandleFunc("/product/{id:[0-9]+}", a.updateProduct).Methods("PUT")
	// a.Router.HandleFunc("/product/{id:[0-9]+}", a.deleteProduct).Methods("DELETE")
}

func (a *App) initUrlRoutes() {
	a.Router.HandleFunc("/api/url", a.createShortUrl).Methods("POST")
	a.Router.HandleFunc("/add", a.addUrlViewHandler).Methods("GET")
	a.Router.HandleFunc("/add", a.addUrlHandler).Methods("POST")
	a.Router.HandleFunc("/", a.indexHandler).Methods("GET")
	a.Router.HandleFunc("/stats", a.statsHandler).Methods("GET")
	a.Router.HandleFunc(`/{fragment:[a-zA-Z0-9=\-\/]+}`, a.forwardUrl).Methods("GET")
}

func (a *App) statsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := getAggregateStats(a.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	t, _ := template.ParseFiles("stats.html")
	t.Execute(w, stats)
}

func (a *App) addUrlViewHandler(w http.ResponseWriter, r *http.Request) {
	renderView("add_url.html", w, r)
}

func (a *App) addUrlHandler(w http.ResponseWriter, r *http.Request) {
	url := &short_url{Destination: r.FormValue("destination"), Shortcode: r.FormValue("shortcode")}
	err := url.createShortUrl(a.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	urls, err := getShortUrls(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, urls)
}

func (a *App) forwardUrl(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var s short_url

	s.Shortcode = vars["fragment"]

	if err := s.getShortUrl(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.emitStat(s.Shortcode)
	http.Redirect(w, r, s.Destination, 307)
}

func (a *App) emitStat(shortcode string) {
	err := emitStat(a.DB, shortcode)
	if err != nil {
		log.Print("error in emitting stat %v ", err.Error())
	}
}

func (a *App) createShortUrl(w http.ResponseWriter, r *http.Request) {
	var s short_url
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload.")
		return
	}
	defer r.Body.Close()

	if err := s.createShortUrl(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	fmt.Print(s)

	respondWithJSON(w, http.StatusCreated, s)
}

func (a *App) getProducts(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	products, err := getProducts(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, products)
}

func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := p.createProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var p product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	defer r.Body.Close()

	p.ID = id

	if err := p.updateProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := product{ID: id}
	if err := p.deleteProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "sucess"})
}

func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := product{ID: id}
	if err := p.getProduct(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Product not found")
		default:
			respondWithError(w, http.StatusInternalServerError, "Some error")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func renderView(filename string, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles(filename)
	t.Execute(w, nil)
}
