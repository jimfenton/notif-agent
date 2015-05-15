/*

agent.go - Prototype notification agent

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
	"encoding/base64"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/jimfenton/notif-agent/notif"
	"os"
)

func main() {

	var uinfo notif.Userinfo

	//Test stuff
	pubkey := getkey("shiny", "bluepopcorn.net")

	pkey, err := base64.StdEncoding.DecodeString(pubkey)
	if err != nil {
		fmt.Println("Key decoding error:", err, pkey)
	}

	// End test stuff

	uri := "mongodb://localhost/notif"
	sess, err := mgo.Dial(uri)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		os.Exit(1)
	}
	defer sess.Close()

	// Channel for notif collectors
	cc:= make(chan notif.Notif, 10)

	go collectNative(sess, cc)  //Listen for native notifs

        methodColl := sess.DB("").C("method")
        ruleColl := sess.DB("").C("rule")
        userinfoColl := sess.DB("").C("userext")


	for notif := range cc {

		err = userinfoColl.Find(bson.M{"user_id": notif.UserID}).One(&uinfo)
		if err != nil {
			fmt.Printf("Can't retrieve user info for push, go error %v\n", err) // non-fatal
		} else {
			ProcessRules(notif, ruleColl, methodColl, uinfo)
		}
	}
	
}
