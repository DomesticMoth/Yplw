package confer


import (
	"os"
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"
)

// Load configuration from toml file located on the path passed in the arguments or on one of the default paths.
func LoadConfig(paths []string, conf interface{}) (err error){
	// If the path is passed through arguments
	if len(os.Args) > 1 {
		// Put it at the top of the list of paths
		paths = append(os.Args[1:2], paths...)
	}
	// An error for the case if there is no path
	err = errors.New("No path was passed to find the configuration file")
	// Trying each of the paths
	for _, path := range paths {
		path, err = homedir.Expand(path)
		if err != nil { return }	
		_, err = toml.DecodeFile(path, conf)
		// If config was loaded, return the result
		if err == nil { return }	
	}
	return
}
