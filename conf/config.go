package conf

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/bbdshow/bkit/gen/defval"
	ptoml "github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

func FlagConfigPath(path ...string) string {
	configPath := "./configs/config.toml"
	if len(path) > 0 && path[0] != "" {
		configPath = path[0]
	}
	flag.StringVar(&configPath, "f", configPath, fmt.Sprintf("config file default(%s)", configPath))
	if !flag.Parsed() {
		flag.Parse()
	}
	return configPath
}

// PrintJSON  print JSON string,  null tag, will hide
func PrintJSON(config interface{}) (string, error) {
	if err := defval.InitialNullVal(config); err != nil {
		return "", err
	}
	byts, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(byts), nil
}

// ReadConfig read config filename
func ReadConfig(filename string, config interface{}) error {
	if err := UnmarshalDefaultVal(config); err != nil {
		return err
	}
	byts, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	_, _, typ := filenameSplit(filename)
	switch typ {
	case "toml":
		return toml.Unmarshal(byts, config)
	case "yaml":
		return yaml.Unmarshal(byts, config)
	case "json":
		return json.Unmarshal(byts, config)
	}
	return fmt.Errorf("not support filename %s type", typ)
}

func filenameSplit(filename string) (dir, file, typ string) {
	// decode typ
	dir, file = filepath.Split(filename)
	typ = file[strings.LastIndex(file, ".")+1:]
	return dir, file, typ
}

// UnmarshalDefaultVal struct, 'defval' tag to default value
func UnmarshalDefaultVal(config interface{}) error {
	return defval.ParseDefaultVal(config)
}

// MarshalToFile struct to file
func MarshalToFile(config interface{}, filename string) error {
	kind := reflect.TypeOf(config).Kind().String()
	if kind != "struct" && kind != "ptr" {
		return fmt.Errorf("not support config struct")
	}
	var byts []byte
	var err error
	dir, _, typ := filenameSplit(filename)
	if err := os.MkdirAll(dir, os.ModeDir); err != nil {
		return err
	}
	switch typ {
	case "toml":
		byts, err = ptoml.Marshal(config)
		if err != nil {
			return err
		}
	case "yaml":
		byts, err = yaml.Marshal(config)
		if err != nil {
			return err
		}
	case "json":
		byts, err = json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("not support filename %s type", typ)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(byts)
	return err
}
