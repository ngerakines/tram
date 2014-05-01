package util

import (
	"os"
	"path/filepath"
	"runtime"
)

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"))
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func CWD() string {
	pwd, err := os.Getwd()
	if err != nil {
		// NKG: I'm sure this is going to fuck up someone's shit somewhere.
		return ""
	}
	return pwd
}
