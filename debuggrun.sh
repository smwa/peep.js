#!/bin/bash
docker run -it -p 80:8080 -p 2000:2000/udp -e DEBUG=1 --name peep --rm smwa/peep
