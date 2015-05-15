/*

native.go - Native notif collector for prototype notification agent

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

/* Design philosophy:

The code in this file is a goroutine that deals with the collection of "native" notifs -- those that are sent via the Notifs API (as opposed to those collected from other services, such as RSS and Twitter). Since the Authorizations collection (database table) is associated only with native notifs, it is handled exclusively by this file in the agent.

*/

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/jimfenton/notif-agent/notif"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type agent struct {
	NotifColl    *mgo.Collection
	AuthColl     *mgo.Collection
	UserinfoColl *mgo.Collection
	CollChan     chan notif.Notif
}

type notifMsg struct { //Notification format "on the wire"
	Header  notifHeader `json:"header"`  //Unprotected headers (to, notID)
	Payload string      `json:"payload"` //Payload (protHeaders.payload.sig), each URLsafe base64 encoded
}

type notifHeader struct {
	To    string `json:"to"`
	NotID string `json:"notid"`
}

type notifProtected struct {
	Algorithm string `json:"alg"`
	Selector  string `json:"kid"` //Key ID in JWS terminology
}

type notifPayload struct {
	To       string         `json:"to"`       //UUID specifying recipient authorization
	Origtime time.Time      `json:"origtime"` //Time as sent by originator
	Priority notif.NotifPri `json:"priority"` //Notification priority
	Expires  time.Time      `json:"expires"`
	Subject  string         `json:"subject"`
	Body     string         `json:"body"` //May become MIME-like JSON
}

// Handle a single native Notif API request

func (ag agent) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request) {

	var nm notifMsg
	var np notifPayload
	var npr notifProtected
	var nd notif.Notif
	var auth notif.Auth
	var uinfo notif.Userinfo
	var body []byte
	var payload []byte
	var flatload []string //"flattened" payload (header.payload.sig each base64)
	var protected []byte
	var err error
	var addr string //auth (POST) or id (PUT, DELE) from URL

	if r.Method != "POST" && r.Method != "PUT" && r.Method != "DELE" {
		w.Header().Add("Allow", "GET, PUT, DELE")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err = ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Read error: ", err)
		return
	}

	if r.URL.Path[0:8] == "/notify/" { // Remove leading /notify/ if present
		addr = r.URL.Path[8:]
	} else {
		addr = r.URL.Path[1:]
	}

	err = json.Unmarshal(body, &nm)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Message unmarshal error")
		return
	}

	flatload = strings.SplitN(nm.Payload, ".", 3)
	payload, err = base64.URLEncoding.DecodeString(pad64(flatload[1]))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Payload base64 decode error")
		return
	}

	err = json.Unmarshal(payload, &np)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Payload unmarshal error")
		return
	}

	protected, err = base64.URLEncoding.DecodeString(pad64(flatload[0]))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Protected headers base64 decode error")
		return
	}

	err = json.Unmarshal(protected, &npr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Protected headers unmarshal error")
		return
	}

	//TODO: Still need to check to see if expiration is not in the past, etc.
	//At this point, basic syntax looks good

	switch r.Method {
	case "POST":
		err = ag.AuthColl.Find(bson.M{"address": addr}).One(&auth)
		if err != nil || auth.Deleted {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Authorization not found")
			return
		}

		if !auth.Active {
			w.WriteHeader(http.StatusConflict) // 409 Conflict
			fmt.Fprint(w, "Inactive authorization")
			return
		}

		if checkSig(npr, w, auth, flatload) {
			return
		}

		if auth.Maxpri > np.Priority {
			fmt.Println("Authorized priority ", auth.Maxpri, " exceeded")
			np.Priority = auth.Maxpri
			// Wonder if a different result code should be returned here
		}

		nd.Id = bson.NewObjectId()
		nd.UserID = auth.UserID
		nd.To = addr
		nd.Origtime = np.Origtime
		nd.Expires = np.Expires
		nd.Subject = np.Subject
		nd.From = auth.Domain
		nd.Description = auth.Description
		nd.Priority = np.Priority
		nd.Body = np.Body
		nd.NotID = uuid.New()
		nd.RecvTime = time.Now()

		//Update the notification count and time on the authorization
		auth.Count += 1
		auth.Latest = nd.RecvTime
		err = ag.AuthColl.UpdateId(auth.Id, auth)
		if err != nil {
			fmt.Println("Error updating authorization: ", err)
			return
		}

		//Writing the notif itself should probably be common code with other collectors

		err = ag.NotifColl.Insert(nd)
		if err != nil {
			fmt.Println("Insert error: ", err)
			return
		}

		//Update the user's notification count and latest notification time
		err = ag.UserinfoColl.Find(bson.M{"user_id": auth.UserID}).One(&uinfo)
		if err != nil {
			fmt.Println("POST: Error finding user info document: ", err, auth.UserID)
			return
		}
		uinfo.Count += 1
		uinfo.Latest = nd.RecvTime
		err = ag.UserinfoColl.UpdateId(uinfo.Id, uinfo)
		if err != nil {
			fmt.Println("POST: Error updating user info: ", err)
			return
		}

		//Tell the notifier the notification ID in the response
		resp := "{ \"notid\": \"" + nd.NotID + "\" }"
		fmt.Fprint(w, resp)

		ag.CollChan <- nd

		//Read the rules and execute any required push actions
		//		ProcessRules(ag, nd, auth, uinfo)

	case "PUT":
		err = ag.NotifColl.Find(bson.M{"notID": addr}).One(&nd)
		if err != nil {
			fmt.Println("NotID not found: ", err, " ", addr)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "PUT: Notification ID not found")
			return
		}

		err = ag.AuthColl.Find(bson.M{"address": nd.To}).One(&auth)

		if err != nil || auth.Deleted {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "PUT: Authorization not found")
			return
		}

		if !auth.Active {
			w.WriteHeader(http.StatusConflict) //409 Conflict
			fmt.Fprint(w, "Inactive authorization")
			return
		}

		if checkSig(npr, w, auth, flatload) {
			return
		}

		if nd.Origtime.After(np.Origtime) { //time has gone backwards!
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, "PUT: Update to later notif")
			return
		}

		nd.Origtime = np.Origtime
		nd.Expires = np.Expires
		nd.Subject = np.Subject
		nd.Priority = np.Priority
		nd.Body = np.Body
		nd.RecvTime = time.Now()
		nd.RevCount = nd.RevCount + 1
		nd.Read = false
		nd.UserID = auth.UserID //should already be there, but just in case

		auth.Latest = nd.RecvTime
		err = ag.AuthColl.UpdateId(auth.Id, auth)
		if err != nil {
			fmt.Println("PUT: Error updating authorization: ", err)
			return
		}

		//Update the user's latest notification time (but not notification count since we're modifying)
		err = ag.UserinfoColl.Find(bson.M{"user_id": auth.UserID}).One(&uinfo)
		if err != nil {
			fmt.Println("POST: Error finding user info document: ", err, auth.UserID)
			return
		}

		uinfo.Latest = nd.RecvTime
		err = ag.UserinfoColl.UpdateId(uinfo.Id, uinfo)
		if err != nil {
			fmt.Println("POST: Error updating user info: ", err)
			return
		}

		//Read the rules and execute any required push actions
		//		ProcessRules(ag, nd, auth, uinfo)

		err = ag.NotifColl.UpdateId(nd.Id, nd)
		if err != nil {
			fmt.Println("PUT: Error updating Notif: ", err, " ", addr)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "PUT: Error updating Notif")
			return
		}

		ag.CollChan <- nd

	case "DELE":
		err = ag.NotifColl.Find(bson.M{"notid": addr}).One(&nd)

		if err != nil {
			fmt.Println("NotID not found: ", err, " ", addr)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "DELE: Notification ID not found")
			return
		}

		err = ag.AuthColl.Find(bson.M{"address": nd.To}).One(&auth)

		if err != nil {
			fmt.Println("Auth not found: ", err, " ", nd.To)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "PUT: Authorization not found")
			return
		}

		if checkSig(npr, w, auth, flatload) {
			return
		}

		nd.Deleted = true
		nd.RecvTime = time.Now()
		nd.UserID = auth.UserID //should already be there, but just in case

		err = ag.NotifColl.UpdateId(nd.Id, nd)
		if err != nil {
			fmt.Println("DELE: Error updating Notif: ", err, " ", addr)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "DELE: Error updating Notif")
			return
		}

	} //method switch
}

func pad64(input string) string {

	switch len(input) % 4 {
	case 0:
		return input
	case 2:
		return input + "=="
	case 3:
		return input + "="
	default:
		return input + "?" //illegal
	}
}

func collectNative(sess *mgo.Session, c chan notif.Notif) {
	var ag agent //Probably doesn't belong in Notif package
	ag.NotifColl = sess.DB("").C("notification")
	ag.AuthColl = sess.DB("").C("authorization")
	ag.UserinfoColl = sess.DB("").C("userext")
	ag.CollChan = c

	log.Fatal(http.ListenAndServe(":5342", ag))
}