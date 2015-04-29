/*

notif.go - Notification definitions and utilities

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

package notif

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type NotifPri uint8

const (
	PriEmergency = iota + 1
	PriPriority
	PriRoutine
	PriInformational
)

type Agent struct {
	NotifColl    *mgo.Collection
	AuthColl     *mgo.Collection
	MethodColl   *mgo.Collection
	RuleColl     *mgo.Collection
	UserinfoColl *mgo.Collection
}

// Representations of various documents in MongoDB

type Notif struct { //Notification document in MongoDB database
	Id          bson.ObjectId `bson:"_id"`
	UserID      bson.ObjectId `bson:"user_id"`
	To          string        `bson:"to"`
	Description string        `bson:"description"`
	Origtime    time.Time     `bson:"origtime"`
	Priority    NotifPri      `bson:"priority"`
	FromDomain  string        `bson:"fromdomain"`
	Expires     time.Time     `bson:"expires"`
	Subject     string        `bson:"subject"`
	Body        string        `bson:"body"`
	NotID       string        `bson:"notID"`
	RecvTime    time.Time     `bson:"recvtime"`
	RevCount    int           `bson:"revcount"`
	Read        bool          `bson:"read"`
	ReadTime    time.Time     `bson:"readtime"`
	Deleted     bool          `bson:"deleted"`
}

type Auth struct {
	Id          bson.ObjectId `bson:"_id"`
	UserID      bson.ObjectId `bson:"user_id"`
	Address     string        `bson:"address"`
	Domain      string        `bson:"domain"`
	Description string        `bson:"description"`
	Created     time.Time     `bson:"created"` //not used in agent
	Maxpri      NotifPri      `bson:"maxpri"`
	Latest      time.Time     `bson:"latest"`
	Count       int           `bson:"count"`
	Active      bool          `bson:"active"`
	Expiration  time.Time     `bson:"expiration"`
	Deleted     bool          `bson:"deleted"`
}

type Userinfo struct {
	Id     bson.ObjectId `bson:"_id"`
	UserID bson.ObjectId `bson:"user_id"`
	Count  int           `bson:"count"`
	//	EmailUsername       string        `bson:"email_username"`
	EmailServer         string    `bson:"email_server"`
	EmailPort           int       `bson:"email_port"`
	EmailFrom           string    `bson:"email_from"`
	EmailSecurity       int       `bson:"email_security"`
	EmailAuthentication int       `bson:"email_authentication"`
	Created             time.Time `bson:"created"`
	Latest              time.Time `bson:"latest"`
	TwilioSID           string    `bson:"twilio_sid"`
	TwilioToken         string    `bson:"twilio_token"`
	TwilioFrom          string    `bson:"twilio_from"`
}

type Rule struct {
	Id       bson.ObjectId `bson:"_id"`
	UserID   bson.ObjectId `bson:"user_id"`
	Domain   string        `bson:"domain"`
	Priority NotifPri      `bson:"priority"`
	Active   bool          `bson:"active"`
	Method   bson.ObjectId `bson:"method_id"`
}
