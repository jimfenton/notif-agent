#N&#x014d;tifs agent

N&#x014d;tifs (short for notifications) provides a method for users to receive opt-in notifications from services of their choice. Examples of such services (notifiers) include emergency alerts, messages from personal devices like alarm systems, news alerts, newsletters, and advertising. A more complete description of N&#x014d;tifs can be found [here](https://altmode.org/notifs).

The N&#x014d;tifs agent is the central collection and distribution point for a given user's notifications. The complete agent actually has two parts: (1) The data path, which accepts HTTP requests from notifiers, stores n&#x014d;tifs in a database, and generates requested alerts to the user, and (2) The management interface, which allows users to manage their active n&#x014d;tifs and authorizations, processes new authorization requests, and allows them to set up alerting methods and rules. This repository contains the code for (1), the data path, which is considered to be the most performance sensitive. The code for (2), the management interface, is in the notif-management repository. In addition, there is a notifier library written in Python and a simple demo application that generates n&#x014d;tifs in the notif-notifier repository.

This code is written in Go, and has been tested using Go version 1.2.1 and 1.3. It interfaces with a MongoDB (tested using version 2.0.6) database, in which n&#x014d;tifs, authorizations, methods, rules, and user settings are stored. It also uses the following libraries that may require separate installation:

* [UUID](https://code.google.com/pgo-uuid/uuid)
* [mgo.v2 (Mongo interface)](http://gopkg.in/mgo.v2)
* [bson](http://gopkg.in/mgo.v2/bson)

Most (perhaps all) of these can be installed with the `go get` command.

The MongoDB database used by the N&#x014d;tif agent is by default located on the same server as the agent. If you want the agent to operate on a different server, change the definition of `uri` in the main program near the end of nagent.go. Also give serious consideration to password-protecting the database.

The agent does not attempt to daemonize itself. One way to run the agent in the background is to use `nohup nagent &`






