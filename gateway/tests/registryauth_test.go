// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package tests

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alexellis/faas/gateway/handlers"
	"github.com/docker/docker/api/types"
)

func TestBuildEncodedAuthConfig(t *testing.T) {
	// custom repository with valid data
	assertValidEncodedAuthConfig(t, "user", "password", "my.repository.com/user/imagename", "my.repository.com")
	assertValidEncodedAuthConfig(t, "user", "weird:password:", "my.repository.com/user/imagename", "my.repository.com")
	assertValidEncodedAuthConfig(t, "userWithNoPassword", "", "my.repository.com/user/imagename", "my.repository.com")
	assertValidEncodedAuthConfig(t, "", "", "my.repository.com/user/imagename", "my.repository.com")

	// docker hub default repository
	assertValidEncodedAuthConfig(t, "user", "password", "user/imagename", "docker.io")
	assertValidEncodedAuthConfig(t, "", "", "user/imagename", "docker.io")

	// invalid base64 basic auth
	assertEncodedAuthError(t, "invalidBasicAuth", "my.repository.com/user/imagename")

	// invalid docker image name
	assertEncodedAuthError(t, b64BasicAuth("user", "password"), "")
	assertEncodedAuthError(t, b64BasicAuth("user", "password"), "invalid name")
}

func assertValidEncodedAuthConfig(t *testing.T, user, password, imageName, expectedRegistryHost string) {
	encodedAuthConfig, err := handlers.BuildEncodedAuthConfig(b64BasicAuth(user, password), imageName)
	if err != nil {
		t.Log("Unexpected error while building auth config with correct values")
		t.Fail()
	}

	authConfig := &types.AuthConfig{}
	authJSON := base64.NewDecoder(base64.URLEncoding, strings.NewReader(encodedAuthConfig))
	if err := json.NewDecoder(authJSON).Decode(authConfig); err != nil {
		t.Log("Invalid encoded auth", err)
		t.Fail()
	}

	if user != authConfig.Username {
		t.Log("Auth config username mismatch", user, authConfig.Username)
		t.Fail()
	}
	if password != authConfig.Password {
		t.Log("Auth config password mismatch", password, authConfig.Password)
		t.Fail()
	}
	if expectedRegistryHost != authConfig.ServerAddress {
		t.Log("Auth config registry server address mismatch", expectedRegistryHost, authConfig.ServerAddress)
		t.Fail()
	}
}

func assertEncodedAuthError(t *testing.T, b64BasicAuth, imageName string) {
	_, err := handlers.BuildEncodedAuthConfig(b64BasicAuth, imageName)
	if err == nil {
		t.Log("Expected an error to be returned")
		t.Fail()
	}
}

func b64BasicAuth(user, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + password))
}
