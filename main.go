package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic"
	"github.com/pborman/uuid"
	"log"
	"net/http"
	"strconv"
)

const (
	DISTANCE   = "200km"
	POST_INDEX = "post"
	POST_TYPE  = "post"

	ES_URL = "http://104.198.39.242:9200/"
)

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Post struct {
	// `json:"user"` is for the json parsing of this User field. Otherwise, by default it's 'User'.
	User     string   `json:"user"`
	Message  string   `json:"message"`
	Location Location `json:"location"`
}

func main() {
	fmt.Println("started-service")
	createIndexIfNotExist()

	http.HandleFunc("/post", handlerPost)
	http.HandleFunc("/search", handlerSearch)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createIndexIfNotExist() {
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}

	exists, err := client.IndexExists(POST_INDEX).Do(context.Background())
	if err != nil {
		panic(err)
	}
	if !exists {
		mapping := `{
		"mappings": {
			"post": {
				"properties": {
					"location": {
						"type": "geo_point"
					}
				}
			}
		}
	}`
		_, err = client.CreateIndex(POST_INDEX).Body(mapping).Do(context.Background())
		if err != nil {
			panic(err)
		}
	}
}

// Save a post to ElasticSearch
func saveToES(post *Post, id string) error {
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		return err
	}
	_, err = client.Index().
		Index(POST_INDEX).
		Type(POST_TYPE).
		Id(id).
		BodyJson(post).
		Refresh("wait_for").
		Do(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("Post is saved to index: %s\n", post.Message)
	return nil
}

func handlerPost(w http.ResponseWriter, r *http.Request) {
	// Parse from body of request to get a json object.
	fmt.Println("Received one post request")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
	if r.Method == "OPTIONS" {
		return
	}

	decoder := json.NewDecoder(r.Body)
	var p Post
	if err := decoder.Decode(&p); err != nil {
		http.Error(w, "Cannot decode post data from client", http.StatusBadRequest)
		fmt.Printf("Cannot decode post data from client %v.\n", err)
		return
	}

	id := uuid.New()
	err := saveToES(&p, id)
	if err != nil {
		http.Error(w, "Failed to save post to ElasticSearch", http.StatusInternalServerError)
		fmt.Printf("Failed to save post to ElasticSearch %v.\n", err)
		return
	}
	fmt.Printf("Saved one post to ElasticSearch: %s", p.Message)
}

func handlerSearch(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one request for search")
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	// range is optional
	ran := DISTANCE
	if val := r.URL.Query().Get("range"); val != "" {
		ran = val + "km"
	}
	fmt.Println("range is ", ran)
	// Return a fake post
	p := &Post{
		User:    "1111",
		Message: "一生必去的100个地方",
		Location: Location{
			Lat: lat,
			Lon: lon,
		},
	}
	js, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
