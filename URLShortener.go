package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/speps/go-hashids"
)

var db *bolt.DB
var port = 7777
var serverURL = "http://localhost:"
var addr = serverURL + strconv.Itoa(port)

//Post does something
type Post struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

//CreateShortURL will create a shortened URL from the given URL and store it in DB
type CreateShortURL struct {
	URL        string `json:"url"`
	CustomName string `json:"custom_name"`
}

//ShortURLData will contain data of shortened URL => used for lookups, etc
type ShortURLData struct {
	FullURL  string `json:"url"`
	ShortURL string `json:"custom_name"`
	Hits     int    `json:"hits"`
}

var posts []Post

func createShortURL(w http.ResponseWriter, r *http.Request) {
	var urlData CreateShortURL
	var dataForDB ShortURLData
	err := json.NewDecoder(r.Body).Decode(&urlData)
	if err != nil {
		//todo handle error case
	}
	hd := hashids.NewData()
	h, err := hashids.NewWithData(hd)
	if err != nil {
		//todo handle error case
	}
	now := time.Now()
	//assign data for DB entry
	dataForDB.ShortURL, err = h.Encode([]int{int(now.Unix())})
	if err != nil {
		//todo handle error case
	}
	if urlData.CustomName != "" {
		//TODO check if custom Name already used in DB
		dataForDB.ShortURL = urlData.CustomName
		log.Println("using custom short URL ", urlData.CustomName, " and db data set to ", dataForDB.ShortURL)
	}

	dataForDB.FullURL = urlData.URL
	log.Print(dataForDB)
	dbErr := db.Update(func(tx *bolt.Tx) error {
		URIList, err := tx.CreateBucketIfNotExists([]byte("uriList"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		enc, err := dataForDB.encode()
		if err != nil {
			return fmt.Errorf("could not encode Person %s: %s", dataForDB.ShortURL, err)
		}
		err = URIList.Put([]byte(dataForDB.ShortURL), enc)
		return err
	})
	if dbErr != nil {
		//todo handle error case
	}
	json.NewEncoder(w).Encode(&dataForDB)
}
func (p *ShortURLData) encode() ([]byte, error) {
	enc, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

//RootEndpoint does xyz
func RootEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Println("begin serving request: ", req.RequestURI)
	params := mux.Vars(req)
	log.Println(params)
	var fetchedData ShortURLData
	//get URL from DB lol
	dbErr := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("uriList"))
		v := bucket.Get([]byte(params["id"]))
		err := json.Unmarshal(v, &fetchedData)
		return err
	})
	if dbErr != nil {

	}
	log.Println("data from DB after unmarshal is ", fetchedData)
	log.Println("Full URL from DB after unmarshal is ", fetchedData.FullURL)

	//TODO handle what if fetched data is nil LOL
	http.Redirect(w, req, fetchedData.FullURL, http.StatusSeeOther)
	// http.Redirect(w, req, "http://"+fetchedData.FullURL, 301)
	log.Println("done redirect to ", fetchedData.FullURL)
}
func main() {
	var err error
	db, err = bolt.Open("shortURL.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {

	}
	log.Print("db init with path ", db.Path())

	router := mux.NewRouter()
	router.HandleFunc("/create", createShortURL).Methods("POST")
	router.HandleFunc("/{id}", RootEndpoint).Methods("GET")
	http.ListenAndServe(":7777", router)
}
