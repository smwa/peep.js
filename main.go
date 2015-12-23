package main

import "io"
import "fmt"
import "net"
import "net/http"
import "time"
import "golang.org/x/net/websocket"
import "encoding/json"

const httpPort = ":8080"

var websocketconnections []websocket.Conn

type Event struct {
    Type string
    Hostname string
    Appname string
    Severity int
}

type State struct {
    Type string
    Hostname string
    Appname string
    Intensity float64
}

func main() {
    http.Handle("/", http.FileServer(http.Dir("static")))
    http.Handle("/websocket", websocket.Handler(websocketOnConnect))
    startHttpServer()
    startSyslogServer()
    twiddleThumbs()
}

func startHttpServer() {
    staticListener, err := net.Listen("tcp", httpPort)
    if err != nil {
        fmt.Println("Creating http server failed: ", err)
        fmt.Println("Retrying in 5")
        time.Sleep(5 * time.Second)
    }
    go http.Serve(staticListener, nil);
    fmt.Println("Started http server")
}

func startSyslogServer() {

}

func websocketOnConnect(ws *websocket.Conn) {
    io.Copy(ws,ws)
}

func twiddleThumbs() {
    for {
        time.Sleep(3 * time.Second)

        event := Event{
            Type: "Event",
            Hostname: "exampleserver",
            Appname: "Apache2",
            Severity: 1,
        }

        eventStringified, _ := json.Marshal(event)

        messageAllWebsockets(eventStringified)
    }
    select{}
}

func messageAllWebsockets(msg []byte) {
    for i,websocketconnection := range websocketconnections {
        _, err := websocketconnection.Write(msg)
        if err != nil {
            websocketconnection.Close()

            //remove i from array
            wsclength := len(websocketconnections)
            websocketconnections[i] = websocketconnections[wsclength-1]
            websocketconnections = websocketconnections[:len(websocketconnections)-1]
        }
    }
}
