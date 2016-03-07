package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/models"
	"github.com/ksheedlo/ghviz/simulate"
)

type IndexParams struct {
	Owner string
	Repo  string
}

func ListStarCounts(gh *github.GithubClient) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		allStargazers, err := gh.ListStargazers(vars["owner"], vars["repo"])
		if err != nil {
			w.WriteHeader(err.Status)
			w.Write([]byte(fmt.Sprintf("%s\n", err.Message)))
			return
		}
		starEvents, decodeErr := models.StarEventsFromApi(allStargazers)
		if decodeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		jsonBlob, jsonErr := json.Marshal(simulate.StarCounts(starEvents))
		if jsonErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		w.Write(jsonBlob)
	}
}

func ListOpenIssuesAndPrs(gh *github.GithubClient) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		allIssues, err := gh.ListIssues(vars["owner"], vars["repo"])
		if err != nil {
			w.WriteHeader(err.Status)
			w.Write([]byte(fmt.Sprintf("%s\n", err.Message)))
			return
		}
		events, decodeErr := models.IssueEventsFromApi(allIssues)
		if decodeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		jsonBlob, jsonErr := json.Marshal(simulate.OpenIssueAndPrCounts(events))
		if jsonErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		w.Write(jsonBlob)
	}
}

var IndexTpl *template.Template = template.Must(template.ParseFiles("index.tpl.html"))

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	indexParams := IndexParams{Owner: os.Getenv("GHVIZ_OWNER"), Repo: os.Getenv("GHVIZ_REPO")}
	IndexTpl.Execute(w, indexParams)
}

func ServeStaticFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	http.ServeFile(w, r, path.Join("dashboard", vars["path"]))
}

func TopIssues(gh *github.GithubClient) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		allItems, httpErr := gh.ListTopIssues(vars["owner"], vars["repo"], 5)
		if httpErr != nil {
			log.Fatal(httpErr)
			w.WriteHeader(httpErr.Status)
			w.Write([]byte(fmt.Sprintf("%s\n", httpErr.Message)))
			return
		}
		jsonBlob, jsonErr := json.Marshal(allItems)
		if jsonErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		w.Write(jsonBlob)
	}
}

func TopPrs(gh *github.GithubClient) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		allItems, httpErr := gh.ListTopPrs(vars["owner"], vars["repo"], 5)
		if httpErr != nil {
			log.Fatal(httpErr)
			w.WriteHeader(httpErr.Status)
			w.Write([]byte(fmt.Sprintf("%s\n", httpErr.Message)))
			return
		}
		jsonBlob, jsonErr := json.Marshal(allItems)
		if jsonErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		w.Write(jsonBlob)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	r := mux.NewRouter()
	gh := github.NewClient(os.Getenv("GITHUB_USERNAME"), os.Getenv("GITHUB_PASSWORD"))
	r.HandleFunc("/", ServeIndex)
	r.HandleFunc("/dashboard/{path:.*}", ServeStaticFile)
	r.HandleFunc("/gh/{owner}/{repo}/star_counts", ListStarCounts(gh))
	r.HandleFunc("/gh/{owner}/{repo}/issue_counts", ListOpenIssuesAndPrs(gh))
	r.HandleFunc("/gh/{owner}/{repo}/top_issues", TopIssues(gh))
	r.HandleFunc("/gh/{owner}/{repo}/top_prs", TopPrs(gh))
	http.ListenAndServe(":4000", r)
}
