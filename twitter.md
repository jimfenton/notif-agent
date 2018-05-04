## Twitter Notif handling -- 27 March 2018 (rev 1 May 2018)

Notifs subscribes to the Twitter stream of a particular user and converts selected tweets to notifs. This is currently done using the Twitter user stream API but that is being deprecated in favor of the user activity API.

Under the current (stream) API, a separate goroutine is created for each user with non-blank Twitter credentials in the userext database. Under the new (user activity) API, only one webhook can be registered for the application with up to 15 user subscriptions associated with it, so a single goroutine will be used which will receive callbacks for all users.

When a tweet or DM is received, the Tweet is first compared with the Twitter filters table ('twitter' table in the database). A match is declared when:
- Filter is active
- Filter is not deleted
- Tweet type (tweet, DM, mention, etc.) matches filter
- Filter username, if not blank, matches tweet
- Filter keyword, if not blank, is found in tweet

If multiple matches are found, a filter yielding the highest priority notif will be used. At most one notif will be generated per tweet or DM for a given user.

When a match is found, the lifetime and priority from the filter are associated iwth the notif. The "latest" time on the filter and count are updated.

The 'tag' field found in the Twitter table was intended to provide some control over push notification rules, but not sure how that was intended to work any more.

When a tweet deletion occurs, the associated, the Deleted tag on the associated notif is set. Need a field in the notif DB to indicate the associated tweet ID, so that the notif can be deleted if necessary.


