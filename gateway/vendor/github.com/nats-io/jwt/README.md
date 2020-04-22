# JWT
A [JWT](https://jwt.io/) implementation that uses [nkeys](https://github.com/nats-io/nkeys) to digitally sign JWT tokens. 
Nkeys use [Ed25519](https://ed25519.cr.yp.to/) to provide authentication of JWT claims.


[![License Apache 2](https://img.shields.io/badge/License-Apache2-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)
[![ReportCard](http://goreportcard.com/badge/nats-io/jwt)](http://goreportcard.com/report/nats-io/jwt)
[![Build Status](https://travis-ci.org/nats-io/jwt.svg?branch=master)](http://travis-ci.org/nats-io/jwt)
[![GoDoc](http://godoc.org/github.com/nats-io/jwt?status.png)](http://godoc.org/github.com/nats-io/jwt)
[![Coverage Status](https://coveralls.io/repos/github/nats-io/jwt/badge.svg?branch=master&t=NmEFup)](https://coveralls.io/github/nats-io/jwt?branch=master)

```go
// Need a private key to sign the claim, nkeys makes it easy to create
kp, err := nkeys.CreateAccount()
if err != nil {
    t.Fatal("unable to create account key", err)
}

pk, err := kp.PublicKey()
if err != nil {
	t.Fatal("error getting public key", err)
}

// create a new claim
claims := NewAccountClaims(pk)
claims.Expires = time.Now().Add(time.Duration(time.Hour)).Unix()


// add details by modifying claims.Account

// serialize the claim to a JWT token
token, err := claims.Encode(kp)
if err != nil {
    t.Fatal("error encoding token", err)
}

// on the receiving side, decode the token
c, err := DecodeAccountClaims(token)
if err != nil {
    t.Fatal(err)
}

// if the token was decoded, it means that it
// validated and it wasn't tampered. the remaining and
// required test is to insure the issuer is trusted
pk, err := kp.PublicKey()
if err != nil {
    t.Fatalf("unable to read public key: %v", err)
}

if c.Issuer != pk {
    t.Fatalf("the public key is not trusted")
}
```