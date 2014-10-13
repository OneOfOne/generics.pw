package memgit

import "testing"

var expected = [...]string{
	"42/4071bba62ccadb9e2ceafd28b7ff169c25e3cf",
	"00/5cf41666a2ea2a5f006375428b46cf9d2a15f1",
	"33/d793d1c7c19ef16dee03ebaf6d4ffbe3ea5d96",
}

func Test(t *testing.T) {
	mg := New()
	fid := mg.AddFile("dummy.txt", []byte("dummies"))
	tid, cid := mg.Commit("dummy commit message", defaultTime)
	t.Logf("dummy repo:\nfile  : %s\ntree  : %s\ncommit: %s", fid, tid, cid)
	for _, v := range expected {
		if _, ok := mg.fs[v]; !ok {
			t.Fatalf("couldn't find %q", v)
		}
	}
}
