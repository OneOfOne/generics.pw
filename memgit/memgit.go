package memgit

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	nulByte       = string(0)
	headObj       = "ref: refs/heads/master\n"
	infoObjSuffix = "\trefs/heads/master\n"

	defaultTime = 1412713536
	dummyCommit = `tree %s%s
author MemoryGit <memgit@generics.pw> %d +0000
committer MemoryGit <memgit@generics.pw> %d +0000

%s
`
)

type hashWriter struct {
	w io.WriteCloser
	h hash.Hash
}

func (hw *hashWriter) Write(p []byte) (n int, err error) {
	hw.h.Write(p)
	return hw.w.Write(p)
}

func (hw *hashWriter) Close() error {
	return hw.w.Close()
}
func (hw *hashWriter) Reset(buf *bytes.Buffer) {
	hw.w, _ = zlib.NewWriterLevel(buf, 9)
	hw.h = sha1.New()
}

func (hw *hashWriter) Hash() ([]byte, string) {
	h := hw.h.Sum(nil)
	hh := hex.EncodeToString(h)
	return h, hh[:2] + "/" + hh[2:]
}

type MemoryGit struct {
	fs map[string]string

	commit string

	b1, b2 *bytes.Buffer
}

// New returns an in-memory simple git repo
func New() *MemoryGit {
	return &MemoryGit{
		b1: getBuffer(),
		b2: getBuffer(),
	}
}

// AddFile adds a file to the repo.
// To pass git fsck, all the file names must be sorted.
func (mg *MemoryGit) AddFile(name string, data []byte) (path string) {
	if mg.fs == nil {
		mg.fs = make(map[string]string, 3) // we will always have a minimum of 2 objects + 1 per file
	}
	if mg.b1 == nil {
		panic("called AddFile after already saving the commit")
	}
	var (
		hw    = &hashWriter{}
		fhash []byte
	)

	mg.b1.Reset()
	hw.Reset(mg.b1)

	fmt.Fprint(hw, "blob ", len(data), nulByte)
	hw.Write(data)
	hw.Close()

	fhash, path = hw.Hash()
	mg.fs[path] = mg.b1.String()

	fmt.Fprint(mg.b2, "100644 ", name, nulByte)
	mg.b2.Write(fhash)
	return
}

// Commit commits all the added files to the repo with the specific message and time.
// Commit will automatically be called if you call HttpHandler().
func (mg *MemoryGit) Commit(msg string, commitTime int64) (treepath, commitpath string) {
	if mg.fs == nil {
		panic("trying to commit an empty repo")
	}
	if mg.b1 == nil {
		panic("called commit twice")
	}
	hw := &hashWriter{}
	mg.b1.Reset()
	hw.Reset(mg.b1)

	fmt.Fprint(hw, "tree ", mg.b2.Len(), nulByte)
	mg.b2.WriteTo(hw)
	hw.Close()

	_, treepath = hw.Hash()
	mg.fs[treepath] = mg.b1.String()

	mg.b1.Reset()
	mg.b2.Reset()
	hw.Reset(mg.b1)

	//	fmt.Fprint(mg.b2, "tree ", treepath[8:10], treepath[11:], "\n")
	fmt.Fprintf(mg.b2, dummyCommit, treepath[:2], treepath[3:], commitTime, commitTime, msg)

	fmt.Fprint(hw, "commit ", mg.b2.Len(), nulByte)
	mg.b2.WriteTo(hw)
	hw.Close()

	_, commitpath = hw.Hash()

	mg.fs[commitpath] = mg.b1.String()

	mg.commit = commitpath[:2] + commitpath[3:]

	putBuffers(mg.b1, mg.b2)
	mg.b1, mg.b2 = nil, nil
	return
}

func (mg *MemoryGit) getObject(path string) string {
	path = strings.TrimPrefix(path, "/objects/")
	return mg.fs[path]
}

// HttpHandler returns an http.Handler to serve the repo over http.
// It will automatically strip the prefix from r.URL.Path.
func (mg *MemoryGit) HttpHandler(prefix string, l *log.Logger) http.Handler {
	if mg.commit == "" && mg.fs != nil {
		mg.Commit("default", defaultTime)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, prefix)
		st := 200
		switch p {
		case "/info/refs":
			l.Println("[memgit]", p, "[", r.UserAgent(), "]")
			io.WriteString(w, mg.commit)
			io.WriteString(w, infoObjSuffix)
		case "/HEAD":
			io.WriteString(w, headObj)
		default:
			obj := mg.getObject(p)
			if obj != "" {
				io.WriteString(w, obj)
			} else {
				st = http.StatusNotFound
				http.Error(w, "not found", st)
			}
		}
		// if l != nil {
		// 	l.Println("[MemoryGit]", st, r.Method, p, "[", r.UserAgent(), "]")
		// }
	})
}

func dump(r io.Reader) {
	var t struct {
		Offset                         [6]byte
		Flags                          uint16
		CompLen, UncompLen             uint32
		BaseRev, LinkRev, P1Rev, P2Rev uint32
		NodeID                         [32]byte
	}
	binary.Read(r, binary.LittleEndian, &t)
	fmt.Printf("%#v\n", t)
}

func main() {
	b, _ := ioutil.ReadFile("/tmp/int/set.go")
	// fg := NewFakeGit("set.go", b)
	// for k := range fg.fs {
	// 	fmt.Println(k)
	// }
	mg := New()
	fmt.Println("mg: ", mg.AddFile("date", []byte("test")))
	fmt.Println("mg: ", mg.AddFile("set.go", b))
	fmt.Println(mg.Commit("hello", time.Now().Unix()))
	//	fmt.Println(http.ListenAndServe(":9020", fg))
	fmt.Println(http.ListenAndServe(":9020", mg.HttpHandler("", log.New(os.Stdout, "", 0))))
}
