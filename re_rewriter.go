package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sync"
)

/*
All replacable variables.

    P = package name
    N = type name
    T = type 1
    .....
    Z = type 7
*/
const Placeholders = "PNTUVWXYZ"

var (
	placeholdersMap = func() (m map[byte]int) {
		m = make(map[byte]int, len(Placeholders))
		for i := range Placeholders {
			m[Placeholders[i]] = i
		}
		return
	}()

	placeholderRe   = regexp.MustCompile(`\b[` + Placeholders + `]\b`)
	pkgLineRe       = regexp.MustCompile(`(?m:^package.*?\n)`)
	ErrTooManyTypes = fmt.Errorf("too many types, allowed types are %s (%d)", Placeholders, len(Placeholders))
)

type ReTemplate struct {
	Path      string
	data      []byte
	phIdx     [][]int
	importIdx int
	lastMod   int64
	l         sync.Mutex
}

func NewReTemplate(b []byte) (*ReTemplate, error) {
	rt := &ReTemplate{
		data: b,
	}
	if err := rt.reload(); err != nil {
		return nil, err
	}
	return rt, nil
}

func NewReTemplateFile(path string) (*ReTemplate, error) {
	rt := &ReTemplate{
		Path: path,
	}
	if err := rt.reload(); err != nil {
		return nil, err
	}
	return rt, nil
}

func (rt *ReTemplate) reload() error {
	rt.l.Lock()
	defer rt.l.Unlock()
	if rt.Path == "" {
		if rt.phIdx != nil {
			return nil
		}
	} else {
		st, err := os.Stat(rt.Path)
		if err != nil {
			return err
		}
		lastMod := st.ModTime().Unix()
		if lastMod == rt.lastMod {
			return nil
		}
		b, err := ioutil.ReadFile(rt.Path)
		if err != nil {
			return err
		}
		rt.data = b
		rt.lastMod = lastMod
	}
	logger.Println("[ReTemplate] loading", rt.Path)
	rt.phIdx = placeholderRe.FindAllSubmatchIndex(rt.data, -1)
	rt.importIdx = pkgLineRe.FindSubmatchIndex(rt.data)[1]
	return nil
}

func (rt *ReTemplate) Output(w io.Writer, pi *PkgInfo) error {
	if err := rt.reload(); err != nil {
		return err
	}
	vals := append([]string{pi.Pkg, pi.Name}, pi.Types...)
	if len(vals) > len(Placeholders) {
		return ErrTooManyTypes
	}
	var (
		lidx int
	)
	wroteImports := len(pi.Imports) == 0
	for _, pidx := range rt.phIdx {
		if !wroteImports && pidx[1] > rt.importIdx {
			w.Write(rt.data[lidx:rt.importIdx])
			io.WriteString(w, "\nimport (\n")
			for _, imp := range pi.Imports {
				path := placeholderRe.ReplaceAllStringFunc(imp[1], func(s string) string {
					return vals[placeholdersMap[s[0]]]
				})
				io.WriteString(w, "\t"+imp[0]+" \""+path+"\"\n")
			}
			io.WriteString(w, ")\n")
			lidx = rt.importIdx
			wroteImports = true
		}
		w.Write(rt.data[lidx:pidx[0]])
		idx := placeholdersMap[rt.data[pidx[0]]]
		if idx < len(vals) {
			io.WriteString(w, vals[idx])
		}
		lidx = pidx[1]
	}
	w.Write(rt.data[lidx:])
	//placeholderRe.Split(s, n)
	return nil
}
