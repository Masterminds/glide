package path

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func generateTestDirectory(t *testing.T) string {
	t.Helper()
	baseDir, err := ioutil.TempDir(os.TempDir(), "mgt")
	if nil != err {
		t.Error("Unable to create temp directory: ", err.Error())
	}
	paths := map[string][]string{
		"github.com/fake/log":                                                                            {"log.go"},
		"github.com/phoney/foo":                                                                          {"bar.go"},
		"github.com/phoney/foo/vendor":                                                                   {"test.go", "foo.bar"},
		"github.com/aws/aws-sdk-go/awsmigrate/awsmigrate-renamer/vendor":                                 {},
		"github.com/aws/aws-sdk-go/awsmigrate/awsmigrate-renamer/vendor/golang.org/x/tools/go/buildutil": {"allpackages.go", "tags.go", "fakecontext.go"},
		"github.com/aws/aws-sdk-go/vendor":                                                               {"key_test.go", "key.go"},
		"github.com/aws/aws-sdk-go/vendor/github.com/go-ini/ini":                                         {"struct_test.go", "error.go", "ini_test.go"},
	}
	os.OpenFile(path.Join(baseDir, "glide.yaml"), os.O_RDONLY|os.O_CREATE, 0666)
	for p, files := range paths {
		p = path.Join(baseDir, "vendor", p)
		if err = os.MkdirAll(p, 0777); nil != err {
			t.Errorf("Unable to create vendor dir: %s\n%s", p, err.Error())
		}
		for _, f := range files {
			os.OpenFile(path.Join(p, f), os.O_RDONLY|os.O_CREATE, 0666)
		}
	}
	return baseDir
}

func TestNestVendorNoError(t *testing.T) {
	workingDir := generateTestDirectory(t)
	os.Chdir(workingDir)
	err := StripVendor()
	if nil != err {
		t.Errorf("Unexpected error in StripVendor: %s", err.Error())
	}
}
