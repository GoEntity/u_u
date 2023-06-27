package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
)

type RepoData struct {
	Name            string
	Description     string
	URL             string
	Stars           int
	StarsIncrease   int
	Forks           int
	ForksIncrease   int
	Commits         int
	CommitsIncrease int
	Views           int
	ViewsIncrease   int
	Clones          int
	ClonesIncrease  int
}

type TrafficData struct {
	Count int `json:"count"`
}

type PreviousData struct {
	RepoStats []RepoData `json:"repoStats"`
}

type PageData struct {
	Date     string
	RepoData []RepoData
}

func getTrafficViews(token, owner, repo string) (int, error) {
	timestamp := time.Now().Unix()
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/traffic/views?%d", owner, repo, timestamp)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data TrafficData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, err
	}
	return data.Count, nil
}

func getTrafficClones(token, owner, repo string) (int, error) {
	timestamp := time.Now().Unix()
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/traffic/clones?%d", owner, repo, timestamp)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data TrafficData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, err
	}
	return data.Count, nil
}

func main() {
	plsdontsteal := os.Getenv("GoEntity_Github")
	if plsdontsteal == "" {
		fmt.Println("token GoEntity_Github not set... attempting to read from local folder")
		tokenBytes, err := os.ReadFile("E:/Git/Secret_Token.txt")
		if err != nil {
			fmt.Printf("Error reading local token file!: %v\n", err)
			return
		}
		plsdontsteal = strings.TrimSpace(string(tokenBytes))
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: plsdontsteal},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	opt := &github.RepositoryListOptions{
		Visibility:  "public",
		Affiliation: "owner",
	}
	repos, _, err := client.Repositories.List(context.Background(), "", opt)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	sort.SliceStable(repos, func(i, j int) bool {
		return *repos[i].StargazersCount > *repos[j].StargazersCount
	})
	var prevData PreviousData
	prevDataBytes, _ := os.ReadFile("previous.json")
	json.Unmarshal(prevDataBytes, &prevData)
	var blogData []RepoData
	for _, repo := range repos {
		if !*repo.Fork {
			var repoData RepoData
			for i := 0; i < 15; i++ {
				repoData = RepoData{
					Name:  *repo.Name,
					Stars: *repo.StargazersCount,
					Forks: *repo.ForksCount,
					URL:   *repo.HTMLURL,
				}
				var prevRepo RepoData
				for _, pr := range prevData.RepoStats {
					if pr.Name == repoData.Name {
						prevRepo = pr
						break
					}
				}
				repoData.StarsIncrease = repoData.Stars - prevRepo.Stars
				repoData.ForksIncrease = repoData.Forks - prevRepo.Forks
				if repo.Description != nil {
					repoData.Description = *repo.Description
				}
				commitInfo, _, err := client.Repositories.ListContributorsStats(ctx, "GoEntity", *repo.Name)
				if err == nil {
					for _, contributor := range commitInfo {
						if *contributor.Author.Login == "GoEntity" {
							repoData.Commits = *contributor.Total
							repoData.CommitsIncrease = repoData.Commits - prevRepo.Commits
						}
					}
					views, err := getTrafficViews(plsdontsteal, "GoEntity", *repo.Name)
					if err == nil {
						repoData.Views = views
						repoData.ViewsIncrease = repoData.Views - prevRepo.Views
						clones, err := getTrafficClones(plsdontsteal, "GoEntity", *repo.Name)
						if err == nil {
							repoData.Clones = clones
							repoData.ClonesIncrease = repoData.Clones - prevRepo.Clones
							blogData = append(blogData, repoData)
							break
						}
					}
				}
				time.Sleep(time.Second * 2)
			}
		}
	}
	if len(blogData) > 0 {
		newDataBytes, _ := json.Marshal(PreviousData{RepoStats: blogData})
		os.WriteFile("previous.json", newDataBytes, 0644)
	} else {
		fmt.Println("No valid data fetched, not updating previous.json")
	}

	const tmpl = `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>GoEntity</title>
		<link rel="stylesheet" href="style.css">
	</head>
	<body>
		<header>
			<h1>GoEntity Public Repositories :&#41;</h1>
			<br>
			<h4>&gt;&gt; this only displays public repositories that I have ownership of</h4>
			<h4>&gt;&gt; if numbers aren't accurate, it's temporary due to 202 error, which should be automatically fixed in a few minutes</h4>
			<br>
			<h5>Updated on {{.Date}} via repo <a href="https://github.com/GoEntity/GoEntity_Github">GoEntity_Github</a></h5>
			<h5 style="color:red";>Please don't steal my git token :&#41;</h5>
		</header>
		<main>
			<div id="exp">
				<h3>*** shows public repo stats in the past <em>14</em> days with +/- count updates in the past <em>1</em> hour(s) ***<h3>
				<h5>git action to update the stats is currently set up to run once every hour</h5>
				<h5>but sometimes I might run it manually if I'm bored</h5>
			</div>
			<div class="grid">
				{{range .RepoData}}
				<article>
					<h2><a href="{{.URL}}">{{.Name}}</a></h2>
					<p>{{.Description}}</p>
					<p><strong>Stars:</strong> {{.Stars}} <span>(+{{.StarsIncrease}})</span></p>
					<p><strong>Forks:</strong> {{.Forks}} <span>(+{{.ForksIncrease}})</span></p>
					<p><strong>Commits:</strong> {{.Commits}} <span>(+{{.CommitsIncrease}})</span></p>
					<p><strong>Views:</strong> {{.Views}} <span>(+{{.ViewsIncrease}})</span></p>
					<p><strong>Clones:</strong> {{.Clones}} <span>(+{{.ClonesIncrease}})</span></p>
				</article>
				{{end}}
			</div>
		</main>
	</body>
	</html>`

	tmplParsed, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		log.Fatalf("Error parsing template: %v\n", err)
	}
	f, err := os.Create("index.html")
	if err != nil {
		log.Fatalf("Error creating index.html: %v\n", err)
	}
	data := &PageData{
		Date:     time.Now().Format("01-02-2006"),
		RepoData: blogData,
	}
	err = tmplParsed.Execute(f, data)
	if err != nil {
		log.Fatalf("Error executing template: %v\n", err)
	}
	err = f.Close()
	if err != nil {
		log.Fatalf("Error closing index.html: %v\n", err)
	}

	gitAdd := exec.Command("git", "add", ".")
	gitCommit := exec.Command("git", "commit", "-m", "::auto commit")

	err = gitAdd.Run()
	if err != nil {
		log.Fatalf("git add failed: %s", err)
	}

	err = gitCommit.Run()
	if err != nil {
		log.Printf("git commit failed: %s", err)
	}

	gitPush := exec.Command("git", "push")
	err = gitPush.Run()
	if err != nil {
		log.Fatalf("git push failed: %s", err)
	}
}
