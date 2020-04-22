package jwt

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/nats-io/nkeys"
)

// DecorateJWT returns a decorated JWT that describes the kind of JWT
func DecorateJWT(jwtString string) ([]byte, error) {
	gc, err := DecodeGeneric(jwtString)
	if err != nil {
		return nil, err
	}
	return formatJwt(string(gc.Type), jwtString)
}

func formatJwt(kind string, jwtString string) ([]byte, error) {
	templ := `-----BEGIN NATS %s JWT-----
%s
------END NATS %s JWT------

`
	w := bytes.NewBuffer(nil)
	kind = strings.ToUpper(kind)
	_, err := fmt.Fprintf(w, templ, kind, jwtString, kind)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// DecorateSeed takes a seed and returns a string that wraps
// the seed in the form:
//  ************************* IMPORTANT *************************
//  NKEY Seed printed below can be used sign and prove identity.
//  NKEYs are sensitive and should be treated as secrets.
//
//  -----BEGIN USER NKEY SEED-----
//  SUAIO3FHUX5PNV2LQIIP7TZ3N4L7TX3W53MQGEIVYFIGA635OZCKEYHFLM
//  ------END USER NKEY SEED------
func DecorateSeed(seed []byte) ([]byte, error) {
	w := bytes.NewBuffer(nil)
	ts := bytes.TrimSpace(seed)
	pre := string(ts[0:2])
	kind := ""
	switch pre {
	case "SU":
		kind = "USER"
	case "SA":
		kind = "ACCOUNT"
	case "SO":
		kind = "OPERATOR"
	default:
		return nil, errors.New("seed is not an operator, account or user seed")
	}
	header := `************************* IMPORTANT *************************
NKEY Seed printed below can be used to sign and prove identity.
NKEYs are sensitive and should be treated as secrets.

-----BEGIN %s NKEY SEED-----
`
	_, err := fmt.Fprintf(w, header, kind)
	if err != nil {
		return nil, err
	}
	w.Write(ts)

	footer := `
------END %s NKEY SEED------

*************************************************************
`
	_, err = fmt.Fprintf(w, footer, kind)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

var userConfigRE = regexp.MustCompile(`\s*(?:(?:[-]{3,}[^\n]*[-]{3,}\n)(.+)(?:\n\s*[-]{3,}[^\n]*[-]{3,}\n))`)

// An user config file looks like this:
//  -----BEGIN NATS USER JWT-----
//  eyJ0eXAiOiJqd3QiLCJhbGciOiJlZDI1NTE5...
//  ------END NATS USER JWT------
//
//  ************************* IMPORTANT *************************
//  NKEY Seed printed below can be used sign and prove identity.
//  NKEYs are sensitive and should be treated as secrets.
//
//  -----BEGIN USER NKEY SEED-----
//  SUAIO3FHUX5PNV2LQIIP7TZ3N4L7TX3W53MQGEIVYFIGA635OZCKEYHFLM
//  ------END USER NKEY SEED------

// FormatUserConfig returns a decorated file with a decorated JWT and decorated seed
func FormatUserConfig(jwtString string, seed []byte) ([]byte, error) {
	gc, err := DecodeGeneric(jwtString)
	if err != nil {
		return nil, err
	}
	if gc.Type != UserClaim {
		return nil, fmt.Errorf("%q cannot be serialized as a user config", string(gc.Type))
	}

	w := bytes.NewBuffer(nil)

	jd, err := formatJwt(string(gc.Type), jwtString)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(jd)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(bytes.TrimSpace(seed), []byte("SU")) {
		return nil, fmt.Errorf("nkey seed is not an user seed")
	}

	d, err := DecorateSeed(seed)
	if err != nil {
		return nil, err
	}
	_, err = w.Write(d)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// ParseDecoratedJWT takes a creds file and returns the JWT portion.
func ParseDecoratedJWT(contents []byte) (string, error) {
	items := userConfigRE.FindAllSubmatch(contents, -1)
	if len(items) == 0 {
		return string(contents), nil
	}
	// First result should be the user JWT.
	// We copy here so that if the file contained a seed file too we wipe appropriately.
	raw := items[0][1]
	tmp := make([]byte, len(raw))
	copy(tmp, raw)
	return string(tmp), nil
}

// ParseDecoratedNKey takes a creds file, finds the NKey portion and creates a
// key pair from it.
func ParseDecoratedNKey(contents []byte) (nkeys.KeyPair, error) {
	var seed []byte

	items := userConfigRE.FindAllSubmatch(contents, -1)
	if len(items) > 1 {
		seed = items[1][1]
	} else {
		lines := bytes.Split(contents, []byte("\n"))
		for _, line := range lines {
			if bytes.HasPrefix(bytes.TrimSpace(line), []byte("SO")) ||
				bytes.HasPrefix(bytes.TrimSpace(line), []byte("SA")) ||
				bytes.HasPrefix(bytes.TrimSpace(line), []byte("SU")) {
				seed = line
				break
			}
		}
	}
	if seed == nil {
		return nil, errors.New("no nkey seed found")
	}
	if !bytes.HasPrefix(seed, []byte("SO")) &&
		!bytes.HasPrefix(seed, []byte("SA")) &&
		!bytes.HasPrefix(seed, []byte("SU")) {
		return nil, errors.New("doesn't contain a seed nkey")
	}
	kp, err := nkeys.FromSeed(seed)
	if err != nil {
		return nil, err
	}
	return kp, nil
}

// ParseDecoratedUserNKey takes a creds file, finds the NKey portion and creates a
// key pair from it. Similar to ParseDecoratedNKey but fails for non-user keys.
func ParseDecoratedUserNKey(contents []byte) (nkeys.KeyPair, error) {
	nk, err := ParseDecoratedNKey(contents)
	if err != nil {
		return nil, err
	}
	seed, err := nk.Seed()
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(seed, []byte("SU")) {
		return nil, errors.New("doesn't contain an user seed nkey")
	}
	kp, err := nkeys.FromSeed(seed)
	if err != nil {
		return nil, err
	}
	return kp, nil
}
