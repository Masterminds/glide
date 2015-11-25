package vcs

import (
	"io/ioutil"
	//"log"
	"os"
	"testing"
)

// Canary test to ensure HgRepo implements the Repo interface.
var _ Repo = &HgRepo{}

// To verify hg is working we perform intergration testing
// with a known hg service.

func TestHg(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "go-vcs-hg-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	repo, err := NewHgRepo("https://bitbucket.org/mattfarina/testhgrepo", tempDir+"/testhgrepo")
	if err != nil {
		t.Error(err)
	}

	if repo.Vcs() != Hg {
		t.Error("Hg is detecting the wrong type")
	}

	// Check the basic getters.
	if repo.Remote() != "https://bitbucket.org/mattfarina/testhgrepo" {
		t.Error("Remote not set properly")
	}
	if repo.LocalPath() != tempDir+"/testhgrepo" {
		t.Error("Local disk location not set properly")
	}

	//Logger = log.New(os.Stdout, "", log.LstdFlags)

	// Do an initial clone.
	err = repo.Get()
	if err != nil {
		t.Errorf("Unable to clone Hg repo. Err was %s", err)
	}

	// Verify Hg repo is a Hg repo
	if repo.CheckLocal() == false {
		t.Error("Problem checking out repo or Hg CheckLocal is not working")
	}

	// Test internal lookup mechanism used outside of Hg specific functionality.
	ltype, err := DetectVcsFromFS(tempDir + "/testhgrepo")
	if err != nil {
		t.Error("detectVcsFromFS unable to Hg repo")
	}
	if ltype != Hg {
		t.Errorf("detectVcsFromFS detected %s instead of Hg type", ltype)
	}

	// Test NewRepo on existing checkout. This should simply provide a working
	// instance without error based on looking at the local directory.
	nrepo, nrerr := NewRepo("https://bitbucket.org/mattfarina/testhgrepo", tempDir+"/testhgrepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
	// Verify the right oject is returned. It will check the local repo type.
	if nrepo.CheckLocal() == false {
		t.Error("Wrong version returned from NewRepo")
	}

	// Set the version using the short hash.
	err = repo.UpdateVersion("a5494ba2177f")
	if err != nil {
		t.Errorf("Unable to update Hg repo version. Err was %s", err)
	}

	// Use Version to verify we are on the right version.
	v, err := repo.Version()
	if v != "a5494ba2177f" {
		t.Error("Error checking checked out Hg version")
	}
	if err != nil {
		t.Error(err)
	}

	// Use Date to verify we are on the right commit.
	d, err := repo.Date()
	if err != nil {
		t.Error(err)
	}
	if d.Format(longForm) != "2015-07-30 16:14:08 -0400" {
		t.Error("Error checking checked out Hg commit date. Got wrong date:", d)
	}

	// Perform an update.
	err = repo.Update()
	if err != nil {
		t.Error(err)
	}

	v, err = repo.Version()
	if v != "9c6ccbca73e8" {
		t.Error("Error checking checked out Hg version")
	}
	if err != nil {
		t.Error(err)
	}

	tags, err := repo.Tags()
	if err != nil {
		t.Error(err)
	}
	if tags[1] != "1.0.0" {
		t.Error("Hg tags is not reporting the correct version")
	}

	branches, err := repo.Branches()
	if err != nil {
		t.Error(err)
	}
	// The branches should be HEAD, master, and test.
	if branches[0] != "test" {
		t.Error("Hg is incorrectly returning branches")
	}

	if repo.IsReference("1.0.0") != true {
		t.Error("Hg is reporting a reference is not one")
	}

	if repo.IsReference("test") != true {
		t.Error("Hg is reporting a reference is not one")
	}

	if repo.IsReference("foo") == true {
		t.Error("Hg is reporting a non-existant reference is one")
	}

	if repo.IsDirty() == true {
		t.Error("Hg incorrectly reporting dirty")
	}

}

func TestHgCheckLocal(t *testing.T) {
	// Verify repo.CheckLocal fails for non-Hg directories.
	// TestHg is already checking on a valid repo
	tempDir, err := ioutil.TempDir("", "go-vcs-hg-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	repo, _ := NewHgRepo("", tempDir)
	if repo.CheckLocal() == true {
		t.Error("Hg CheckLocal does not identify non-Hg location")
	}

	// Test NewRepo when there's no local. This should simply provide a working
	// instance without error based on looking at the remote localtion.
	_, nrerr := NewRepo("https://bitbucket.org/mattfarina/testhgrepo", tempDir+"/testhgrepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
}
