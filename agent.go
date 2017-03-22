/*

agent.go - Prototype notification agent

Copyright (c) 2015, 2017 Jim Fenton

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
	"fmt"
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	"github.com/jimfenton/notif-agent/notif"
	"io/ioutil"
	"os"
)

type AgentDbCfg struct {
	Host string `json:"host"`
	User string `json:"user"`
	Dbname string `json:"dbname"`
	Password  string `json:"password"`
}

// Find an user record by ID
func findUser(db *sql.DB, userID int, user *notif.Userinfo) error {
	err := db.QueryRow(`SELECT id,email_username,email_server,email_port,email_authentication,email_security,twilio_sid,twilio_token,twilio_from,count,latest,created,user_id FROM userext WHERE user_id = $1`,userID).Scan(&user.Id,
		&user.EmailUsername,
		&user.EmailServer,
		&user.EmailPort,
		&user.EmailAuthentication,
		&user.EmailSecurity,
		&user.TwilioSID,
		&user.TwilioToken,
		&user.TwilioFrom,
		&user.Count,
		&user.Latest,
		&user.Created,
		&user.UserID)
	return err
}

func main() {

	var user notif.Userinfo
	var adc AgentDbCfg

	dat, err := ioutil.ReadFile("/etc/notifs/agent.cfg") //keeps passwords out of source code
	fmt.Printf("%s\n",string(dat))
	err = json.Unmarshal(dat, &adc)
	if err != nil {
		fmt.Printf("DB config unmarshal error\n")
		os.Exit(1)
	}

	// Database parameters are stored in JSON form in /etc/notifs/agent.cfg
	// Sample configuration:
	// {"host":"localhost","dbname":"notifs","user":"notifs","password":"whatever"}
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s host=%s password=%s",adc.User,adc.Dbname,adc.Host,adc.Password))
	if err != nil {
		fmt.Printf("Can't connect to database, go error %v\n", err)
		os.Exit(1)
	}
	
	defer db.Close()

	// Channel for notif collectors
	cc:= make(chan notif.Notif, 10)

	go collectNative(db, cc)  //Listen for native notifs

	for notif := range cc {

		err := findUser(db, notif.UserID, &user)
		if err != nil {
			fmt.Printf("Can't retrieve user info for push, go error %v\n", err) // non-fatal
		} else {
			ProcessRules(notif, db, user)
		}
	}
	
}
