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
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

// ExportType defines the type of import/export.
type ExportType int

const (
	// Unknown is used if we don't know the type
	Unknown ExportType = iota
	// Stream defines the type field value for a stream "stream"
	Stream
	// Service defines the type field value for a service "service"
	Service
)

func (t ExportType) String() string {
	switch t {
	case Stream:
		return "stream"
	case Service:
		return "service"
	}
	return "unknown"
}

// MarshalJSON marshals the enum as a quoted json string
func (t *ExportType) MarshalJSON() ([]byte, error) {
	switch *t {
	case Stream:
		return []byte("\"stream\""), nil
	case Service:
		return []byte("\"service\""), nil
	}
	return nil, fmt.Errorf("unknown export type")
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (t *ExportType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	switch j {
	case "stream":
		*t = Stream
		return nil
	case "service":
		*t = Service
		return nil
	}
	return fmt.Errorf("unknown export type")
}

// Subject is a string that represents a NATS subject
type Subject string

// Validate checks that a subject string is valid, ie not empty and without spaces
func (s Subject) Validate(vr *ValidationResults) {
	v := string(s)
	if v == "" {
		vr.AddError("subject cannot be empty")
	}
	if strings.Contains(v, " ") {
		vr.AddError("subject %q cannot have spaces", v)
	}
}

// HasWildCards is used to check if a subject contains a > or *
func (s Subject) HasWildCards() bool {
	v := string(s)
	return strings.HasSuffix(v, ".>") ||
		strings.Contains(v, ".*.") ||
		strings.HasSuffix(v, ".*") ||
		strings.HasPrefix(v, "*.") ||
		v == "*" ||
		v == ">"
}

// IsContainedIn does a simple test to see if the subject is contained in another subject
func (s Subject) IsContainedIn(other Subject) bool {
	otherArray := strings.Split(string(other), ".")
	myArray := strings.Split(string(s), ".")

	if len(myArray) > len(otherArray) && otherArray[len(otherArray)-1] != ">" {
		return false
	}

	if len(myArray) < len(otherArray) {
		return false
	}

	for ind, tok := range otherArray {
		myTok := myArray[ind]

		if ind == len(otherArray)-1 && tok == ">" {
			return true
		}

		if tok != myTok && tok != "*" {
			return false
		}
	}

	return true
}

// NamedSubject is the combination of a subject and a name for it
type NamedSubject struct {
	Name    string  `json:"name,omitempty"`
	Subject Subject `json:"subject,omitempty"`
}

// Validate checks the subject
func (ns *NamedSubject) Validate(vr *ValidationResults) {
	ns.Subject.Validate(vr)
}

// TimeRange is used to represent a start and end time
type TimeRange struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

// Validate checks the values in a time range struct
func (tr *TimeRange) Validate(vr *ValidationResults) {
	format := "15:04:05"

	if tr.Start == "" {
		vr.AddError("time ranges start must contain a start")
	} else {
		_, err := time.Parse(format, tr.Start)
		if err != nil {
			vr.AddError("start in time range is invalid %q", tr.Start)
		}
	}

	if tr.End == "" {
		vr.AddError("time ranges end must contain an end")
	} else {
		_, err := time.Parse(format, tr.End)
		if err != nil {
			vr.AddError("end in time range is invalid %q", tr.End)
		}
	}
}

// Limits are used to control acccess for users and importing accounts
// Src is a comma separated list of CIDR specifications
type Limits struct {
	Max     int64       `json:"max,omitempty"`
	Payload int64       `json:"payload,omitempty"`
	Src     string      `json:"src,omitempty"`
	Times   []TimeRange `json:"times,omitempty"`
}

// Validate checks the values in a limit struct
func (l *Limits) Validate(vr *ValidationResults) {
	if l.Max < 0 {
		vr.AddError("limits cannot contain a negative maximum, %d", l.Max)
	}
	if l.Payload < 0 {
		vr.AddError("limits cannot contain a negative payload, %d", l.Payload)
	}

	if l.Src != "" {
		elements := strings.Split(l.Src, ",")

		for _, cidr := range elements {
			cidr = strings.TrimSpace(cidr)
			_, ipNet, err := net.ParseCIDR(cidr)
			if err != nil || ipNet == nil {
				vr.AddError("invalid cidr %q in user src limits", cidr)
			}
		}
	}

	if l.Times != nil && len(l.Times) > 0 {
		for _, t := range l.Times {
			t.Validate(vr)
		}
	}
}

// Permission defines allow/deny subjects
type Permission struct {
	Allow StringList `json:"allow,omitempty"`
	Deny  StringList `json:"deny,omitempty"`
}

// Validate the allow, deny elements of a permission
func (p *Permission) Validate(vr *ValidationResults) {
	for _, subj := range p.Allow {
		Subject(subj).Validate(vr)
	}
	for _, subj := range p.Deny {
		Subject(subj).Validate(vr)
	}
}

// ResponsePermission can be used to allow responses to any reply subject
// that is received on a valid subscription.
type ResponsePermission struct {
	MaxMsgs int           `json:"max"`
	Expires time.Duration `json:"ttl"`
}

// Validate the response permission.
func (p *ResponsePermission) Validate(vr *ValidationResults) {
	// Any values can be valid for now.
}

// Permissions are used to restrict subject access, either on a user or for everyone on a server by default
type Permissions struct {
	Pub  Permission          `json:"pub,omitempty"`
	Sub  Permission          `json:"sub,omitempty"`
	Resp *ResponsePermission `json:"resp,omitempty"`
}

// Validate the pub and sub fields in the permissions list
func (p *Permissions) Validate(vr *ValidationResults) {
	p.Pub.Validate(vr)
	p.Sub.Validate(vr)
	if p.Resp != nil {
		p.Resp.Validate(vr)
	}
}

// StringList is a wrapper for an array of strings
type StringList []string

// Contains returns true if the list contains the string
func (u *StringList) Contains(p string) bool {
	for _, t := range *u {
		if t == p {
			return true
		}
	}
	return false
}

// Add appends 1 or more strings to a list
func (u *StringList) Add(p ...string) {
	for _, v := range p {
		if !u.Contains(v) && v != "" {
			*u = append(*u, v)
		}
	}
}

// Remove removes 1 or more strings from a list
func (u *StringList) Remove(p ...string) {
	for _, v := range p {
		for i, t := range *u {
			if t == v {
				a := *u
				*u = append(a[:i], a[i+1:]...)
				break
			}
		}
	}
}

// TagList is a unique array of lower case strings
// All tag list methods lower case the strings in the arguments
type TagList []string

// Contains returns true if the list contains the tags
func (u *TagList) Contains(p string) bool {
	p = strings.ToLower(p)
	for _, t := range *u {
		if t == p {
			return true
		}
	}
	return false
}

// Add appends 1 or more tags to a list
func (u *TagList) Add(p ...string) {
	for _, v := range p {
		v = strings.ToLower(v)
		if !u.Contains(v) && v != "" {
			*u = append(*u, v)
		}
	}
}

// Remove removes 1 or more tags from a list
func (u *TagList) Remove(p ...string) {
	for _, v := range p {
		v = strings.ToLower(v)
		for i, t := range *u {
			if t == v {
				a := *u
				*u = append(a[:i], a[i+1:]...)
				break
			}
		}
	}
}

// Identity is used to associate an account or operator with a real entity
type Identity struct {
	ID    string `json:"id,omitempty"`
	Proof string `json:"proof,omitempty"`
}

// Validate checks the values in an Identity
func (u *Identity) Validate(vr *ValidationResults) {
	//Fixme identity validation
}
