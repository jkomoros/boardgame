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

func fileNamesToUse(dir string) (publicConfig, privateConfig string, err error) {

	if dir == "" {
		dir = "."
	}

	//Try to interpret it as a file
	if public, private, err := fileNamesToUseWithFile(dir); err == nil {
		return public, private, nil
	}

	//Guess it wasn't a file, try interpreting as a directory.

	goPath, err := filepath.Abs(os.Getenv("GOPATH"))

	if err != nil {
		//Gopath isn't set correctly
		return "", "", errors.New("Couldn't absolutize gopath: " + err.Error())
	}

	for {

		abs, err := filepath.Abs(dir)

		if err != nil {
			//Maybe fell off the end of what is a real file?
			return "", "", errors.New("Got err absolutizing search directory: " + dir + " : " + err.Error())
		}

		if !strings.HasPrefix(abs, goPath) {
			return "", "", errors.New("Fell out of gopath without finding config: " + abs)
		}

		public, private := fileNamesToUseInDir(dir)

		if public != "" || private != "" {
			return public, private, nil
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

//Get fetches a fully realized config. If dir is a config file itself, loads
//that (and any private component in same directory). Next it interprets dir
//as a directory to search within for any config files. If none are found,
//walks upwards in the directory hierarchy (as long as that's still in
//$GOPATH) until it finds a folder that appears to work. If dir is "", working
//directory is assumed.
func Get(dir string) (*Config, error) {
	publicConfigName, privateConfigName, err := fileNamesToUse(dir)

	if err != nil {
		return nil, errors.New("Couldn't get file names to use: " + err.Error())
	}

	publicConfig, err := NewRawConfig(publicConfigName)

	if err != nil {
		return nil, errors.New("Couldn't get public config: " + err.Error())
	}

	privateConfig, err := NewRawConfig(privateConfigName)

	if err != nil {
		return nil, errors.New("Couldn't get private config: " + err.Error())
	}

	return NewConfig(publicConfig, privateConfig)

}
