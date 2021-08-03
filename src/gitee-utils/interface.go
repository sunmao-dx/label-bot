package gitee_utils

import (
	sdk "gitee.com/openeuler/go-gitee/gitee"
	"net/http"
)

// Client interface for Gitee API
type Client interface {

	CreatePullRequest(org, repo, title, body, head, base string, canModify bool) (sdk.PullRequest, error)
	GetPullRequests(org, repo string, opts ListPullRequestOpt) ([]sdk.PullRequest, error)
	UpdatePullRequest(org, repo string, number int32, param sdk.PullRequestUpdateParam) (sdk.PullRequest, error)

	GetRef(org, repo, ref string) (string, error)
	GetPRLabels(org, repo string, number int) ([]sdk.Label, error)
	ListPRComments(org, repo string, number int) ([]sdk.PullRequestComments, error)
	ListPrIssues(org, repo string, number int32) ([] sdk.Issue, error)
	DeletePRComment(org, repo string, ID int) error
	CreatePRComment(org, repo string, number int, comment string) error
	UpdatePRComment(org, repo string, commentID int, comment string) error
	AddPRLabel(org, repo string, number int, labels []string) error
	RemovePRLabel(org, repo string, number int, label string) error

	AssignPR(owner, repo string, number int, logins []string) error
	UnassignPR(owner, repo string, number int, logins []string) error
	GetPRCommits(org, repo string, number int) ([]sdk.PullRequestCommits, error)

	AssignGiteeIssue(org, repo, labels string, number string, login string) error
	UnassignGiteeIssue(org, repo, labels string, number string, login string) error
	CreateGiteeIssueComment(org, repo string, number string, comment string) error

	IsCollaborator(owner, repo, login string) (bool, error)
	IsMember(org, login string) (bool, error)
	GetGiteePullRequest(org, repo string, number int) (sdk.PullRequest, error)
	GetGiteeRepo(org, repo string) (sdk.Project, error)
	MergePR(owner, repo string, number int, opt sdk.PullRequestMergePutParam) error

	GetRepos(org string) ([]sdk.Project, error)
	RemoveIssueLabel(org, repo, number, label string) error
	AddIssueLabel(org, repo, number string, label []string) error
	AddIssueAssignee(org, repo, number, token, assignee string) error
	GetUserOrg(login string) ([]sdk.Group ,error)
	GetUserEnt(ent, login string) (sdk.EnterpriseMember ,error)

	ListIssues(owner, repo, state, since, createAt string, page, perPage int) ([]sdk.Issue, *http.Response, error)
	ListLabels(owner, repo string) ([]sdk.Label ,error)
	GetRecommendation(labels string) (string, error)
}

type ListPullRequestOpt struct {
	State           string
	Head            string
	Base            string
	Sort            string
	Direction       string
	MilestoneNumber int
	Labels          []string
}
