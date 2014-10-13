// build !client

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	importPathFmt = `<meta name="go-import" content="%s git %s://%s">%c`
	tmplPath      = "tmpl/"
)

func getStr(k, def string) string {
	if k == "" {
		return def
	}
	return k
}

var (
	tm = func() *TemplateManager {
		tm, err := NewTemplateManager(tmplPath)
		if err != nil {
			panic(err)
		}
		return tm
	}()

	logger = log.New(os.Stdout, "generics.pw ", log.LstdFlags)

	local    = flag.Bool("local", false, "run locally")
	https    = flag.Bool("https", true, "use https for redirects")
	hostname = flag.String("host", "generics.pw", "hostname")
)

func init() {
	flag.Parse()
	if *local {
		*https = false
		*hostname = "lfx.x64.me"
	}
}

func writeImportPath(w io.Writer, p string) {
	scheme := "http"
	if *https {
		scheme = "https"
	}
	url := *hostname + "/t/" + p
	fmt.Fprintf(w, importPathFmt, url, scheme, url, '\n')
}

func cleanPath(p string) (prefix, path string) {
	ln := len(p)
	switch {
	case strings.HasSuffix(p, "/info/refs"):
		prefix = p[:ln-10]
		path = prefix[3:] //strip /t/
	case strings.HasSuffix(p, "/HEAD"):
		prefix = p[:ln-5]
		path = prefix[3:]
	case ln > 50 && p[ln-49:ln-42] == "objects":
		prefix = p[:ln-50]
		path = prefix[3:]
	default:
		path = p[3:]
	}
	return
}

func tHandler(w http.ResponseWriter, r *http.Request) {
	pre, p := cleanPath(r.URL.Path)
	//logger.Printf("[/t/] %s %q %q %q %q %q", r.Method, pre, p, r.UserAgent(), r.RemoteAddr, r.URL.RawQuery)
	dotGo := strings.HasSuffix(p, ".go")
	if dotGo {
		p = p[:len(p)-3]
	}
	switch {
	case r.URL.Query().Get("go-get") == "1":
		writeImportPath(w, p)
		return
	}
	isgit := strings.HasPrefix(r.Header.Get("User-Agent"), "git")
	//	logger.Printf("%#v", NewPkgInfo(p, isgit))
	pi := NewPkgInfo(p, isgit)
	switch {
	case pi == nil:
		w.WriteHeader(400)
	case isgit:
		mg, err := tm.GitHandler(p, pi)
		if err != nil {
			w.WriteHeader(404)
			logger.Printf("[tm] error: %v", err)
			return
		}
		mg.HttpHandler(pre, logger).ServeHTTP(w, r)
	default:
		logger.Println("default")
		w.Header().Set("Content-Type", "text/go; charset=utf-8")
		if dotGo {
			fullFn := strings.ToLower(pi.Name + "_" + strings.Join(pi.Types, "_") + ".go")
			w.Header().Set("Content-Disposition", "attachment; filename="+fullFn)
		}
		if err := tm.Output(w, pi); err != nil {
			w.WriteHeader(404)
			logger.Printf("[tm] error: %v", err)
		}
	}
}

func rootHandler(static string) func(w http.ResponseWriter, req *http.Request) {
	h := http.FileServer(http.Dir(static))
	validExt := regexp.MustCompile(`\.(?:js|css|gz|html|png|jpg)$`)
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			f, err := os.Open("index.html")
			if err != nil {
				logger.Println("error opening index.html:", err)
				return
			}
			io.Copy(w, f)
			f.Close()
			return
		}
		fmt.Println(req.URL)
		if !validExt.MatchString(req.URL.Path) {
			http.Error(w, "not found", 404)
			return
		}
		h.ServeHTTP(w, req)
	}
}

func main() {
	http.HandleFunc("/t/", tHandler)
	http.HandleFunc("/", rootHandler("./static"))
	addr := "127.0.0.1:9020"
	if *local {
		addr = ":80"
	}
	logger.Println("listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Fatal(err)
	}
}
