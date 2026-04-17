// Package handlers defines HTTP endpoints for the kuang API.
package handlers

import (
	"strconv"

	"github.com/zoobz-io/kuang/types"
	"github.com/zoobz-io/kuang/contracts"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

// GitHub returns all GitHub API endpoints.
func GitHub() []rocco.Endpoint {
	return []rocco.Endpoint{
		listRepos(),
		getRepo(),
		listIssues(),
		getIssue(),
		createIssue(),
		listPRs(),
		getPR(),
		createPR(),
		getFile(),
		createOrUpdateFile(),
		searchCode(),
	}
}

func listRepos() rocco.Endpoint {
	return rocco.GET[rocco.NoBody, types.RepoList]("/github/repos", func(r *rocco.Request[rocco.NoBody]) (types.RepoList, error) {
		return sum.MustUse[contracts.GitHub](r).ListRepos(r)
	}).WithSummary("List repositories").WithTags("github")
}

func getRepo() rocco.Endpoint {
	return rocco.GET[rocco.NoBody, types.Repo]("/github/repos/{name}", func(r *rocco.Request[rocco.NoBody]) (types.Repo, error) {
		return sum.MustUse[contracts.GitHub](r).GetRepo(r, r.Params.Path["name"])
	}).WithSummary("Get repository").WithTags("github").WithPathParams("name")
}

func listIssues() rocco.Endpoint {
	return rocco.GET[rocco.NoBody, types.IssueList]("/github/repos/{repo}/issues", func(r *rocco.Request[rocco.NoBody]) (types.IssueList, error) {
		return sum.MustUse[contracts.GitHub](r).ListIssues(r, r.Params.Path["repo"])
	}).WithSummary("List issues").WithTags("github").WithPathParams("repo")
}

func getIssue() rocco.Endpoint {
	return rocco.GET[rocco.NoBody, types.Issue]("/github/repos/{repo}/issues/{number}", func(r *rocco.Request[rocco.NoBody]) (types.Issue, error) {
		num, err := strconv.Atoi(r.Params.Path["number"])
		if err != nil {
			return types.Issue{}, rocco.ErrBadRequest.WithMessage("invalid issue number")
		}
		return sum.MustUse[contracts.GitHub](r).GetIssue(r, r.Params.Path["repo"], num)
	}).WithSummary("Get issue").WithTags("github").WithPathParams("repo", "number")
}

func createIssue() rocco.Endpoint {
	return rocco.POST[types.CreateIssueRequest, types.Issue]("/github/repos/{repo}/issues", func(r *rocco.Request[types.CreateIssueRequest]) (types.Issue, error) {
		return sum.MustUse[contracts.GitHub](r).CreateIssue(r, r.Params.Path["repo"], r.Body.Title, r.Body.Body)
	}).WithSummary("Create issue").WithTags("github").WithPathParams("repo")
}

func listPRs() rocco.Endpoint {
	return rocco.GET[rocco.NoBody, types.PRList]("/github/repos/{repo}/pulls", func(r *rocco.Request[rocco.NoBody]) (types.PRList, error) {
		return sum.MustUse[contracts.GitHub](r).ListPRs(r, r.Params.Path["repo"])
	}).WithSummary("List pull requests").WithTags("github").WithPathParams("repo")
}

func getPR() rocco.Endpoint {
	return rocco.GET[rocco.NoBody, types.PullRequest]("/github/repos/{repo}/pulls/{number}", func(r *rocco.Request[rocco.NoBody]) (types.PullRequest, error) {
		num, err := strconv.Atoi(r.Params.Path["number"])
		if err != nil {
			return types.PullRequest{}, rocco.ErrBadRequest.WithMessage("invalid PR number")
		}
		return sum.MustUse[contracts.GitHub](r).GetPR(r, r.Params.Path["repo"], num)
	}).WithSummary("Get pull request").WithTags("github").WithPathParams("repo", "number")
}

func createPR() rocco.Endpoint {
	return rocco.POST[types.CreatePRRequest, types.PullRequest]("/github/repos/{repo}/pulls", func(r *rocco.Request[types.CreatePRRequest]) (types.PullRequest, error) {
		return sum.MustUse[contracts.GitHub](r).CreatePR(r, r.Params.Path["repo"], r.Body.Title, r.Body.Body, r.Body.Head, r.Body.Base)
	}).WithSummary("Create pull request").WithTags("github").WithPathParams("repo")
}

func getFile() rocco.Endpoint {
	return rocco.GET[rocco.NoBody, types.FileContent]("/github/repos/{repo}/content/{path...}", func(r *rocco.Request[rocco.NoBody]) (types.FileContent, error) {
		return sum.MustUse[contracts.GitHub](r).GetFile(r, r.Params.Path["repo"], r.Params.Path["path..."], r.Params.Query["ref"])
	}).WithSummary("Get file content").WithTags("github").WithPathParams("repo", "path...").WithQueryParams("ref")
}

func createOrUpdateFile() rocco.Endpoint {
	return rocco.PUT[types.CreateOrUpdateFileRequest, types.FileContent]("/github/repos/{repo}/content/{path...}", func(r *rocco.Request[types.CreateOrUpdateFileRequest]) (types.FileContent, error) {
		return sum.MustUse[contracts.GitHub](r).CreateOrUpdateFile(r, r.Params.Path["repo"], r.Params.Path["path..."], r.Body.Message, r.Body.Content, r.Body.SHA)
	}).WithSummary("Create or update file").WithTags("github").WithPathParams("repo", "path...")
}

func searchCode() rocco.Endpoint {
	return rocco.GET[rocco.NoBody, types.CodeSearchResult]("/github/search/code", func(r *rocco.Request[rocco.NoBody]) (types.CodeSearchResult, error) {
		return sum.MustUse[contracts.GitHub](r).SearchCode(r, r.Params.Query["query"])
	}).WithSummary("Search code").WithTags("github").WithQueryParams("query")
}
