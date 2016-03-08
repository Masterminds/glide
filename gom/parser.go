package gom

// This is copied + slightly adapted from gom's `gomfile.go` file.
//
// gom's license is MIT-style.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var qx = `'[^']*'|"[^"]*"`
var kx = `:[a-z][a-z0-9_]*`
var ax = `(?:\s*` + kx + `\s*|,\s*` + kx + `\s*)`
var re_group = regexp.MustCompile(`\s*group\s+((?:` + kx + `\s*|,\s*` + kx + `\s*)*)\s*do\s*$`)
var re_end = regexp.MustCompile(`\s*end\s*$`)
var re_gom = regexp.MustCompile(`^\s*gom\s+(` + qx + `)\s*((?:,\s*` + kx + `\s*=>\s*(?:` + qx + `|\s*\[\s*` + ax + `*\s*\]\s*))*)$`)
var re_options = regexp.MustCompile(`(,\s*` + kx + `\s*=>\s*(?:` + qx + `|\s*\[\s*` + ax + `*\s*\]\s*)\s*)`)

func unquote(name string) string {
	name = strings.TrimSpace(name)
	if len(name) > 2 {
		if (name[0] == '\'' && name[len(name)-1] == '\'') || (name[0] == '"' && name[len(name)-1] == '"') {
			return name[1 : len(name)-1]
		}
	}
	return name
}

func parseOptions(line string, options map[string]interface{}) {
	ss := re_options.FindAllStringSubmatch(line, -1)
	re_a := regexp.MustCompile(ax)
	for _, s := range ss {
		kvs := strings.SplitN(strings.TrimSpace(s[0])[1:], "=>", 2)
		kvs[0], kvs[1] = strings.TrimSpace(kvs[0]), strings.TrimSpace(kvs[1])
		if kvs[1][0] == '[' {
			as := re_a.FindAllStringSubmatch(kvs[1][1:len(kvs[1])-1], -1)
			a := []string{}
			for i := range as {
				it := strings.TrimSpace(as[i][0])
				if strings.HasPrefix(it, ",") {
					it = strings.TrimSpace(it[1:])
				}
				if strings.HasPrefix(it, ":") {
					it = strings.TrimSpace(it[1:])
				}
				a = append(a, it)
			}
			options[kvs[0][1:]] = a
		} else {
			options[kvs[0][1:]] = unquote(kvs[1])
		}
	}
}

type Gom struct {
	name    string
	options map[string]interface{}
}

func parseGomfile(filename string) ([]Gom, error) {
	f, err := os.Open(filename + ".lock")
	if err != nil {
		f, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
	}
	br := bufio.NewReader(f)

	goms := make([]Gom, 0)

	n := 0
	skip := 0
	valid := true
	var envs []string
	for {
		n++
		lb, _, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				return goms, nil
			}
			return nil, err
		}
		line := strings.TrimSpace(string(lb))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		name := ""
		options := make(map[string]interface{})
		var items []string
		if re_group.MatchString(line) {
			envs = strings.Split(re_group.FindStringSubmatch(line)[1], ",")
			for i := range envs {
				envs[i] = strings.TrimSpace(envs[i])[1:]
			}
			valid = true
			continue
		} else if re_end.MatchString(line) {
			if !valid {
				skip--
				if skip < 0 {
					return nil, fmt.Errorf("Syntax Error at line %d", n)
				}
			}
			valid = false
			envs = nil
			continue
		} else if skip > 0 {
			continue
		} else if re_gom.MatchString(line) {
			items = re_gom.FindStringSubmatch(line)[1:]
			name = unquote(items[0])
			parseOptions(items[1], options)
		} else {
			return nil, fmt.Errorf("Syntax Error at line %d", n)
		}
		if envs != nil {
			options["group"] = envs
		}
		goms = append(goms, Gom{name, options})
	}
	return goms, nil
}