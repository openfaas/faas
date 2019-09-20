// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import "testing"

func Test_getNamespace_Default(t *testing.T) {
	root, ns := getNamespace("openfaas-fn", "figlet.openfaas-fn")
	wantRoot := "figlet"
	wantNs := "openfaas-fn"

	if root != wantRoot {
		t.Errorf("function root: want %s, got %s", wantRoot, root)
	}
	if ns != wantNs {
		t.Errorf("function ns: want %s, got %s", wantNs, ns)
	}
}

func Test_getNamespace_Override(t *testing.T) {
	root, ns := getNamespace("fn", "figlet.fn")
	wantRoot := "figlet"
	wantNs := "fn"

	if root != wantRoot {
		t.Errorf("function root: want %s, got %s", wantRoot, root)
	}
	if ns != wantNs {
		t.Errorf("function ns: want %s, got %s", wantNs, ns)
	}
}

func Test_getNamespace_Empty(t *testing.T) {
	root, ns := getNamespace("", "figlet")
	wantRoot := "figlet"
	wantNs := ""

	if root != wantRoot {
		t.Errorf("function root: want %s, got %s", wantRoot, root)
	}
	if ns != wantNs {
		t.Errorf("function ns: want %s, got %s", wantNs, ns)
	}
}
