/*

push.go - Push notifications for prototype notification agent

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
	"bitbucket.org/ckvist/twilio/twirest"
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/jimfenton/notif-agent/notif"
	"net/url"
	"regexp"
	"strings"
)

type Method struct {
	Id       int //`bson:"_id"`
	User     int //`bson:"user_id"`
	Active   bool          //`bson:"active"`
	Name     string        //`bson:"name"`
	Mode     int           //`bson:"type"` //TODO: Change field name to "mode"
	Address  string        //`bson:"address"`
	Preamble string        //`bson:"preamble"`
}

const (
	ModeEmail = iota
	ModeText
	ModeVoice
)

func ProcessRules(n notif.Notif, db *sql.DB, user notif.Userinfo, site notif.Siteinfo) {
	var m Method
	var r notif.Rule
	var u []int
	rules, err := db.Query(`SELECT active, priority, domain, method_id FROM rule WHERE user_id = $1`, n.UserID)
	if err != nil {
		fmt.Println("Push: Ruleset query error: ", err, " user ", n.UserID)
		return
	}


ruleloop:
	for rules.Next() {
		err = rules.Scan(&r.Active, &r.Priority, &r.Domain, &r.Method)
		if err != nil {
			fmt.Println("Push: Rule scan error: ",err)
			continue
		}

		if r.Active &&
			(r.Domain == "" || r.Domain == n.From) &&
			(r.Priority == 0 || r.Priority == n.Priority) {

			// check to make sure each method only executed once per notif
			for _, mu := range u {
				if mu == r.Method {
					continue ruleloop
				} // if mu
			} // for mu
			u = append(u, r.Method)
			err = db.QueryRow(`SELECT id, user_id, active, name, type, address, preamble FROM method WHERE id = $1`,r.Method).Scan(&m.Id, &m.User, &m.Active, &m.Name, &m.Mode, &m.Address, &m.Preamble)
			if err != nil {
				fmt.Println("Push: Method query error: ",err)
				continue
			}
			doMethod(m, n, user, site)
		} //if r.Active...

	} // for rules.Next (ruleloop)
}

func doMethod(m Method, n notif.Notif, user notif.Userinfo, site notif.Siteinfo) {
	twilioSID := site.TwilioSID
	twilioToken := site.TwilioToken
	twilioFrom := site.TwilioFrom

	if user.TwilioSID != "" {
		twilioSID = user.TwilioSID
		twilioToken = user.TwilioToken
		twilioFrom = user.TwilioFrom
	}
	
	switch m.Mode {
	case ModeText:
		if m.Address == "" {
			fmt.Println("Can't send text: method address empty")
			return
		}
		if twilioFrom == "" {
			fmt.Println("Can't send text: user 'from' phone number empty")
			return
		}
		twclient := twirest.NewClient(twilioSID, twilioToken)
		//TODO: should probably cache twclient for reuse (are we leaking these now?)
		msg := twirest.SendMessage{
			Text: m.Preamble + ": " + n.Subject,
			To:   e164norm(m.Address),
			From: e164norm(twilioFrom)}
		_, err := twclient.Request(msg)
		if err != nil {
			fmt.Println("Twilio text request error: ", err)
			return
		}

	case ModeVoice:
		if m.Address == "" {
			fmt.Println("Can't send voice message: method address empty")
			return
		}
		if twilioFrom == "" {
			fmt.Println("Can't send voice message: user 'from' phone number empty")
			return
		}
		twclient := twirest.NewClient(twilioSID, twilioToken)
		//TODO: again, need to cache this (probably in Userinfo)
		twimlurl := "http://twimlets.com/message?Message%5B0%5D=" + url.QueryEscape(m.Preamble+" "+n.Subject)

		msg := twirest.MakeCall{
			From: e164norm(twilioFrom),
			To:   e164norm(m.Address),
			Url:  twimlurl}
		_, err := twclient.Request(msg)
		if err != nil {
			fmt.Println("Twilio voice request error: ", err)
			return
		}

	} // switch m.mode
} // doMethod

//Normalize a string to E.164 format
func e164norm(ph string) string {
	if ph == "" {
		return ""
	}

	p := strings.Replace(ph, "-", "", -1)
	p = strings.Replace(p, " ", "", -1)
	p = strings.Replace(p, "(", "", -1)
	p = strings.Replace(p, ")", "", -1)
	p = strings.Replace(p, ".", "", -1)

	if p[0] != '+' {
		p = "+1" + p
	}

	matched, err := regexp.MatchString("^[0-9]*$", p[1:])

	if matched {
		return p
	}

	if err != nil {
		fmt.Println("E.164 normalization error: ", err)
		return p
	}

	fmt.Println("Illegal phone number: ", ph)
	return p
}
