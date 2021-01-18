package cli

import (
	"errors"
	"fnd/config"
	"github.com/spf13/cobra"
)

func GetHomeDir(cmd *cobra.Command) string {
	homeDirUnexp, err := cmd.Flags().GetString(FlagHome)
	if err != nil {
		panic(err)
	}
	homeDir := config.ExpandHomePath(homeDirUnexp)
	return homeDir
}

func InitHomeDir(cmd *cobra.Command) (string, error) {
	homeDir := GetHomeDir(cmd)
	exists, err := config.HomeDirExists(homeDir)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("home directory is already initialized")
	}
	if err := config.InitHomeDir(homeDir); err != nil {
		return "", err
	}
	return homeDir, nil
}
