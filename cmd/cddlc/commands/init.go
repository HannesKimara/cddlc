package commands

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/HannesKimara/cddlc/config"
	"github.com/urfave/cli/v2"
)

const configFile string = "cddlc.json"

func getFileHandle(name string) (io.Writer, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	fp := filepath.Join(root, configFile)

	_, err = os.Stat(fp)
	switch {
	case err == nil:
		// file exists
		return nil, errors.New("existing configuration for project found in " + fp)
	case os.IsNotExist(err):
		return os.OpenFile(fp, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	default:
		return nil, err
	}
}

func InitCmd(cCtx *cli.Context) error {
	conf := config.NewDefaultConfig()
	f, err := getFileHandle(configFile)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "	")

	if err := enc.Encode(conf); err != nil {
		return err
	}
	return nil
}
