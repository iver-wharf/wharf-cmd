package wharfyml

import "strings"

type BuiltinVar string

const (
	BuiltinVarGitCommit              BuiltinVar = "GIT_COMMIT"
	BuiltinVarGitCommitSubject       BuiltinVar = "GIT_COMMIT_SUBJECT"
	BuiltinVarGitCommitAuthorDate    BuiltinVar = "GIT_COMMIT_AUTHOR_DATE"
	BuiltinVarGitCommitCommitterDate BuiltinVar = "GIT_COMMIT_COMMITTER_DATE"
	BuiltinVarGitTag                 BuiltinVar = "GIT_TAG"
	BuiltinVarGitBranch              BuiltinVar = "GIT_BRANCH"
	BuiltinVarGitSafeBranch          BuiltinVar = "GIT_SAFEBRANCH"
	BuiltinVarRegURL                 BuiltinVar = "REG_URL"
	BuiltinVarChartRepo              BuiltinVar = "CHART_REPO"
	BuiltinVarRepoName               BuiltinVar = "REPO_NAME"
	BuiltinVarRepoGroup              BuiltinVar = "REPO_GROUP"
	BuiltinVarRepoBranch             BuiltinVar = "REPO_BRANCH"
	BuiltinVarDefaultDomain          BuiltinVar = "DEFAULT_DOMAIN"
	BuiltinVarBuildRef               BuiltinVar = "BUILD_REF"
	BuiltinVarWharfProjectID         BuiltinVar = "WHARF_PROJECT_ID"
	BuiltinVarWharfInstance          BuiltinVar = "WHARF_INSTANCE"
)

func (t BuiltinVar) String() string {
	return string(t)
}

func ToSafeBranchName(name string) string {
	return strings.ReplaceAll(name, "/", "-")
}

func ConvertToParams(builtinVars map[BuiltinVar]string) map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range builtinVars {
		params[k.String()] = v
	}
	return params
}
