package config

import "github.com/zalando/go-keyring"

const keyringService = "sqlpilot"

func SavePassword(profileName, password string) error {
	if password == "" {
		return nil
	}
	return keyring.Set(keyringService, profileName, password)
}

func LoadPassword(profileName string) (string, error) {
	return keyring.Get(keyringService, profileName)
}
