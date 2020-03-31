package avancement

import (
	"testing"

	"github.com/jenkins-x/go-scm/scm"
	fakescm "github.com/jenkins-x/go-scm/scm/driver/fake"
	"github.com/rhd-gitops-example/services/pkg/git"
	"github.com/rhd-gitops-example/services/pkg/git/mock"
)

const (
	dev     = "https://example.com/testing/dev-env"
	staging = "https://example.com/testing/staging-env"
)

func TestPromoteWithSuccessKeepCacheTrue(t *testing.T) {
	promoteWithSuccess(t, true)
}

func TestPromoteWithSuccessKeepCacheFalse(t *testing.T) {
	promoteWithSuccess(t, false)
}

func promoteWithSuccess(t *testing.T, keepCache bool) {
	dstBranch := "test-branch"
	author := &git.Author{Name: "Testing User", Email: "testing@example.com", Token: "test-token"}
	client, _ := fakescm.NewDefault()
	fakeClientFactory := func(s string) *scm.Client {
		return client
	}
	devRepo, stagingRepo := mock.New("/dev", "master"), mock.New("/staging", "master")
	repos := map[string]*mock.Repository{
		mustAddCredentials(t, dev, author):     devRepo,
		mustAddCredentials(t, staging, author): stagingRepo,
	}
	sm := New("tmp", author)
	sm.clientFactory = fakeClientFactory
	sm.repoFactory = func(url, _ string, _ bool) (git.Repo, error) {
		return git.Repo(repos[url]), nil
	}
	devRepo.AddFiles("/services/my-service/base/config/myfile.yaml")

	err := sm.Promote("my-service", dev, staging, dstBranch, keepCache)
	if err != nil {
		t.Fatal(err)
	}

	stagingRepo.AssertBranchCreated(t, "master", dstBranch)
	stagingRepo.AssertFileCopiedInBranch(t, dstBranch, "/dev/services/my-service/base/config/myfile.yaml", "/staging/services/my-service/base/config/myfile.yaml")
	stagingRepo.AssertCommit(t, dstBranch, defaultCommitMsg, author)
	stagingRepo.AssertPush(t, dstBranch)

	if keepCache {
		stagingRepo.AssertNotDeletedFromCache(t)
		devRepo.AssertNotDeletedFromCache(t)
	} else {
		stagingRepo.AssertDeletedFromCache(t)
		devRepo.AssertDeletedFromCache(t)
	}
}

func TestAddCredentials(t *testing.T) {
	testUser := &git.Author{Name: "Test User", Email: "test@example.com", Token: "test-token"}
	tests := []struct {
		repoURL string
		a       *git.Author
		want    string
	}{
		{"https://testing.example.com/test", testUser, "https://promotion:test-token@testing.example.com/test"},
		{"https://promotion:my-token@testing.example.com/test", testUser, "https://promotion:my-token@testing.example.com/test"},
		{"https://testing:atoken@testing.example.com/test", testUser, "https://testing:atoken@testing.example.com/test"},
		{"/mydir/test", testUser, "/mydir/test"},
	}

	for i, tt := range tests {
		got, err := addCredentialsIfNecessary(tt.repoURL, tt.a)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("addCredentials() %d got %s, want %s", i, got, tt.want)
		}
	}
}

func mustAddCredentials(t *testing.T, repoURL string, a *git.Author) string {
	u, err := addCredentialsIfNecessary(repoURL, a)
	if err != nil {
		t.Fatalf("failed to add credentials to %s: %e", repoURL, err)
	}
	return u
}

func TestPromoteWithCacheDeletionFailure(t *testing.T) {
	dstBranch := mock.FAIL_DELETING_REPO_BRANCH
	author := &git.Author{Name: "Testing User", Email: "testing@example.com", Token: "test-token"}
	client, _ := fakescm.NewDefault()
	fakeClientFactory := func(s string) *scm.Client {
		return client
	}
	devRepo, stagingRepo := mock.New("/dev", "master"), mock.New("/staging", "master")
	repos := map[string]*mock.Repository{
		mustAddCredentials(t, dev, author):     devRepo,
		mustAddCredentials(t, staging, author): stagingRepo,
	}
	sm := New("tmp", author)
	sm.clientFactory = fakeClientFactory
	sm.repoFactory = func(url, _ string, _ bool) (git.Repo, error) {
		return git.Repo(repos[url]), nil
	}
	devRepo.AddFiles("/services/my-service/base/config/myfile.yaml")

	err := sm.Promote("my-service", dev, staging, dstBranch, false)
	if err != nil {
		t.Fatal(err)
	}

	stagingRepo.AssertBranchCreated(t, "master", dstBranch)
	stagingRepo.AssertFileCopiedInBranch(t, dstBranch, "/dev/services/my-service/base/config/myfile.yaml", "/staging/services/my-service/base/config/myfile.yaml")
	stagingRepo.AssertCommit(t, dstBranch, defaultCommitMsg, author)
	stagingRepo.AssertPush(t, dstBranch)

	stagingRepo.AssertNotDeletedFromCache(t)
	devRepo.AssertDeletedFromCache(t)
}
