package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
)

type RepoData struct {
	Name            string
	Description     string
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
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/traffic/views", owner, repo)
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
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/traffic/clones", owner, repo)
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
	tokenBytes, err := os.ReadFile("E:/Git/Secret_Token.txt")
	if err != nil {
		fmt.Printf("Error reading token file!!: %v\n", err)
		return
	}
	plsdontsteal := strings.TrimSpace(string(tokenBytes))
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
			repoData := RepoData{
				Name:  *repo.Name,
				Stars: *repo.StargazersCount,
				Forks: *repo.ForksCount,
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
			if err != nil {
				fmt.Printf("Error fetching commit counts for repository %s: %v\n", *repo.Name, err)
				continue
			}
			for _, contributor := range commitInfo {
				if *contributor.Author.Login == "GoEntity" {
					repoData.Commits = *contributor.Total
					repoData.CommitsIncrease = repoData.Commits - prevRepo.Commits
				}
			}
			views, err := getTrafficViews(plsdontsteal, "GoEntity", *repo.Name)
			if err != nil {
				fmt.Printf("Error fetching traffic views for repository %s: %v\n", *repo.Name, err)
				continue
			}
			repoData.Views = views
			repoData.ViewsIncrease = repoData.Views - prevRepo.Views
			clones, err := getTrafficClones(plsdontsteal, "GoEntity", *repo.Name)
			if err != nil {
				fmt.Printf("Error fetching traffic clones for repository %s: %v\n", *repo.Name, err)
				continue
			}
			repoData.Clones = clones
			repoData.ClonesIncrease = repoData.Clones - prevRepo.Clones
			blogData = append(blogData, repoData)
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
        <h1>GoEntity Public Repositories :)</h1>
		<h3>this only displays public repositories that I have ownership of</h3>
		<h3>if you are not seeing certain repos, it's because git server is returning 202 error</h3>
		<h3>which means it's still processing the data and should be available in a few minutes</h3>
		<h2>Updated on {{.Date}}</h2>
		<h2>Uploaded by <a href="https://github.com/GoEntity/GoEntity_Github">This Repo</a></h2>
		<h3>Please don't steal my git token :)</h3>
    </header>
    <main>
        <div class="grid">
            {{range .RepoData}}
            <article>
                <h2>{{.Name}}</h2>
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
</html>
	`
	t, _ := template.New("webpage").Parse(tmpl)
	pageData := PageData{
		Date:     time.Now().Format("2006-01-02 15:04:05"),
		RepoData: blogData,
	}
	f, _ := os.Create("index.html")
	err = t.Execute(f, pageData)
	if err != nil {
		fmt.Printf("Error executing template: %v\n", err)
		return
	}
	f.Close()
}
