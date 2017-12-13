package path

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type mockFileInfo struct {
	name  string
	isDir bool
}

func (mfi *mockFileInfo) Name() string {
	return mfi.name
}

func (mfi *mockFileInfo) Size() int64 {
	panic("not implemented")
}

func (mfi *mockFileInfo) Mode() os.FileMode {
	panic("not implemented")
}

func (mfi *mockFileInfo) ModTime() time.Time {
	panic("not implemented")
}

func (mfi *mockFileInfo) IsDir() bool {
	return mfi.isDir
}

func (mfi *mockFileInfo) Sys() interface{} {
	panic("not implemented")
}

type removeAll struct {
	calledWith string
	err        error
}

func (rah *removeAll) removeAll(p string) error {
	rah.calledWith = p
	return rah.err
}

func TestWalkFunction(t *testing.T) {
	type args struct {
		searchPath string
		removeAll  *removeAll
		path       string
		info       os.FileInfo
		err        error
	}
	tests := []struct {
		name           string
		args           args
		want           error
		wantCalledWith string
	}{
		{
			name: "WalkFunctionSkipsNonVendor",
			args: args{searchPath: "foo",
				removeAll: &removeAll{},
				path:      "foo/bar",
				info:      &mockFileInfo{name: "bar", isDir: true},
				err:       nil,
			},
			want:           nil,
			wantCalledWith: "",
		},
		{
			name: "WalkFunctionSkipsNonDir",
			args: args{searchPath: "foo",
				removeAll: &removeAll{},
				path:      "foo/vendor",
				info:      &mockFileInfo{name: "vendor", isDir: false},
				err:       nil,
			},
			want:           nil,
			wantCalledWith: "",
		},
		{
			name: "WalkFunctionDeletesVendor",
			args: args{searchPath: "foo",
				removeAll: &removeAll{},
				path:      "foo/vendor",
				info:      &mockFileInfo{name: "vendor", isDir: true},
				err:       nil,
			},
			want:           filepath.SkipDir,
			wantCalledWith: "foo/vendor",
		},
		{
			name: "WalkFunctionReturnsPassedError",
			args: args{searchPath: "foo",
				removeAll: &removeAll{},
				path:      "foo/vendor",
				info:      &mockFileInfo{name: "vendor", isDir: true},
				err:       errors.New("expected"),
			},
			want:           errors.New("expected"),
			wantCalledWith: "",
		},
		{
			name: "WalkFunctionReturnsRemoveAllError",
			args: args{searchPath: "foo",
				removeAll: &removeAll{err: errors.New("expected")},
				path:      "foo/vendor",
				info:      &mockFileInfo{name: "vendor", isDir: true},
				err:       nil,
			},
			want:           errors.New("expected"),
			wantCalledWith: "foo/vendor",
		},
		{
			name: "WalkFunctionSkipsBaseDir",
			args: args{searchPath: "vendor",
				removeAll: &removeAll{},
				path:      "vendor",
				info:      &mockFileInfo{name: "vendor", isDir: true},
				err:       nil,
			},
			want:           nil,
			wantCalledWith: "",
		},
	}
	for _, test := range tests {
		walkFunction := getWalkFunction(test.args.searchPath, test.args.removeAll.removeAll)
		if actual := walkFunction(test.args.path, test.args.info, test.args.err); !reflect.DeepEqual(actual, test.want) {
			t.Errorf("walkFunction() = %v, want %v", actual, test.want)
		}
		if test.args.removeAll.calledWith != test.wantCalledWith {
			t.Errorf("removeAll argument = \"%s\", want \"%s\"", test.args.removeAll.calledWith, test.wantCalledWith)
		}
	}
}
