// Backup mimics the GNU cp --backup=numbered command in pure Go

package backup

import (
        "errors"
        "fmt"
        "io"
        "io/ioutil"
        "os"
        "path/filepath"
)

// Backups returns number of existing consecutive backups for given filename.
func Backups(filename string) (int, error) {
        if b, err := IsFile(filename); err == nil {
                if !b {
                        // TODO type
                        return -1, errors.New("not a file: " + filename)
                }
        } else {
                return -1, err
        }
        // NOTSURE read all entries in dir instead of distinct stat()s
        dir := filepath.Dir(filename)
        fis, err := ioutil.ReadDir(dir)
        if err != nil {
                return -1, err
        }

        // store in Set
        m := make(map[string]bool)
        for _, fi := range fis {
                m[fi.Name()] = true
        }

        basename := filepath.Base(filename)
        for i := 1; ; i++ {
                if !m[basename+Ext(i)] {
                        return i - 1, nil
                }
        }
        return 0, nil
}

// Copy copies any number of files into destination file.
// Same behaviour as GNU cp.
// Destination file will be deleted in case of any errors during copying.
func Copy(destination string, sources ...string) error {
        dst, err := os.Create(destination)
        if err != nil {
                return err
        }
        dc := func() {
                dst.Close()
                os.Remove(destination)
        }
        for _, source := range sources {
                // open read only
                src, err := os.Open(source)
                if err != nil {
                        dc()
                        return err
                }
                _, err = io.Copy(dst, src)
                if err != nil {
                        src.Close()
                        dc()
                        return err
                }
                if err := src.Close(); err != nil {
                        dc()
                        return err
                }
        }
        return dst.Close()
}

// Exists returns true if given filename exists, false otherwise.
// Works for any kind of file type such as files, directories, links, ...
func Exists(filename string) (bool, error) {
        _, err := os.Stat(filename)
        if err == nil {
                return true, nil
        }
        if os.IsNotExist(err) {
                return false, nil
        }
        var mu bool
        return mu, err
}

// Ext returns backup extension for generation n.
func Ext(n int) string {
        return fmt.Sprintf(".~%d~", n)
}

// IsFile returns true for files, and false for directories, links and the like
func IsFile(filename string) (bool, error) {
        fi, err := os.Stat(filename)
        if err != nil {
                var mu bool
                return mu, err
        }
        return fi.Mode().IsRegular(), nil
}

// Numbered will create a numbered backup copy of the given file.
// Filename is taken as is, no fancy logic such as absolute path resolution or
// following symlink is taking place.
// The copy will reside in the same directory.
// Number of backups can be limited, creating a backup copy when the limit is
// reached will return an error.
// Negative limits will not do anything at all.
// This operation is file based and therefore probably not atomic, meaning there
// is an implicit race condition between checking for the next available backup
// number, and the copying of the file.
// The working directory will not be locked.
// Negative limits and 0 are considered a NOP.
//
//   Numbered("/etc/ssh/ssh_config", 1)
//
// will create /etc/ssh/ssh_config.~1~, running the same again will not create a
// numbered backup because limit 1 is reached.
// Returns the newly created filename.
func Numbered(filename string, limit int) (string, error) {
        if limit < 1 {
                return "", nil
        }
        n, err := Backups(filename)
        if err != nil {
                return "", err
        }
        if n >= limit {
                msg := fmt.Sprintf("existing backups (%d) exceed limit (%d)",
                        n, limit)
                return "", errors.New(msg)
        }
        // slot into next generation
        next := filename + Ext(n+1)
        return next, Copy(next, filename)
}

