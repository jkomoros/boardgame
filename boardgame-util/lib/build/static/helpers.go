package static

import (
	"errors"
	"github.com/jkomoros/boardgame/boardgame-util/lib/path"
	"io/ioutil"
	"os"
	"path/filepath"
)

func absoluteStaticServerPath() (string, error) {

	pth, err := path.AbsoluteGoPkgPath(mainPackage)

	if err != nil {
		return "", errors.New("Couldn't load main boardgame package location: " + err.Error())
	}

	return filepath.Join(pth, staticServerPath), nil

}

//staticBuildDir returns the static build directory within dir, creating it
//if it doesn't exist. For example, for dir="temp", returns "temp/static".
func staticBuildDir(dir string) (string, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", errors.New(dir + " did not already exist.")
	}

	staticDir := filepath.Join(dir, staticSubFolder)

	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		if err := os.Mkdir(staticDir, 0700); err != nil {
			return "", errors.New("Couldn't create static directory: " + err.Error())
		}
	}

	return staticDir, nil
}

//copyFile copies the file at location remote to location local, copying
//cotents and perms.
func copyFile(remote, local string) error {

	info, err := os.Stat(remote)

	if err != nil {
		return errors.New("Couldn't get info for remote: " + err.Error())
	}

	contents, err := ioutil.ReadFile(remote)

	if err != nil {
		return errors.New("Couldn't read file " + remote + ": " + err.Error())
	}

	if err := ioutil.WriteFile(local, contents, info.Mode()); err != nil {
		return errors.New("Couldn't write file: " + err.Error())
	}

	return nil

}

//buildCachePath returns where we store our build cache (or where we WOULD if
//it existed).
func buildCachePath() (string, error) {
	userCacheDir, err := os.UserCacheDir()

	if err != nil {
		return "", errors.New("Couldn't get usercachedir: " + err.Error())
	}

	return filepath.Join(userCacheDir, nodeModulesCacheDir), nil
}
