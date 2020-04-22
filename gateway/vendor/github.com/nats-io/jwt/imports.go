/*
 * Copyright 2018-2019 The NATS Authors
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package jwt

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Import describes a mapping from another account into this one
type Import struct {
	Name string `json:"name,omitempty"`
	// Subject field in an import is always from the perspective of the
	// initial publisher - in the case of a stream it is the account owning
	// the stream (the exporter), and in the case of a service it is the
	// account making the request (the importer).
	Subject Subject `json:"subject,omitempty"`
	Account string  `json:"account,omitempty"`
	Token   string  `json:"token,omitempty"`
	// To field in an import is always from the perspective of the subscriber
	// in the case of a stream it is the client of the stream (the importer),
	// from the perspective of a service, it is the subscription waiting for
	// requests (the exporter). If the field is empty, it will default to the
	// value in the Subject field.
	To   Subject    `json:"to,omitempty"`
	Type ExportType `json:"type,omitempty"`
}

// IsService returns true if the import is of type service
func (i *Import) IsService() bool {
	return i.Type == Service
}

// IsStream returns true if the import is of type stream
func (i *Import) IsStream() bool {
	return i.Type == Stream
}

// Validate checks if an import is valid for the wrapping account
func (i *Import) Validate(actPubKey string, vr *ValidationResults) {
	if !i.IsService() && !i.IsStream() {
		vr.AddError("invalid import type: %q", i.Type)
	}

	if i.Account == "" {
		vr.AddWarning("account to import from is not specified")
	}

	i.Subject.Validate(vr)

	if i.IsService() && i.Subject.HasWildCards() {
		vr.AddError("services cannot have wildcard subject: %q", i.Subject)
	}
	if i.IsStream() && i.To.HasWildCards() {
		vr.AddError("streams cannot have wildcard to subject: %q", i.Subject)
	}

	var act *ActivationClaims

	if i.Token != "" {
		// Check to see if its an embedded JWT or a URL.
		if url, err := url.Parse(i.Token); err == nil && url.Scheme != "" {
			c := &http.Client{Timeout: 5 * time.Second}
			resp, err := c.Get(url.String())
			if err != nil {
				vr.AddWarning("import %s contains an unreachable token URL %q", i.Subject, i.Token)
			}

			if resp != nil {
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					vr.AddWarning("import %s contains an unreadable token URL %q", i.Subject, i.Token)
				} else {
					act, err = DecodeActivationClaims(string(body))
					if err != nil {
						vr.AddWarning("import %s contains a url %q with an invalid activation token", i.Subject, i.Token)
					}
				}
			}
		} else {
			var err error
			act, err = DecodeActivationClaims(i.Token)
			if err != nil {
				vr.AddWarning("import %q contains an invalid activation token", i.Subject)
			}
		}
	}

	if act != nil {
		if act.Issuer != i.Account {
			vr.AddWarning("activation token doesn't match account for import %q", i.Subject)
		}

		if act.ClaimsData.Subject != actPubKey {
			vr.AddWarning("activation token doesn't match account it is being included in, %q", i.Subject)
		}
	} else {
		vr.AddWarning("no activation provided for import %s", i.Subject)
	}

}

// Imports is a list of import structs
type Imports []*Import

// Validate checks if an import is valid for the wrapping account
func (i *Imports) Validate(acctPubKey string, vr *ValidationResults) {
	toSet := make(map[Subject]bool, len(*i))
	for _, v := range *i {
		if v.Type == Service {
			if _, ok := toSet[v.To]; ok {
				vr.AddError("Duplicate To subjects for %q", v.To)
			}
			toSet[v.To] = true
		}
		v.Validate(acctPubKey, vr)
	}
}

// Add is a simple way to add imports
func (i *Imports) Add(a ...*Import) {
	*i = append(*i, a...)
}

func (i Imports) Len() int {
	return len(i)
}

func (i Imports) Swap(j, k int) {
	i[j], i[k] = i[k], i[j]
}

func (i Imports) Less(j, k int) bool {
	return i[j].Subject < i[k].Subject
}
