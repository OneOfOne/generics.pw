package main

import (
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/OneOfOne/generics.pw/memgit"
)

type TemplateManager struct {
	path string
	m    map[string]*ReTemplate
	git  map[string]*memgit.MemoryGit
	l    sync.Mutex
}

func NewTemplateManager(path string) (*TemplateManager, error) {
	pathes, err := filepath.Glob(filepath.Join(path, "*.go"))
	if err != nil {
		return nil, err
	}
	tm := &TemplateManager{
		path: path,
		m:    make(map[string]*ReTemplate, len(pathes)),
	}
	for _, p := range pathes {
		tname := filepath.Base(p)
		if _, err = tm.load(tname[:len(tname)-3]); err != nil {
			return nil, err
		}
	}
	return tm, nil
}

func (tm *TemplateManager) load(tname string) (rt *ReTemplate, err error) {
	var ok bool
	rt, ok = tm.m[tname]
	if !ok {
		rt, err = NewReTemplateFile(filepath.Join(tm.path, tname+".go"))
		if err != nil {
			return
		}
		tm.m[tname] = rt
	}
	return
}

func (tm *TemplateManager) GitHandler(path string, pi *PkgInfo) (mg *memgit.MemoryGit, err error) {
	tm.l.Lock()
	defer tm.l.Unlock()
	if tm.git == nil {
		tm.git = make(map[string]*memgit.MemoryGit, 1)
	}
	if mg = tm.git[path]; mg == nil {
		buf := getBuffer()
		rt, err := tm.load(pi.tmpl)
		if err != nil {
			return nil, err
		}
		mg = memgit.New()
		rt.Output(buf, pi)
		mg.AddFile(pi.tmpl+".go", buf.Bytes())
		_, c := mg.Commit(path, rt.lastMod)
		logger.Println("[manager] created commit", c, "for", path)
		tm.git[path] = mg
		time.AfterFunc(time.Second*10, func() {
			tm.l.Lock()
			delete(tm.git, path)
			tm.l.Unlock()
		})
		putBuffers(buf)
	}

	return
}

func (tm *TemplateManager) Output(w io.Writer, pi *PkgInfo) error {
	tm.l.Lock()
	defer tm.l.Unlock()
	tname := strings.ToLower(pi.Name)
	rt, err := tm.load(tname)
	if err != nil {
		return err
	}
	return rt.Output(w, pi)
}
