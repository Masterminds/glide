// Package govendor provides compatibility with govendor vendorfiles.

// This is 1-to-1 copy of govendor's `vendorfile/file.go` file.
// govendor is governed by a BSD-style
// license that can be found in the LICENSE file of govendor project

package govendor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
)

// Name of the vendor file.
const Name = "vendor.json"

// File is the structure of the vendor file.
type File struct {
	RootPath string // Import path of vendor folder

	Comment string

	Ignore string

	Package []*Package

	// all preserves unknown values.
	all map[string]interface{}
}

// Package represents each package.
type Package struct {
	field map[string]interface{}

	// If delete is set to true the package will not be written to the vendor file.
	Remove bool

	// If new is set to true the package will be treated as a new package to the file.
	Add bool

	// See the vendor spec for definitions.
	Origin       string
	Path         string
	Tree         bool
	Revision     string
	RevisionTime string
	Version      string
	VersionExact string
	ChecksumSHA1 string
	Comment      string
}

func (pkg *Package) PathOrigin() string {
	if len(pkg.Origin) > 0 {
		return pkg.Origin
	}
	return pkg.Path
}

// The following stringer functions are useful for debugging.

type packageList []*Package

func (list packageList) String() string {
	buf := &bytes.Buffer{}
	for _, item := range list {
		buf.WriteString("\t")
		buf.WriteString(fmt.Sprintf("(%v) ", item.field))
		if item.Remove {
			buf.WriteString(" X ")
		}
		buf.WriteString(item.Path)
		buf.WriteRune('\n')
	}
	buf.WriteRune('\n')
	return buf.String()
}

func allString(all map[string]interface{}) string {
	obj, _ := all["package"]
	buf := &bytes.Buffer{}
	for _, itemObj := range obj.([]interface{}) {
		item := itemObj.(map[string]interface{})
		buf.WriteString("\t")
		buf.WriteString(item["path"].(string))
		buf.WriteRune('\n')
	}
	buf.WriteRune('\n')
	return buf.String()
}

var (
	rootPathNames     = []string{"rootPath"}
	packageNames      = []string{"package", "Package"}
	ignoreNames       = []string{"ignore"}
	originNames       = []string{"origin"}
	pathNames         = []string{"path", "canonical", "Canonical", "vendor", "Vendor"}
	treeNames         = []string{"tree"}
	revisionNames     = []string{"revision", "Revision", "version", "Version"}
	revisionTimeNames = []string{"revisionTime", "RevisionTime", "versionTime", "VersionTime"}
	versionNames      = []string{"version"}
	versionExactNames = []string{"versionExact"}
	checksumSHA1Names = []string{"checksumSHA1"}
	commentNames      = []string{"comment", "Comment"}
)

type vendorPackageSort []interface{}

func (vp vendorPackageSort) Len() int      { return len(vp) }
func (vp vendorPackageSort) Swap(i, j int) { vp[i], vp[j] = vp[j], vp[i] }
func (vp vendorPackageSort) Less(i, j int) bool {
	a := vp[i].(map[string]interface{})
	b := vp[j].(map[string]interface{})
	aPath, _ := a[pathNames[0]].(string)
	bPath, _ := b[pathNames[0]].(string)

	if aPath == bPath {
		aOrigin, _ := a[originNames[0]].(string)
		bOrigin, _ := b[originNames[0]].(string)
		return len(aOrigin) > len(bOrigin)
	}
	return aPath < bPath
}

func setField(fieldObj interface{}, object map[string]interface{}, names []string) {
loop:
	for _, name := range names {
		raw, found := object[name]
		if !found {
			continue
		}
		switch field := fieldObj.(type) {
		default:
			panic("unknown type")
		case *string:
			value, is := raw.(string)
			if !is {
				continue loop
			}
			*field = value
			if len(value) != 0 {
				break loop
			}
		case *bool:
			value, is := raw.(bool)
			if !is {
				continue loop
			}
			*field = value
			if value == true {
				break loop
			}
		}
	}
}

func setObject(fieldObj interface{}, object map[string]interface{}, names []string, hideEmpty bool) {
	switch field := fieldObj.(type) {
	default:
		panic("unknown type")
	case string:
		for i, name := range names {
			if i != 0 || (hideEmpty && len(field) == 0) {
				delete(object, name)
				continue
			}
			object[name] = field
		}
	case bool:
		for i, name := range names {
			if i != 0 || (hideEmpty && field == false) {
				delete(object, name)
				continue
			}
			object[name] = field
		}
	}
}

// getRawPackageList gets the array of items from all object.
func (vf *File) getRawPackageList() []interface{} {
	var rawPackageList []interface{}
	for index, name := range packageNames {
		rawPackageListObject, found := vf.all[name]
		if !found {
			continue
		}
		if index != 0 {
			vf.all[packageNames[0]] = rawPackageListObject
			delete(vf.all, name)
		}
		var is bool
		rawPackageList, is = rawPackageListObject.([]interface{})
		if is {
			break
		}
	}
	return rawPackageList
}

// toFields moves values from "all" to the field values.
func (vf *File) toFields() {
	setField(&vf.RootPath, vf.all, rootPathNames)
	setField(&vf.Comment, vf.all, commentNames)
	setField(&vf.Ignore, vf.all, ignoreNames)

	rawPackageList := vf.getRawPackageList()

	vf.Package = make([]*Package, len(rawPackageList))

	for index, rawPackage := range rawPackageList {
		object, is := rawPackage.(map[string]interface{})
		if !is {
			continue
		}
		pkg := &Package{}
		vf.Package[index] = pkg
		pkg.field = object
		setField(&pkg.Origin, object, originNames)
		setField(&pkg.Path, object, pathNames)
		setField(&pkg.Tree, object, treeNames)
		setField(&pkg.Revision, object, revisionNames)
		setField(&pkg.RevisionTime, object, revisionTimeNames)
		setField(&pkg.Version, object, versionNames)
		setField(&pkg.VersionExact, object, versionExactNames)
		setField(&pkg.ChecksumSHA1, object, checksumSHA1Names)
		setField(&pkg.Comment, object, commentNames)
	}
}

// toAll moves values from field values to "all".
func (vf *File) toAll() {
	delete(vf.all, "Tool")

	setObject(vf.RootPath, vf.all, rootPathNames, true)
	setObject(vf.Comment, vf.all, commentNames, false)
	setObject(vf.Ignore, vf.all, ignoreNames, false)

	rawPackageList := vf.getRawPackageList()

	setPkgFields := func(pkg *Package) {
		if pkg.Origin == pkg.Path {
			pkg.Origin = ""
		}
		if pkg.field == nil {
			pkg.field = make(map[string]interface{}, 10)
		}
		setObject(pkg.Origin, pkg.field, originNames, true)
		setObject(pkg.Path, pkg.field, pathNames, false)
		setObject(pkg.Tree, pkg.field, treeNames, true)
		setObject(pkg.Revision, pkg.field, revisionNames, false)
		setObject(pkg.RevisionTime, pkg.field, revisionTimeNames, true)
		setObject(pkg.Version, pkg.field, versionNames, true)
		setObject(pkg.VersionExact, pkg.field, versionExactNames, true)
		setObject(pkg.ChecksumSHA1, pkg.field, checksumSHA1Names, true)
		setObject(pkg.Comment, pkg.field, commentNames, true)
	}

	for i := len(vf.Package) - 1; i >= 0; i-- {
		pkg := vf.Package[i]
		switch {
		case pkg.Remove:
			for index, rawObj := range rawPackageList {
				raw, is := rawObj.(map[string]interface{})
				if !is {
					continue
				}
				same := true
				for key, value := range pkg.field {
					if raw[key] != value {
						same = false
						break
					}
				}
				if same {
					rawPackageList[index] = nil
				}
			}
		case pkg.Add:
			setPkgFields(pkg)
			rawPackageList = append(rawPackageList, pkg.field)
		default:
			if pkg.field == nil {
				pkg.field = make(map[string]interface{}, 10)
			}

			delete(pkg.field, "local")
			delete(pkg.field, "Local")
			setPkgFields(pkg)
		}
	}
	nextRawPackageList := make([]interface{}, 0, len(rawPackageList))
	for _, raw := range rawPackageList {
		if raw == nil {
			continue
		}
		nextRawPackageList = append(nextRawPackageList, raw)
	}
	vf.all[packageNames[0]] = nextRawPackageList
}

// Marshal the vendor file to the specified writer.
// Retains read fields.
func (vf *File) Marshal(w io.Writer) error {
	if vf.all == nil {
		vf.all = map[string]interface{}{}
	}
	vf.toAll()

	rawList := vf.getRawPackageList()
	sort.Sort(vendorPackageSort(rawList))

	jb, err := json.Marshal(vf.all)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	err = json.Indent(buf, jb, "", "\t")
	if err != nil {
		return err
	}
	_, err = io.Copy(w, buf)
	return err
}

// Unmarshal the vendor file from the specified reader.
// Stores internally all fields.
func (vf *File) Unmarshal(r io.Reader) error {
	bb, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	if vf.all == nil {
		vf.all = make(map[string]interface{}, 3)
	}
	err = json.Unmarshal(bb, &vf.all)
	if err != nil {
		return err
	}
	vf.toFields()
	return nil
}
