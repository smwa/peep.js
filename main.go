package main

import "fmt"
import "net"
import "net/http"
import "time"
import "golang.org/x/net/websocket"
import "encoding/json"
import "gopkg.in/mcuadros/go-syslog.v2"

const httpPort = ":8080"
const syslogUdpPort = ":2000"

var websocketconnections []*websocket.Conn

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
    channel := make(syslog.LogPartsChannel)
    handler := syslog.NewChannelHandler(channel)

    server := syslog.NewServer()
    server.SetFormat(syslog.RFC5424)
    server.SetHandler(handler)
    server.ListenUDP(syslogUdpPort)
    server.Boot()

    go func(channel syslog.LogPartsChannel) {
        for logParts := range channel {
            event := Event{
                Type: "Event",
                Hostname: logParts["hostname"].(string),
                Appname: logParts["app_name"].(string),
                Severity: logParts["severity"].(int),
            }

            eventStringified, _ := json.Marshal(event)
            messageAllWebsockets(eventStringified)
            fmt.Println("syslog", logParts)
        }
    }(channel)
}

func websocketOnConnect(ws *websocket.Conn) {
    fmt.Println("New connection added")
    websocketconnections = append(websocketconnections, ws)
    select{}
}

func twiddleThumbs() {
    select{}
}

func messageAllWebsockets(msg []byte) {
    for i,websocketconnection := range websocketconnections {
        _, err := websocketconnection.Write(msg)
        if err != nil {
            websocketconnection.Close()
            fmt.Println("Failed to send message, removing ws connection")
            websocketconnections = append(websocketconnections[:i], websocketconnections[i+1:]...)
        }
    }
}
