#N&#x014d;tifs Agent

N&#x014d;tifs (short for notifications) provides a method for users to receive opt-in notifications from services of their choice. Examples of such services (notifiers) include emergency alerts, messages from personal devices like alarm systems, news alerts, newsletters, and advertising. A more complete description of N&#x014d;tifs can be found [here](https://altmode.org/notifs).

The N&#x014d;tifs agent is the central collection and distribution point for a given user's notifications. The complete agent actually has two parts:

1.  The data path, which accepts HTTP requests from notifiers, stores n&#x014d;tifs in a database, and generates requested alerts to the user.
2.  The management interface, which allows users to manage their active n&#x014d;tifs and authorizations, processes new authorization requests, and allows them to set up alerting methods and rules.

This repository contains the code for (1), the data path, which is considered to be the most performance sensitive. The code for (2), the management interface, is in the [notif-mgmt](https://github.com/jimfenton/notif-mgmt) repository. In addition, there is a notifier library written in Python and a simple demo application that generates n&#x014d;tifs in the [notif-notifier](https://github.com/jimfenton/notif-notifier) repository.

This code is written in Go, and has been tested using Go version 1.3.3. It interfaces with a SQL database (tested using PostgreSQL 9.4.9), in which n&#x014d;tifs, authorizations, methods, rules, and user settings are stored. It also uses the following library that may require separate installation:

* [UUID](https://github.com/pborman/uuid)

This can be installed with the `go get` command.

The SQL database used by the N&#x014d;tifs agent is specified through a configuration file that is located at `/etc/notifs/agent.cfg` . This file contains a bit of JSON to specify the hostname, username, database name, and password for the database. For example, it might contain:

`{"host":"localhost","dbname":"notifs","user":"notifs","password":"whatever"}`

It is highly recommended that the database be password protected (and not with "whatever")!

The agent does not attempt to daemonize itself. One way to run the agent in the background is to use `nohup notif-agent &`
