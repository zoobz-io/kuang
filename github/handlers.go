package github

import (
	"context"
	"strconv"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/kuang/contracts"
	"github.com/zoobz-io/kuang/httpc"
	"github.com/zoobz-io/kuang/internal/auth"
)

// CredentialKey is the credential key used to resolve GitHub tokens.
const CredentialKey = "github-token"

// Scope constants for the GitHub module. These are used in kuang.yaml to
// grant agents access to specific operations. Patterns like "github-*-read"
// match all read scopes via the CLI's glob matching.
const (
	ScopeReposRead    = "github-repos-read"
	ScopeIssuesRead   = "github-issues-read"
	ScopeIssuesWrite  = "github-issues-write"
	ScopePullsRead    = "github-pulls-read"
	ScopePullsWrite   = "github-pulls-write"
	ScopeContentRead  = "github-content-read"
	ScopeContentWrite = "github-content-write"
	ScopeSearchRead   = "github-search-read"
)

// Scopes returns all scope strings defined by the GitHub module.
// Use this to set the server's permission ceiling in kuang.yaml or tests.
func Scopes() []string {
	return []string{
		ScopeReposRead, ScopeIssuesRead, ScopeIssuesWrite,
		ScopePullsRead, ScopePullsWrite, ScopeContentRead,
		ScopeContentWrite, ScopeSearchRead,
	}
}

// resolveAuth resolves the agent's GitHub bearer token from the credential store.
func resolveAuth(ctx context.Context) (httpc.RequestOption, error) {
	creds, err := sum.Use[contracts.Credentials](ctx)
	if err != nil {
		return nil, rocco.ErrForbidden.WithMessage("credential store unavailable")
	}
	identity := auth.IdentityFromContext(ctx)
	if identity == nil {
		return nil, rocco.ErrUnauthorized.WithMessage("identity required")
	}
	token, err := creds.Resolve(ctx, identity.ID(), CredentialKey)
	if err != nil {
		return nil, rocco.ErrForbidden.WithMessage("github credentials not configured")
	}
	return httpc.WithRequestBearerToken(token), nil
}

func endpoints(svc *service) []rocco.Endpoint {
	return []rocco.Endpoint{
		listRepos(svc),
		getRepo(svc),
		listIssues(svc),
		getIssue(svc),
		createIssue(svc),
		listPRs(svc),
		getPR(svc),
		createPR(svc),
		getFile(svc),
		createOrUpdateFile(svc),
		searchCode(svc),
	}
}

func listRepos(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, RepoList]("/github/repos", func(r *rocco.Request[rocco.NoBody]) (RepoList, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return RepoList{}, err
		}
		return svc.ListRepos(r, auth)
	}).WithName("listRepos").WithSummary("List repositories").WithTags("github").WithScopes(ScopeReposRead)
}

func getRepo(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, Repo]("/github/repos/{name}", func(r *rocco.Request[rocco.NoBody]) (Repo, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return Repo{}, err
		}
		return svc.GetRepo(r, r.Params.Path["name"], auth)
	}).WithName("getRepo").WithSummary("Get repository").WithTags("github").WithPathParams("name").WithScopes(ScopeReposRead)
}

func listIssues(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, IssueList]("/github/repos/{repo}/issues", func(r *rocco.Request[rocco.NoBody]) (IssueList, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return IssueList{}, err
		}
		return svc.ListIssues(r, r.Params.Path["repo"], auth)
	}).WithName("listIssues").WithSummary("List issues").WithTags("github").WithPathParams("repo").WithScopes(ScopeIssuesRead)
}

func getIssue(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, Issue]("/github/repos/{repo}/issues/{number}", func(r *rocco.Request[rocco.NoBody]) (Issue, error) {
		num, err := strconv.Atoi(r.Params.Path["number"])
		if err != nil {
			return Issue{}, rocco.ErrBadRequest.WithMessage("invalid issue number")
		}
		auth, err := resolveAuth(r)
		if err != nil {
			return Issue{}, err
		}
		return svc.GetIssue(r, r.Params.Path["repo"], num, auth)
	}).WithName("getIssue").WithSummary("Get issue").WithTags("github").WithPathParams("repo", "number").WithScopes(ScopeIssuesRead)
}

func createIssue(svc *service) rocco.Endpoint {
	return rocco.POST[CreateIssueRequest, Issue]("/github/repos/{repo}/issues", func(r *rocco.Request[CreateIssueRequest]) (Issue, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return Issue{}, err
		}
		return svc.CreateIssue(r, r.Params.Path["repo"], r.Body.Title, r.Body.Body, auth)
	}).WithName("createIssue").WithSummary("Create issue").WithTags("github").WithPathParams("repo").WithScopes(ScopeIssuesWrite)
}

func listPRs(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, PRList]("/github/repos/{repo}/pulls", func(r *rocco.Request[rocco.NoBody]) (PRList, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return PRList{}, err
		}
		return svc.ListPRs(r, r.Params.Path["repo"], auth)
	}).WithName("listPRs").WithSummary("List pull requests").WithTags("github").WithPathParams("repo").WithScopes(ScopePullsRead)
}

func getPR(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, PullRequest]("/github/repos/{repo}/pulls/{number}", func(r *rocco.Request[rocco.NoBody]) (PullRequest, error) {
		num, err := strconv.Atoi(r.Params.Path["number"])
		if err != nil {
			return PullRequest{}, rocco.ErrBadRequest.WithMessage("invalid PR number")
		}
		auth, err := resolveAuth(r)
		if err != nil {
			return PullRequest{}, err
		}
		return svc.GetPR(r, r.Params.Path["repo"], num, auth)
	}).WithName("getPR").WithSummary("Get pull request").WithTags("github").WithPathParams("repo", "number").WithScopes(ScopePullsRead)
}

func createPR(svc *service) rocco.Endpoint {
	return rocco.POST[CreatePRRequest, PullRequest]("/github/repos/{repo}/pulls", func(r *rocco.Request[CreatePRRequest]) (PullRequest, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return PullRequest{}, err
		}
		return svc.CreatePR(r, r.Params.Path["repo"], r.Body.Title, r.Body.Body, r.Body.Head, r.Body.Base, auth)
	}).WithName("createPR").WithSummary("Create pull request").WithTags("github").WithPathParams("repo").WithScopes(ScopePullsWrite)
}

func getFile(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, FileContent]("/github/repos/{repo}/content/{path...}", func(r *rocco.Request[rocco.NoBody]) (FileContent, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return FileContent{}, err
		}
		return svc.GetFile(r, r.Params.Path["repo"], r.Params.Path["path..."], r.Params.Query["ref"], auth)
	}).WithName("getFile").WithSummary("Get file content").WithTags("github").WithPathParams("repo", "path...").WithQueryParams("ref").WithScopes(ScopeContentRead)
}

func createOrUpdateFile(svc *service) rocco.Endpoint {
	return rocco.PUT[CreateOrUpdateFileRequest, FileContent]("/github/repos/{repo}/content/{path...}", func(r *rocco.Request[CreateOrUpdateFileRequest]) (FileContent, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return FileContent{}, err
		}
		return svc.CreateOrUpdateFile(r, r.Params.Path["repo"], r.Params.Path["path..."], r.Body.Message, r.Body.Content, r.Body.SHA, auth)
	}).WithName("createOrUpdateFile").WithSummary("Create or update file").WithTags("github").WithPathParams("repo", "path...").WithScopes(ScopeContentWrite)
}

func searchCode(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, CodeSearchResult]("/github/search/code", func(r *rocco.Request[rocco.NoBody]) (CodeSearchResult, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return CodeSearchResult{}, err
		}
		return svc.SearchCode(r, r.Params.Query["query"], auth)
	}).WithName("searchCode").WithSummary("Search code").WithTags("github").WithQueryParams("query").WithScopes(ScopeSearchRead)
}
