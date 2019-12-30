package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type MatchFile struct {
	Path      string
	Filename  string
	Extension string
	Contents  []byte
}

func NewMatchFile(path string) MatchFile {
	_, filename := filepath.Split(path)
	extension := filepath.Ext(path)
	contents, _ := ioutil.ReadFile(path)

	return MatchFile{
		Path:      path,
		Filename:  filename,
		Extension: extension,
		Contents:  contents,
	}
}


func isBinary(path string) bool { 
    f, err := os.Open(path)
    if err != nil{
        session.Log.Important("error opening path %s - %s",path,err)
        return false
    }
    buffer := make([]byte, 512)

    count, err := f.Read(buffer)
    if err != nil {
        session.Log.Important("error reading path %s",path)
        return false
    }
    err = f.Close()
    if err != nil {
        session.Log.Important("error closing path %s",path)
        return false
    }

    b := buffer[:count]

    for _, s := range b {
        if s == byte(0) {
            return true
        }
    }


    // Use the net/http package's handy DectectContentType function. Always returns a valid
    // content-type by returning "application/octet-stream" if no others seemed to match.

    return false
}


func IsSkippableFile(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))

	for _, skippableExt := range session.Config.BlacklistedExtensions {
		if extension == skippableExt {
			return true
		}
	}

    if *session.Options.ExcludeDirs != ""{
        if strings.Contains(path, *session.Options.ExcludeDirs){
            session.Log.Debug("Excluding path because of  exclude-dirs  %s",path)
            return true
        }
    }


    //Dont check or follow symlinks
    fi, err := os.Lstat(path)
    if err != nil{
         session.Log.Important("error lstatting path %s",path)
    }
    if fi.Mode() & os.ModeSymlink == os.ModeSymlink {
        session.Log.Debug("Skipping symlink with path %s",path)
        return true
    }


    if *session.Options.SkipBinaries == true{
        if isBinary(path) {
            session.Log.Debug("[DEBUG] Skipping binary file with path %s",path)
            return true
        }
    }


	for _, skippablePathIndicator := range session.Config.BlacklistedPaths {
		skippablePathIndicator = strings.Replace(skippablePathIndicator, "{sep}", string(os.PathSeparator), -1)
		if strings.Contains(path, skippablePathIndicator) {
			return true
		}
	}

	return false
}

func (match MatchFile) CanCheckEntropy() bool {
	if match.Filename == "id_rsa" {
		return false
	}

	for _, skippableExt := range session.Config.BlacklistedEntropyExtensions {
		if match.Extension == skippableExt {
			return false
		}
	}

	return true
}

func GetMatchingFiles(dir string) []MatchFile {
	fileList := make([]MatchFile, 0)
	maxFileSize := *session.Options.MaximumFileSize * 1024
    count := 0

	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil || f.IsDir() || uint(f.Size()) > maxFileSize || uint(f.Size()) == 0 || IsSkippableFile(path) {
			return nil
		}
        session.Log.Debug("Adding file %s",path)
		fileList = append(fileList, NewMatchFile(path))
        count +=1 
		return nil
	})

    session.Log.Important("added %d files",count)
	return fileList
}
