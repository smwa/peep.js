# Peep.js

This is based on the [Peep network auralizer](http://peep.sourceforge.net/intro.html)

It receives log events via syslog and forwards them to connected browsers,
which then plays sounds accordingly.
It can also receive state events which indicate a changing value like cpu usage or connected users. These will be done via a simple http api.

This project should also include an application that runs on a server to monitor cpu and memory usage, and forward it to the main peep.js application.
