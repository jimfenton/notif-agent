/*

notif.go - Notification definitions and utilities

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

package notif

import (
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

type Notif struct { //Notification document in MongoDB database
	Id          int       //Database: "_id"
	To          string    //Database: "to"
	Description string    //Database: "description"
	Origtime    time.Time //Database: "origtime"
	Priority    NotifPri  //Database: "priority"
	From        string    //Database: "fromdomain"  //TODO: rename field 'from'
	Expires     time.Time //Database: "expires"
	Subject     string    //Database: "subject"
	Body        string    //Database: "body"
	NotID       string    //Database: "notID"
	RecvTime    time.Time //Database: "recvtime"
	RevCount    int       //Database: "revcount"
	Read        bool      //Database: "read"
	ReadTime    time.Time //Database: "readtime"
	Deleted     bool      //Database: "deleted"
	Source      string
	UserID      int //Database: "user_id"
}

type Auth struct {
	Id          int       //Database: "_id"
	UserID      int       //Database: "user_id"
	Address     string    //Database: "address"
	Domain      string    //Database: "domain"
	Description string    //Database: "description"
	Created     time.Time //Database: "created" //not used in agent
	Maxpri      NotifPri  //Database: "maxpri"
	Latest      time.Time //Database: "latest"
	Count       int       //Database: "count"
	Active      bool      //Database: "active"
	Expiration  time.Time //Database: "expiration"
	Deleted     bool      //Database: "deleted"
}

// Per-user settings and information
type Userinfo struct {
	Id                  int       //Database: "_id"
	UserID              int       //Database: "user_id"
	Count               int       //Database: "count"
	EmailUsername       string    //Database: "email_username"
	EmailServer         string    //Database: "email_server"
	EmailPort           int       //Database: "email_port"
	EmailFrom           string    //Database: "email_from"
	EmailSecurity       int       //Database: "email_security"
	EmailAuthentication int       //Database: "email_authentication"
	Created             time.Time //Database: "created"
	Latest              time.Time //Database: "latest"
	TwilioSID           string    //Database: "twilio_sid"  (Overrides site setting if present)
	TwilioToken         string    //Database: "twilio_token"
	TwilioFrom          string    //Database: "twilio_from"
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
// overridden by user settings if present
type Siteinfo struct {
	TwilioSID   string //Database: "twilio_sid"
	TwilioToken string //Database: "twilio_token"
	TwilioFrom  string //Database: "twilio_from"
}
