// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package auth

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func Test_ReadFromCustomLocation_AndNames(t *testing.T) {
	tmp := os.TempDir()

	userWant := "admin"
	ioutil.WriteFile(path.Join(tmp, "user.txt"), []byte(userWant), 0700)

	passWant := "test1234"
	ioutil.WriteFile(path.Join(tmp, "pass.txt"), []byte(passWant), 0700)

	reader := ReadBasicAuthFromDisk{
		SecretMountPath:  tmp,
		UserFilename:     "user.txt",
		PasswordFilename: "pass.txt",
	}

	creds, err := reader.Read()
	if err != nil {
		t.Errorf("can't read secrets: %s", err.Error())
	}

	if creds.User != userWant {
		t.Errorf("user, want: %s, got %s", userWant, creds.User)
	}
	if creds.Password != passWant {
		t.Errorf("password, want: %s, got %s", passWant, creds.Password)
	}
}

func Test_ReadFromCustomLocation_DefaultNames(t *testing.T) {
	tmp := os.TempDir()
	userWant := "admin"
	ioutil.WriteFile(path.Join(tmp, "basic-auth-user"), []byte(userWant), 0700)

	passWant := "test1234"
	ioutil.WriteFile(path.Join(tmp, "basic-auth-password"), []byte(passWant), 0700)

	reader := ReadBasicAuthFromDisk{
		SecretMountPath: tmp,
	}

	creds, err := reader.Read()
	if err != nil {
		t.Errorf("can't read secrets: %s", err.Error())
	}

	if creds.User != userWant {
		t.Errorf("user, want: %s, got %s", userWant, creds.User)
	}
	if creds.Password != passWant {
		t.Errorf("password, want: %s, got %s", passWant, creds.Password)
	}
}
