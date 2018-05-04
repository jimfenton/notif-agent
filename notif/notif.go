/*

notif.go - Notification definitions and utilities

Copyright (c) 2015, 2017, 2018 Jim Fenton

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

package notif

import (
	"database/sql"
	"fmt"
	"time"
)

type NotifPri uint8

const (
	PriEmergency = iota + 1
	PriPriority
	PriRoutine
	PriInformational
)

// Representations of various data structures in PostgreSQL

// Table notification -- Notification table
type Notif struct { //Notification record in database
	Id          int       //Database: "id" //Sequence number
	To          string    //Database: "to" //Destination address (UUID if native)
	Description string    //Database: "description" //Description of authorization
	Origtime    time.Time //Database: "origtime" //Notif origination timestamp
	Priority    NotifPri  //Database: "priority" //Notif priority
	From        string    //Database: "fromdomain"  //TODO: rename field 'from'
	Expires     time.Time //Database: "expires" //Notif expiration timestamp
	Subject     string    //Database: "subject" //Notif subject header
	Body        string    //Database: "body" //Notif body text
	NotID       string    //Database: "notID" //Assigned notif ID handle
	RecvTime    time.Time //Database: "recvtime" //Notif received time (updated on notif update)
	RevCount    int       //Database: "revcount" //Notif revision count (0 initially)
	Read        bool      //Database: "read" //Notif seen flag (true if seen)
	ReadTime    time.Time //Database: "readtime" //Notif seen timestamp
	Deleted     bool      //Database: "deleted" //Notif deletion flag (true if deleted)
	Source      string    //Database: "source" //Notif source, e.g., "native", "tweet"
	UserID      int       //Database: "user_id"
}

// Table authorization -- native notif authorizations
type Auth struct {
	Id          int       //Database: "id" //Sequence number
	UserID      int       //Database: "user_id" //ID of user owning authorization
	Address     string    //Database: "address" //UUID address of destination address
	Domain      string    //Database: "domain" //Name of source domain
	Description string    //Database: "description" //Human-readable description of notification
	Created     time.Time //Database: "created" //Authorization creation timestamp (not used in agent)
	Maxpri      NotifPri  //Database: "maxpri" //Maximum priority for authorization
	Latest      time.Time //Database: "latest" //Latest notif timestamp
	Count       int       //Database: "count" //Count of notifs received for this authorization
	Active      bool      //Database: "active" //Authorization active flag (true if active)
	Expiration  time.Time //Database: "expiration" //Timestamp of authorization lifetime (implemented???)
	Deleted     bool      //Database: "deleted" //Authorization deletion flag (true if deleted)
}

// Per-user settings and information
type Userinfo struct {
	Id                       int       //Database: "_id"
	UserID                   int       //Database: "user_id" //Perhaps redundant??
	Count                    int       //Database: "count"
	EmailUsername            string    //Database: "email_username"
	EmailServer              string    //Database: "email_server"
	EmailPort                int       //Database: "email_port"
	EmailFrom                string    //Database: "email_from"
	EmailSecurity            int       //Database: "email_security"
	EmailAuthentication      int       //Database: "email_authentication"
	Created                  time.Time //Database: "created"
	Latest                   time.Time //Database: "latest"
	TwilioSID                string    //Database: "twilio_sid"  (Overrides site setting if present)
	TwilioToken              string    //Database: "twilio_token"
	TwilioFrom               string    //Database: "twilio_from"
	TwitterAccessToken       string    //Database: "twitter_access_token"
	TwitterAccessTokenSecret string    //Database: "twitter_access_token_secret"
}

type Rule struct {
	Id       int      //Database: "_id"
	UserID   int      //Database: "user_id"
	Domain   string   //Database: "domain"
	Priority NotifPri //Database: "priority"
	Active   bool     //Database: "active"
	Method   int      //Database: "method_id"
}

// Global settings for the site
// Twilio values overridden by user settings if present
type Siteinfo struct {
	TwilioSID             string //Database: "twilio_sid"
	TwilioToken           string //Database: "twilio_token"
	TwilioFrom            string //Database: "twilio_from"
	TwitterConsumerKey    string //Database: "twitter_consumer_key"
	TwitterConsumerSecret string //Database: "twitter_consumer_secret"
}

//Store a notif in the database

func (n Notif) Store(db *sql.DB) error {
	stmt, err := db.Prepare(`INSERT INTO notification (user_id,toaddr,description,origtime,priority,fromdomain,expires,subject,body,notid,recvtime,revcount,read,readtime,source,deleted) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`)
	if err != nil {
		fmt.Println("Notification insert prepare error: ", err)
		return err
	}
	_, err = stmt.Exec(n.UserID, n.To, n.Description, n.Origtime, n.Priority, n.From, n.Expires,
		n.Subject, n.Body, n.NotID, n.RecvTime, 0, false, nil, n.Source, false)
	if err != nil {
		fmt.Println("Notification insert error: ", err)
	}
	return err
}
