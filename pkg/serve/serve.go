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
	log "github.com/sirupsen/logrus"
	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/run"
	"k8s.io/client-go/rest"
)

type Server struct {
	Kubeconfig *rest.Config
	Namespace  string
}

func (s Server) Serve(listen string) {
	log.WithField("listen", listen).Infoln("Serving...")

	r := gin.Default()

	r.POST("/api/build", s.build)

	r.Run(listen)
}

func (s Server) build(c *gin.Context) {
	env := c.Query("ENVIRONMENT")

	gitFullUrl, err := url.QueryUnescape(c.DefaultQuery("GIT_FULLURL", ""))
	if err != nil {
		log.WithError(err).Errorln("Error query git full url")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	fixedUrl := gitToHttpsUrl(gitFullUrl)

	gitToken := c.Query("GIT_TOKEN")
	runStage := c.DefaultQuery("RUN_STAGES", "*")

	buildInVars, err := getBuiltinVarsFromQueryParams(c)
	if err != nil {
		log.WithError(err).Errorln("Error query params")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	buildInVars[wharfyml.BuiltinVarWharfInstance] = os.Getenv(wharfyml.BuiltinVarWharfInstance.String())

	buildID, err := strconv.Atoi(buildInVars[wharfyml.BuiltinVarBuildRef])
	if err != nil {
		log.WithError(err).Errorln(fmt.Sprintf("invalid %s number: %s", wharfyml.BuiltinVarBuildRef, buildInVars[wharfyml.BuiltinVarBuildRef]))
		c.JSON(http.StatusBadRequest, err)
		return
	}

	log.WithFields(log.Fields{
		"env":        env,
		"branch":     buildInVars[wharfyml.BuiltinVarGitBranch],
		"repo":       buildInVars[wharfyml.BuiltinVarRepoName],
		"group":      buildInVars[wharfyml.BuiltinVarRepoGroup],
		"registry":   buildInVars[wharfyml.BuiltinVarRegURL],
		"gitFullUrl": gitFullUrl,
		"fixedUrl":   fixedUrl}).
		Infoln("Starting build!")

	tempDir, err := ioutil.TempDir(os.TempDir(), "wharf")
	if err != nil {
		log.WithError(err).Errorln("Error creating temp-dir.")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	defer os.RemoveAll(tempDir)

	log.WithField("dir", tempDir).Infoln("Created temp-dir")

	repo, err := gitClone(tempDir, fixedUrl, buildInVars[wharfyml.BuiltinVarGitBranch], gitToken)
	if err != nil {
		log.WithError(err).Errorln("Error cloning repo.")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	repoParams, err := getBuiltinVarsFromCommit(repo)
	if err != nil {
		log.WithError(err).Errorln("Error getting built-in params from repo.")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	for k, v := range repoParams {
		buildInVars[k] = v
	}

	def, err := wharfyml.Parse(filepath.Join(tempDir, wharfyml.WharfCIFileName), buildInVars)
	if err != nil {
		log.WithError(err).Errorln("Error parsing build-definition.")
		c.JSON(http.StatusBadRequest, err)
		return
	}

	log.Infoln("Parsed build definition")

	runner := run.NewRunner(s.Kubeconfig, c.GetHeader("Authorization"))
	err = runner.RunDefinition(def, env, s.Namespace, runStage, buildID, buildInVars)
	if err != nil {
		log.WithError(err).Errorln("Error running build definition.")
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

func getBuiltinVarsFromQueryParams(c *gin.Context) (map[wharfyml.BuiltinVar]string, error) {
	var err error
	buildInVars := map[wharfyml.BuiltinVar]string{
		wharfyml.BuiltinVarBuildRef:       c.Query(wharfyml.BuiltinVarBuildRef.String()),
		wharfyml.BuiltinVarWharfProjectID: c.Query(wharfyml.BuiltinVarWharfProjectID.String()),
		wharfyml.BuiltinVarGitBranch:      c.Query(wharfyml.BuiltinVarGitBranch.String()),
		wharfyml.BuiltinVarRepoBranch:     c.Query(wharfyml.BuiltinVarRepoBranch.String()),
		wharfyml.BuiltinVarRepoName:       c.Query(wharfyml.BuiltinVarRepoName.String()),
		wharfyml.BuiltinVarRepoGroup:      c.Query(wharfyml.BuiltinVarRepoGroup.String()),
		wharfyml.BuiltinVarChartRepo:      c.Query(wharfyml.BuiltinVarChartRepo.String()),
		wharfyml.BuiltinVarDefaultDomain:  c.Query(wharfyml.BuiltinVarDefaultDomain.String()),
	}

	buildInVars[wharfyml.BuiltinVarGitSafeBranch] = wharfyml.ToSafeBranchName(buildInVars[wharfyml.BuiltinVarGitBranch])

	buildInVars[wharfyml.BuiltinVarRegURL], err = url.QueryUnescape(c.Query(wharfyml.BuiltinVarRegURL.String()))
	if err != nil {
		log.WithError(err).Errorln(fmt.Sprintf("Error query %s", wharfyml.BuiltinVarRegURL))
		c.JSON(http.StatusBadRequest, err)
		return buildInVars, fmt.Errorf("error query %s: %w", wharfyml.BuiltinVarRegURL, err)
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

	log.Infoln("Cloned repository")

	wharfcifilePath := filepath.Join(tempDir, wharfyml.WharfCIFileName)
	_, err = os.Stat(wharfcifilePath)
	if err != nil {
		return nil, fmt.Errorf("repository is missing %s file: %w", wharfyml.WharfCIFileName, err)
	}

	return repo, nil
}

func getBuiltinVarsFromCommit(repo *git.Repository) (map[wharfyml.BuiltinVar]string, error) {
	commitIter, err := repo.Log(&git.LogOptions{
		From:     plumbing.Hash{},
		Order:    0,
		FileName: nil,
		All:      false,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting log: %w", err)
	}

	defer commitIter.Close()
	commit, err := commitIter.Next()
	if err != nil {
		return nil, fmt.Errorf("error getting latest commit: %w", err)
	}

	buildInVars := map[wharfyml.BuiltinVar]string{
		wharfyml.BuiltinVarGitCommit:              commit.Hash.String(),
		wharfyml.BuiltinVarGitCommitSubject:       commit.Message,
		wharfyml.BuiltinVarGitCommitAuthorDate:    commit.Author.When.Format(time.RFC3339),
		wharfyml.BuiltinVarGitCommitCommitterDate: commit.Committer.When.Format(time.RFC3339),
	}

	tagName, err := getLatestTagNameFromCommit(repo, commit)
	if err != nil {
		log.WithError(err).Warn("Error getting built-in params from repo.")
	}

	buildInVars[wharfyml.BuiltinVarGitTag] = tagName

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
