package backup

import (
        "bytes"
        "fmt"
        "io/ioutil"
        "log"
        "os"
        "path/filepath"
        "testing"
)

const (
        prefix  = "backup-"
        testdir = "testdata"
)

func die(t *testing.T, err error) {
        if err != nil {
                t.Fatal(err)
        }
}

func TestDir(t *testing.T) {
        dir := os.TempDir()
        b, err := IsFile(dir)
        die(t, err)
        if b {
                t.Fatalf("IsFile() reports directory %s as file\n", dir)
        }
}

func ExampleExt() {
        fmt.Println(Ext(4))
        // Output: .~4~
}

func ExampleDir() {
        fmt.Println(IsFile(os.TempDir()))
        // Output: false <nil>
}

func TestFile(t *testing.T) {
        filename := filepath.Join(os.TempDir(), "testfile")
        os.Remove(filename)
        if _, err := os.Create(filename); err != nil {
                t.Fatal(err)
        }
        defer os.Remove(filename)

        if b, err := IsFile(filename); err == nil {
                if !b {
                        t.Fatalf("IsFile() does not report %s as file\n",
                                filename)
                }
        } else {
                t.Fatal(err)
        }
}

func TestNoBackup(t *testing.T) {
        dir, err := ioutil.TempDir("", prefix)
        die(t, err)
        defer os.Remove(dir)

        filename := filepath.Join(dir, "not-there")
        _, err = os.Create(filename)
        want := 0
        got, err := Backups(filename)
        die(t, err)
        if want != got {
                t.Fatalf("want %d but got %d\n", want, got)
        }
}

func TestBackups(t *testing.T) {
        dir, err := ioutil.TempDir("", prefix)
        die(t, err)
        defer os.Remove(dir)

        filename := filepath.Join(dir, "file1")
        if _, err = os.Create(filename); err != nil {
                t.Fatal(err)
        }
        if _, err = os.Create(filepath.Join(dir, "file1"+Ext(1))); err != nil {
                t.Fatal(err)
        }
        if _, err = os.Create(filepath.Join(dir, "file1"+Ext(2))); err != nil {
                t.Fatal(err)
        }
        want := 2
        got, err := Backups(filename)
        die(t, err)
        if want != got {
                t.Fatalf("want %d but got %d\n", want, got)
        }
}

func TestFirstBackup(t *testing.T) {
        dir, err := ioutil.TempDir("", prefix)
        die(t, err)
        defer os.Remove(dir)

        filename := filepath.Join(dir, "file")
        _, err = os.Create(filename)
        die(t, err)
        Numbered(filename, 1)
        want := true
        cp := filename + ".~1~"
        got, err := Exists(cp)
        die(t, err)
        if want != got {
                t.Fatalf("want file %s to exist but it does not\n", cp)
        }
}

func TestLimit(t *testing.T) {
        dir, err := ioutil.TempDir("", prefix)
        die(t, err)
        defer os.Remove(dir)

        // dir must be empty
        expect := func(want int) {
                fis, err := ioutil.ReadDir(dir)
                die(t, err)
                got := len(fis)
                if want != got {
                        log.Fatalf("want %d files but got %+v\n", want, fis)
                }
        }

        // genesis: empty directory
        expect(0)

        filename := filepath.Join(dir, "file")
        _, err = os.Create(filename)
        die(t, err)
        // one original file
        expect(1)

        _, err = Numbered(filename, 1)
        die(t, err)
        // one original, one backup
        expect(2)

        // create second backup with limit == 1
        _, err = Numbered(filename, 1)
        if err == nil {
                t.Fatalf("want error but got nothing\n")
        }
        // one original, still one backup
        expect(2)
}

func TestNegativeLimit(t *testing.T) {
        _, got := Numbered("i-just-do-not-exist", -1)
        if got != nil {
                t.Fatalf("want nil but got %v\n", got)
        }
}

func TestMultipleSources(t *testing.T) {
        d, err := ioutil.TempDir("", prefix)
        die(t, err)
        defer os.Remove(d)

        // resolve files into tmp dir
        tmpfile := func(filename string) string {
                return filepath.Join(d, filename)
        }

        // resolve files into test directory
        testfile := func(filename string) string {
                return filepath.Join(testdir, filename)
        }

        f := "12.txt"
        wantfile := testfile(f)
        gotfile := tmpfile(f)
        src1 := testfile("1.txt")
        src2 := testfile("2.txt")
        if err := Copy(gotfile, src1, src2); err != nil {
                t.Fatal(err)
        }
        // compare same file in tmp and test dir
        want, err := ioutil.ReadFile(wantfile)
        if err != nil {
                t.Fatal(err)
        }
        got, err := ioutil.ReadFile(gotfile)
        if err != nil {
                t.Fatal(err)
        }
        if !bytes.Equal(want, got) {
                t.Fatalf("want %s but got %s\n", string(want), string(got))
        }
}

