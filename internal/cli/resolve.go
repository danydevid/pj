package cli

import (
	"errors"
	"fmt"
	"os"
)

func Resolve(envName string) (string, error) {
	home := os.Getenv(envName)
	if home == "" {
		msg := fmt.Sprintf("cannot find %s", envName)
		fmt.Print(msg)
		return "", errors.New(msg)
	}

	return home, nil
}
