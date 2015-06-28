package utils

import (
	"bufio"
	"crypto/md5"
	"fmt"
	log "github.com/cihub/seelog"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ListFilesInDir(doc_path string) []FsItem {
	searchDir := doc_path

	fsItems := []FsItem{}

	err := filepath.Walk(searchDir, func(doc_path string, info os.FileInfo, err error) error {
		//add to slice here
		if MatchesIgnore(searchDir, doc_path) {
			log.Debugf("Ignoring file: %s", doc_path)
		} else {
			fsItems = append(fsItems, FsItem{
				Filename: doc_path,
				IsDir:    info.IsDir(),
				Checksum: GetMd5Checksum(doc_path),
				Mtime:    info.ModTime().UTC(),
				Perms:    fmt.Sprintf("%#o", info.Mode().Perm()),
			})
		}
		return nil
	})
	if err != nil {
		checkErr(err, "Unable to read file info")
	}

	return fsItems
}

func MatchesIgnore(directory_path, filename string) bool {
	var MatchFilename = ""

	if strings.Contains(filename, "/") {
		MatchFilename = strings.TrimPrefix(filename, directory_path + "/")
	} else {
		MatchFilename = filename
	}
    var MatchFilenameAll = filepath.Base(MatchFilename)
	exclude_file := directory_path + "/.goignore"
	if _, err := os.Stat(exclude_file); os.IsNotExist(err) {
		log.Infof("No ignore file found (%s)", exclude_file)
		return false
	} else {
		file, err := os.Open(exclude_file)
		if err != nil {
			log.Infof("Could not read exclude file: %s\nError: %s", exclude_file, err.Error())
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if err := scanner.Err(); err != nil {
			log.Critical("Error reading lines in exclude file (%s): %s", exclude_file, err.Error())
		}
		for scanner.Scan() {
			pattern := scanner.Text()
			if pattern != "" {
				match, err := filepath.Match(scanner.Text(), MatchFilename)
				log.Debugf("Match (%s) => (%s) = (%v)", MatchFilename, pattern, match)
                if err != nil {
					log.Debugf("Invalid patern %s", scanner.Text())
				}
                match2, err := filepath.Match(scanner.Text(), MatchFilenameAll)
                if err != nil {
                    log.Debugf("Invalid patern %s", scanner.Text())
                }
                if match || match2{
                    return true
                }

			}

		}
		return false
	}
}

func GetFileInfo(doc_path string) (FsItem, error) {
	fi, err := os.Stat(doc_path)
	var item = FsItem{}
	if err != nil {
		return item, err
	} else {
		item.Filename = doc_path
		item.IsDir = fi.IsDir()
		item.Checksum = GetMd5Checksum(doc_path)
		item.Mtime = fi.ModTime().UTC()
		item.Perms = fmt.Sprintf("%#o", fi.Mode().Perm())
	}
	return item, nil
}

func GetMd5Checksum(filepath string) string {
	if b, err := computeMd5(filepath); err == nil {
		md5string := fmt.Sprintf("%x", b)
		return md5string
	} else {
		return "DirectoryMD5"
	}
}

func computeMd5(filepath string) (string, error) {
	const filechunk = 8192
	file, err := os.Open(filepath)

	if err != nil {
		return "", err
	}

	defer file.Close()

	// calculate the file size
	info, _ := file.Stat()

	filesize := info.Size()

	blocks := uint64(math.Ceil(float64(filesize) / float64(filechunk)))

	hash := md5.New()

	for i := uint64(0); i < blocks; i++ {
		blocksize := int(math.Min(filechunk, float64(filesize-int64(i*filechunk))))
		buf := make([]byte, blocksize)

		file.Read(buf)
		io.WriteString(hash, string(buf)) // append into the hash
	}

	//fmt.Printf("%s checksum is %x\n",file.Name(), hash.Sum(nil))
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func checkErr(err error, msg string) {
	if err != nil {
		//log.Fatalln(msg, err)
	}
}

// Fetches the home directory of the listener from the config file
func GetListenerHomeDir(listener string) string {
	cfg := GetConfig()
	return cfg.Listeners[listener].Directory
}

// Fetches the base directory (upload path) from configuration file
func GetListenerUploadPath(listener string) string {
	cfg := GetConfig()
	return cfg.Listeners[listener].BasePath
}

// Gets the relative path based on the home directory, essentially stripping the listener,
// home directory from the absolute path
func GetRelativePath(listener, local_path string) string {
	return strings.TrimPrefix(local_path, GetListenerHomeDir(listener))
}

// Gets the relative path for the item by trimming the absolute path with the home dir,
// it then appends the base directory (Upload Path) to the front of the relative path.
func GetRelativeUploadPath(listener, local_path string) string {
	return GetListenerUploadPath(listener) + GetRelativePath(listener, local_path)
}

func GetAbsPath(listener, db_path string) string {
	absPath := GetListenerHomeDir(listener) + db_path
	//logs.LogWriteF("=====> Absolute Path: %s <=====", absPath)
	return absPath
}

func GetListenerFromDir(dir string) string {
	cfg := GetConfig()
	var listener = ""
	for lname, ldata := range cfg.Listeners {
		if ldata.Directory == dir {
			listener = lname
		}
	}
	return listener
}

func GetSystemHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

// exists returns whether the given file or directory exists or not
func ItemExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func FileWrite(path string, r io.Reader, uid, gid int, perms string) (int64, error) {
	iPerm, _ := strconv.Atoi(perms)
	mode := int(iPerm)
	w, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(mode))
	if err != nil {
		if path == "" {
			w = os.Stdout
		} else {
			return 0, err
		}
	}
	defer w.Close()

	size, err := io.Copy(w, r)

	if err != nil {
		return 0, err
	}
	err = w.Chown(uid, gid)
	return size, err
}
