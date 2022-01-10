package serve

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	auth "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	gitenv "github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/git"
	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/run"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"k8s.io/client-go/rest"
)

var log = logger.New()

type Server struct {
	Kubeconfig *rest.Config
	Namespace  string
}

func (s Server) Serve(bindAddress string) {
	log.Info().WithString("address", bindAddress).Message("Listening...")

	r := gin.Default()

	r.POST("/api/build", s.build)

	r.Run(bindAddress)
}

func (s Server) build(c *gin.Context) {
	env := c.Query("ENVIRONMENT")

	gitFullUrl, err := url.QueryUnescape(c.DefaultQuery("GIT_FULLURL", ""))
	if err != nil {
		log.Error().WithError(err).Message("Missing GIT_FULLURL query parameter.")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	fixedUrl := gitToHttpsUrl(gitFullUrl)

	gitToken := c.Query("GIT_TOKEN")
	runStage := c.DefaultQuery("RUN_STAGES", "*")

	buildInVars, err := getBuiltinVarsFromQueryParams(c)
	if err != nil {
		log.Error().WithError(err).Message("Missing GIT_FULLURL query parameter.")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	buildInVars[containercreator.BuiltinVarWharfInstance] = os.Getenv(containercreator.BuiltinVarWharfInstance.String())

	buildID, err := strconv.Atoi(buildInVars[containercreator.BuiltinVarBuildRef])
	if err != nil {
		log.Error().WithError(err).
			WithString("number", buildInVars[containercreator.BuiltinVarBuildRef]).
			Messagef("Failed parsing %s.", containercreator.BuiltinVarBuildRef)
		c.JSON(http.StatusBadRequest, err)
		return
	}

	log.Info().
		WithString("env", env).
		WithString("branch", buildInVars[containercreator.BuiltinVarGitBranch]).
		WithString("repo", buildInVars[containercreator.BuiltinVarRepoName]).
		WithString("group", buildInVars[containercreator.BuiltinVarRepoGroup]).
		WithString("registry", buildInVars[containercreator.BuiltinVarRegURL]).
		WithString("gitFullUrl", gitFullUrl).
		WithString("fixedUrl", fixedUrl).
		Message("Started build.")

	tempDir, err := ioutil.TempDir(os.TempDir(), "wharf")
	if err != nil {
		log.Error().WithError(err).Message("Error creating temp-dir.")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	defer os.RemoveAll(tempDir)

	log.Info().WithString("dir", tempDir).Message("Created temp-dir.")

	repo, err := gitClone(tempDir, fixedUrl, buildInVars[containercreator.BuiltinVarGitBranch], gitToken)
	if err != nil {
		log.Error().WithError(err).Message("Error cloning repo.")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	repoParams, err := getBuiltinVarsFromCommit(repo)
	if err != nil {
		log.Error().WithError(err).Message("Error getting built-in params from repo.")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	for k, v := range repoParams {
		buildInVars[k] = v
	}

	def, err := wharfyml.Parse(filepath.Join(tempDir, wharfyml.WharfCIFileName), buildInVars)
	if err != nil {
		log.Error().WithError(err).Message("Error parsing build-definition.")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	log.Info().Message("Parsed build definition.")

	gitParams := gitenv.NewGitPropertiesMap(fixedUrl, buildInVars[containercreator.BuiltinVarGitBranch], gitToken)
	runner := run.NewRunner(s.Kubeconfig, c.GetHeader("Authorization"))
	err = runner.RunDefinition(def, env, s.Namespace, runStage, buildID, gitParams, buildInVars)
	if err != nil {
		log.Error().WithError(err).Message("Error running build definition.")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, "Success!")
}

func gitToHttpsUrl(gitUrl string) string {
	u := strings.ReplaceAll(gitUrl, ":", "/")
	u = strings.ReplaceAll(u, "git://", "https://")
	u = strings.ReplaceAll(u, "git@", "")
	return fmt.Sprintf("https://%s", u)
}

func getBuiltinVarsFromQueryParams(c *gin.Context) (map[containercreator.BuiltinVar]string, error) {
	var err error
	buildInVars := map[containercreator.BuiltinVar]string{
		containercreator.BuiltinVarBuildRef:       c.Query(containercreator.BuiltinVarBuildRef.String()),
		containercreator.BuiltinVarWharfProjectID: c.Query(containercreator.BuiltinVarWharfProjectID.String()),
		containercreator.BuiltinVarGitBranch:      c.Query(containercreator.BuiltinVarGitBranch.String()),
		containercreator.BuiltinVarRepoBranch:     c.Query(containercreator.BuiltinVarRepoBranch.String()),
		containercreator.BuiltinVarRepoName:       c.Query(containercreator.BuiltinVarRepoName.String()),
		containercreator.BuiltinVarRepoGroup:      c.Query(containercreator.BuiltinVarRepoGroup.String()),
		containercreator.BuiltinVarChartRepo:      c.Query(containercreator.BuiltinVarChartRepo.String()),
		containercreator.BuiltinVarDefaultDomain:  c.Query(containercreator.BuiltinVarDefaultDomain.String()),
	}
	buildInVars[containercreator.BuiltinVarGitSafeBranch] = containercreator.ToSafeBranchName(buildInVars[containercreator.BuiltinVarGitBranch])

	buildInVars[containercreator.BuiltinVarRegURL], err = url.QueryUnescape(c.Query(containercreator.BuiltinVarRegURL.String()))
	if err != nil {
		log.Error().WithError(err).
			WithStringer("param", containercreator.BuiltinVarRegURL).
			Message("Failed unescaping query parameter.")
		c.JSON(http.StatusBadRequest, err)
		return buildInVars, fmt.Errorf("error query %s: %w", containercreator.BuiltinVarRegURL, err)
	}

	return buildInVars, nil
}

func gitClone(tempDir string, gitUrl string, branch string, token string) (*git.Repository, error) {
	repo, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL: gitUrl,
		Auth: &auth.BasicAuth{
			Password: token,
		},
		SingleBranch:  true,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		return nil, fmt.Errorf("error cloning repo: %w", err)
	}

	log.Info().
		WithString("git", gitUrl).
		WithString("branch", branch).
		Message("Cloned repository.")

	wharfcifilePath := filepath.Join(tempDir, wharfyml.WharfCIFileName)
	_, err = os.Stat(wharfcifilePath)
	if err != nil {
	}

	return repo, nil
}

func getBuiltinVarsFromCommit(repo *git.Repository) (map[containercreator.BuiltinVar]string, error) {
	commitIter, err := repo.Log(&git.LogOptions{
		From:     plumbing.Hash{},
		Order:    0,
		FileName: nil,
		All:      false,
	})
	if err != nil {
	}

	defer commitIter.Close()
	commit, err := commitIter.Next()
	if err != nil {
		return nil, fmt.Errorf("error getting log: %w", err)
	}

	buildInVars := map[containercreator.BuiltinVar]string{
		containercreator.BuiltinVarGitCommit:              commit.Hash.String(),
		containercreator.BuiltinVarGitCommitSubject:       commit.Message,
		containercreator.BuiltinVarGitCommitAuthorDate:    commit.Author.When.Format(time.RFC3339),
		containercreator.BuiltinVarGitCommitCommitterDate: commit.Committer.When.Format(time.RFC3339),
	}

	tagName, err := getLatestTagNameFromCommit(repo, commit)
	if err != nil {
		log.Warn().WithError(err).Message("Failed getting some built-in params from Git commit.")
	}

	buildInVars[containercreator.BuiltinVarGitTag] = tagName

	return buildInVars, nil
}

func getLatestTagNameFromCommit(repo *git.Repository, commit *object.Commit) (string, error) {
	tagsIter, err := repo.Tags()
	if err != nil {
		return "", fmt.Errorf("error getting tags iterator: %w", err)
	}

	defer tagsIter.Close()

	var latest plumbing.Reference
	var when time.Time
	tagExists := false

	if err := tagsIter.ForEach(
		func(ref *plumbing.Reference) error {
			if ref.Hash().String() == commit.Hash.String() {
				tagExists = true

				time, err := getTaggedTime(repo, ref.Hash(), commit.Committer.When)
				if err != nil {
					return err
				}

				if time.After(when) {
					when = time
					latest = *ref
				}
			}
			return nil
		}); err != nil {
		return "", fmt.Errorf("error getting tag: %w", err)
	}

	if !tagExists {
		return "", nil
	}

	return latest.Name().String(), nil
}

func getTaggedTime(repo *git.Repository, tagHash plumbing.Hash, commiterTime time.Time) (time.Time, error) {
	obj, err := repo.TagObject(tagHash)
	if err == plumbing.ErrObjectNotFound {
		return commiterTime, nil
	} else if err == nil {
		return obj.Tagger.When, nil
	} else {
		return time.Time{}, err
	}
}
