package bkit

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

func FlagConfigPath(path ...string) string {
	configPath := "./configs/config.yaml"
	if len(path) > 0 && path[0] != "" {
		configPath = path[0]
	}
	flag.StringVar(&configPath, "config", configPath, fmt.Sprintf("config file default(%s)", configPath))
	if !flag.Parsed() {
		flag.Parse()
	}
	return configPath
}

// ReadConfig read config filename
func ReadConfig(filename string, config interface{}) error {
	if err := NewDefaultValueTag().SetDefaultVal(config); err != nil {
		return fmt.Errorf("set default value error: %w", err)
	}

	byt, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	_, _, typ := filenameSplit(filename)
	switch typ {
	case "toml":
		return toml.Unmarshal(byt, config)
	case "yaml":
		return yaml.Unmarshal(byt, config)
	case "json":
		return json.Unmarshal(byt, config)
	}
	return fmt.Errorf("not support filename %s type", typ)
}

func filenameSplit(filename string) (dir, file, typ string) {
	// decode typ
	dir, file = filepath.Split(filename)
	typ = file[strings.LastIndex(file, ".")+1:]
	return dir, file, typ
}

// MarshalToFile struct to file
func MarshalToFile(config interface{}, filename string) error {
	kind := reflect.TypeOf(config).Kind().String()
	if kind != "struct" && kind != "ptr" {
		return fmt.Errorf("not support config struct")
	}
	var byt []byte
	var err error
	dir, _, typ := filenameSplit(filename)
	if err := os.MkdirAll(dir, os.ModeDir); err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	switch typ {
	case "toml":
		err = toml.NewEncoder(f).Encode(config)
		if err != nil {
			return err
		}
		return nil
	case "yaml":
		byt, err = yaml.Marshal(config)
		if err != nil {
			return err
		}
	case "json":
		byt, err = json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("not support filename %s type", typ)
	}

	_, err = f.Write(byt)

	return err
}
