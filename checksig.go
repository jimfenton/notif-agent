/*

checksig.go - Signature checking for prototype notification agent

Copyright (c) 2015 Jim Fenton

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to
deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/jimfenton/notif-agent/notif"
	"io"
	"net"
	"net/http"
	"strings"
)

//Check the signature
func checkSig(
	npr notifProtected,
	w http.ResponseWriter,
	auth notif.Auth,
	flatload []string) bool { //TODO: args a bit redundant (npr, flatload). Rationalize.

	var publickey *rsa.PublicKey
	var h crypto.Hash
	var err error
	var pubkey []byte

	if npr.Algorithm != "RS256" {

		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Unsupported signature algorithm")
		return true
	}

	//Retrieve the public key for the signature. This is a DKIM key found in DNS at
	// <kid>._domainkey.<domain>, as the value of the p= tag.

	pubkey64 := getkey(npr.Selector, auth.Domain)

	//Parse the public key
	pubkey, err = base64.StdEncoding.DecodeString(pubkey64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Public key decode error")
		return true
	}

	pub, err := x509.ParsePKIXPublicKey(pubkey)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Public key parse error")
		return true
	}

	publickey, ok := pub.(*rsa.PublicKey)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Value not a public key")
		return true
	}

	sig, err := base64.URLEncoding.DecodeString(pad64(flatload[2]))
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Signature decode error")
		return true
	}

	h = crypto.SHA256
	hash := sha256.New()
	io.WriteString(hash, flatload[0]+"."+flatload[1])
	hashstr := hash.Sum(nil)

	err = rsa.VerifyPKCS1v15(publickey, h, hashstr, sig)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Signature verification error")
		return true
	}

	return false
}

//Find the next tag/value in a DKIM key record
// return values:
// tag (string) - name of the tag found
// value (string) - value associated with tag
// next (int) - index to next tag/value if any, or 0 if last one

func nexttag(keyrec string, start int) (string, string, int) {
	var tag string
	var value string

	eq := strings.Index(keyrec[start:], "=")
	if eq == -1 {
		return "", "", -1
	}
	tag = strings.TrimSpace(keyrec[start : eq+start])
	sc := strings.Index(keyrec[start:], ";")

	if sc == -1 { //Last field
		value = strings.TrimSpace(keyrec[eq+start+1:])
		return tag, value, -1
	}

	if eq > sc { //no equals sign between semicolons so empty answer
		return "", "", sc + start + 1
	}
	value = strings.TrimSpace(keyrec[eq+start+1 : sc+start])

	return tag, value, sc + start + 1
}

// Retrieve and verify a DKIM public key from DNS.
func getkey(selector string, domain string) string {

	var pubkey string
	var tag string
	var value string

	selectors, err := net.LookupTXT(selector + "._domainkey." + domain)
	if err != nil {
		fmt.Println("Signature key not found: ", err)
		return ""
	}

	//Just looking at the first TXT record. Multiple TXT records are not a good idea.
	for i := 0; i >= 0; {
		tag, value, i = nexttag(selectors[0], i)

		switch tag {
		case "v": //Check version if present
			if value != "DKIM1" {
				return ""
			}
		case "p": //extract key
			pubkey = value
		case "k": //Check key type
			if value != "rsa" {
				return ""
			}
		case "h": //Check allowed hashes
			// Should handle a list here
			if value != "sha256" {
				return ""
			}
		case "s": //service type
			// Should potentially handle a service list here
			if value != "*" && value != "notif" {
				return ""
			}
		}
	}
	return pubkey
}
