package main

import (
	"fmt"
	"regexp"
	"strings"
)

func capitalize(n string) string {
	if n == "" {
		return n
	}
	return strings.ToUpper(n[:1]) + n[1:]
}

var knownHosts = map[string]string{
	"g":   "generics.pw/t",
	"gc":  "code.google.com/p",
	"gcg": "code.google.com/p/go.",
	"gh":  "github.com",
	"bb":  "bitbucket.org",
	"lp":  "launchpad.net",
	"gp":  "gopkg.in",
}

type PkgInfo struct {
	Name       string
	Pkg        string
	Types      []string
	Imports    [][2]string
	Comparable bool
}

func (pi *PkgInfo) String() string {
	return fmt.Sprintf("%#v", pi)
}

// /t/tmpl-name[=typename,package-name,cmp]/type1[=import-package],type2[=importpackage]
var (
	reSplitParts = regexp.MustCompile(`([^=/]+)(?:=([^/]+))?`)
	reImportName = regexp.MustCompile(`\*?([^.]+)\.`)
	// 1 = name, 2 = import/packagename, 3 = options
)

func strOrDef(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func NewPkgInfo(path string, git bool) (pi *PkgInfo) {
	p := reSplitParts.FindAllStringSubmatch(path, -1)
	if len(p) < 2 {
		return
	}
	pi = &PkgInfo{
		Name: capitalize(p[0][1]),
		Pkg:  "main",
	}
	opts := strings.Split(p[0][2], ",")
	if len(opts) > 0 && opts[0] != "" {
		pi.Name = opts[0]
	}
	if len(opts) > 1 && opts[1] != "" {
		pi.Pkg = opts[1]
	}
	if len(opts) > 2 && opts[2] == "opt" {
		pi.Comparable = true
	}

	if git && pi.Pkg == "main" {
		pi.Pkg = strings.ToLower(pi.Name)
	}
	for _, ti := range p[1:] {
		pi.Types = append(pi.Types, ti[1])
		if ti[2] != "" {
			ipath := strings.Split(ti[2], ":")
			if len(ipath) > 1 {
				host := ipath[0]
				ipath[0] = strOrDef(knownHosts[host], host)
			}
			importInfo := [2]string{".", strings.Join(ipath, "/")}
			if pname := reImportName.FindStringSubmatch(ti[1]); pname != nil {
				importInfo[0] = pname[1]
			}
			pi.Imports = append(pi.Imports, importInfo)
		}
	}
	return
}

/*
func getPkgInfo(path string, git bool) (tname string, pi *PkgInfo) {
	var (
		pname string
	)
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return
	}
	tname, pname = parts[0], ""
	if idx := strings.Index(tname, "="); idx > 0 {
		pname = tname[idx+1:]
		tname = tname[:idx]
	} else if git {
		pname = tname
	}
	pi = &PkgInfo{
		Pkg:   getStr(pname, "main"),
		Name:  capitalize(tname),
		Types: parts[1:],
	}
	return
}
*/
