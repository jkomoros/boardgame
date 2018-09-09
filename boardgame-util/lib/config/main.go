/*

config is a simple library that manages config set-up for boardgame-util and
friends, reading from config.json and config.SECRET.json files. See boardgame-
util/README.md for more on the structure of config.json files.

Although a number of the details are exposed in this package, generally you
just use Get() and then directly read the values of the returned Config's Dev
and Prod properties.

*/
package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	privateConfigFileName = "config.SECRET.json"
	publicConfigFileName  = "config.PUBLIC.json"
)

//DefaultFileNames returns the publicConfig and privateConfig names for the
//given path, even if they don't exist. If dirOrFile ends in ".json" then that
//will be returned, with privateConfig being in the same folder. If it's a
//dir, it will be the default filenames in that folder.
func DefaultFileNames(dirOrFile string) (publicConfig, privateConfig string, err error) {

	if dirOrFile == "" {
		dirOrFile = "."
	}

	if strings.HasSuffix(dirOrFile, ".json") {
		dir := filepath.Dir(dirOrFile)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return "", "", errors.New("Dir " + dir + " does not exist.")
		}

		return dirOrFile, filepath.Join(dir, privateConfigFileName), nil
	}

	//OK, we'll interpret as Dir.

	if _, err := os.Stat(dirOrFile); os.IsNotExist(err) {
		return "", "", errors.New(dirOrFile + " is interpreted as a directory but does not exist")
	}
	return filepath.Join(dirOrFile, publicConfigFileName), filepath.Join(dirOrFile, privateConfigFileName), nil
}

//FileNames returns the publicConfig filename and privateConfig filename to
//use given the search path. If dir is a config file itself, loads that (and
//any private component in same directory). Next it interprets dir as a
//directory to search within for any config files. If none are found, walks
//upwards in the directory hierarchy (as long as that's still in $GOPATH)
//until it finds a folder that appears to work. If dir is "", working
//directory is assumed. If skipUpwardSearch is true, then if dir is non-blank
//upward searching in the dir hiearchy will not be done.
func FileNames(dir string, skipUpwardSearch bool) (publicConfig, privateConfig string, err error) {

	if dir == "" {
		dir = "."
		skipUpwardSearch = false
	}

	//Try to interpret it as a file
	if public, private, err := fileNamesToUseWithFile(dir); err == nil {
		return public, private, nil
	}

	//Guess it wasn't a file, try interpreting as a directory.

	for {

		if absDir, err := filepath.Abs(dir); err != nil {
			//Must have fallen off the end at the top.
			break
		} else if absDir == string(filepath.Separator) {
			break
		}

		public, private := fileNamesToUseInDir(dir)

		if public != "" || private != "" {
			return public, private, nil
		}

		if skipUpwardSearch {
			break
		}

		dir = filepath.Join("..", dir)
	}

	return "", "", errors.New("Couldn't find a path")

}

//fileNamesToUseWithFile takes a filename of the public component. Returns the
//string to the publicComponent and also the private component if it exists in
//that folder.z
func fileNamesToUseWithFile(filename string) (publicConfig, privateConfig string, err error) {

	if info, err := os.Stat(filename); err != nil {
		return "", "", errors.New("That file does not exist: " + err.Error())
	} else {
		if info.IsDir() {
			return "", "", errors.New(filename + " points to a dir, not a file")
		}
	}

	//Check to see if there's a private config in that folder
	dir := filepath.Dir(filename)

	privatePath := filepath.Join(dir, privateConfigFileName)

	if _, err := os.Stat(privatePath); err != nil {
		// No private path I guess
		return filename, "", nil
	}

	return filename, privatePath, nil

}

//fileNamesToUseInDir looks for public/private values precisely in the given folder.
func fileNamesToUseInDir(dir string) (publicConfig, privateConfig string) {

	possiblePrivateConfig := filepath.Join(dir, privateConfigFileName)

	if _, err := os.Stat(possiblePrivateConfig); err == nil {
		privateConfig = possiblePrivateConfig
	}

	infos, err := ioutil.ReadDir(dir)

	if err != nil {
		return "", ""
	}

	foundNames := make(map[string]bool)

	for _, info := range infos {
		if info.Name() == privateConfigFileName {
			continue
		}
		if strings.HasPrefix(info.Name(), "config.") && strings.HasSuffix(info.Name(), ".json") {
			foundNames[info.Name()] = true
		}
	}

	prioritizedNames := []string{
		publicConfigFileName,
		"config.json",
	}

	for _, name := range prioritizedNames {
		if foundNames[name] {
			return filepath.Join(dir, name), privateConfig
		}
	}

	//Whatever, return the first one
	for name := range foundNames {
		return filepath.Join(dir, name), privateConfig
	}

	//None of the preferred names were found, just return whatever is in
	//publicConfig, privateConfig. publicConfig is "", privateConfig already
	//has the dir in it
	return

}

//GetConfig returns a Config for those two named files. publicConfig and
//privateConfig may both be "" without erroring. If createIfNotExist is true,
//then NewRawConfig will be told to create the Configs even if they don't
//exist on disk.
func GetConfig(publicConfigFile, privateConfigFile string, createIfNotExist bool) (*Config, error) {
	publicConfig, err := NewRawConfig(publicConfigFile, createIfNotExist)

	if err != nil {
		return nil, errors.New("Couldn't get public config: " + err.Error())
	}

	privateConfig, err := NewRawConfig(privateConfigFile, createIfNotExist)

	if err != nil {
		return nil, errors.New("Couldn't get private config: " + err.Error())
	}

	return NewConfig(publicConfig, privateConfig), nil
}

//Get fetches a fully realized config. It is a simple convenience wrapper
//around FileNames and GetConfig. If createIfNotExist is true, then if the
//files don't exist on disk we'll generate file names and return raw configs
//for those filenames.
func Get(dir string, createIfNotExist bool) (*Config, error) {
	publicConfigName, privateConfigName, err := FileNames(dir, createIfNotExist)

	if err != nil {
		if createIfNotExist {
			publicConfigName, privateConfigName, err = DefaultFileNames(dir)
			if err != nil {
				return nil, errors.New("Couldn't get default file names to create: " + err.Error())
			}
		} else {
			return nil, errors.New("Couldn't get file names to use: " + err.Error())
		}
	}

	return GetConfig(publicConfigName, privateConfigName, createIfNotExist)

}
