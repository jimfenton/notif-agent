/*

twitter.go - Source notifs from Twitter

Copyright (c) 2018 Jim Fenton

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
	//       "net/http"
	//	"strings"
	"database/sql"
	"fmt"
	"time"
	"github.com/ChimeraCoder/anaconda"
	"github.com/jimfenton/notif-agent/notif"
	"net/url"
)

func doTweets(db *sql.DB, c chan notif.Notif, site notif.Siteinfo) {

	var u notif.Userinfo

	users, err := db.Query(`SELECT user_id, twitter_access_token, twitter_access_token_secret FROM userext WHERE twitter_access_token IS NOT NULL AND twitter_access_token_secret IS NOT NULL`)
	if err != nil {
		fmt.Println("Twitter: User config query error: ", err)
		return
	}

	anaconda.SetConsumerKey(site.TwitterConsumerKey)
	anaconda.SetConsumerSecret(site.TwitterConsumerSecret)

	for users.Next() {
		err = users.Scan(&u.UserID, &u.TwitterAccessToken, &u.TwitterAccessTokenSecret)
		if err != nil {
			fmt.Println("Twitter: User config scan error: ", err)
			continue
		}

		go collectTweets(db, c, u)
	}
}

func filterTweets(db *sql.DB, t anaconda.Tweet, u notif.Userinfo) (notif.Notif, error) {
	var n notif.Notif
	var pri notif.NotifPri
	var lifetime time.Duration
	var tag string

	//TODO: Filter on tweet type (but maybe not DM)
	err := db.QueryRow(`SELECT priority, lifetime, tag FROM twitter WHERE active == TRUE AND deleted == FALSE AND user_id == ? AND (source IS NULL or source == ?) AND (keyword IS NULL or keyword LIKE '%?%' ORDER BY priority LIMIT 1`, u.UserID, t.User, t.Text).Scan(&pri, &lifetime, &tag)
	switch {
	case err == sql.ErrNoRows:
		return n, nil //Nothing to do with tweet. Figure out how to signal this
	case err != nil:
		{
			fmt.Println("Twitter: Tweet filter query error: ", err)
			return n, err
		}
	default:
		{ //Build a notif from the tweet
			n.To = "" //Unused for Twitter
			n.Description = tag
			n.Origtime = time.Now()
			n.Expires = n.Origtime.Add(lifetime)
			n.Subject = "@" + t.User.IdStr
			n.From = t.User.Name
			n.Description = tag
			n.Priority = pri
			n.Body = t.Text
			n.NotID = t.IdStr
			n.RecvTime = n.Origtime
			n.RevCount = 0
			n.Read = false
			n.ReadTime = n.Origtime //ignored since not read
			n.Deleted = false
			n.Source = "tweet"
			n.UserID = u.UserID

			err = n.Store(db)
		}

	}
	return n, err

}

func collectTweets(db *sql.DB, c chan notif.Notif, u notif.Userinfo) {

	api := anaconda.NewTwitterApi(u.TwitterAccessToken, u.TwitterAccessTokenSecret)
	api.SetLogger(anaconda.BasicLogger)

	v := url.Values{}
	s := api.UserStream(v)

	for msg := range s.C {
		switch msg.(type) {
		case anaconda.Tweet:
			t, ok := msg.(anaconda.Tweet)
			if !ok {
				fmt.Println("Error getting message from stream: ", ok)
				break
			}
			fmt.Print("From: ", t.User.Name, "::", t.Text, "\n")
		case anaconda.StatusDeletionNotice:
			t, _ := msg.(anaconda.StatusDeletionNotice)
			fmt.Print("Status deletion from ", t.UserIdStr, " ID ", t.IdStr, "\n")
		case anaconda.DirectMessage:
			t, _ := msg.(anaconda.DirectMessage)
			fmt.Print("DM From::", t.Sender.Name, "::", t.Text, "\n")
		default:
			fmt.Printf("Received %T message\n", msg)
		}

	}
}
