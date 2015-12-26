package main

import "log"
import "os"
import "io"
import "io/ioutil"
import "net"
import "net/http"
import "time"
import "encoding/json"
import "golang.org/x/net/websocket"
import "gopkg.in/mcuadros/go-syslog.v2"

const httpPort = ":8080"
const syslogUdpPort = ":2000"

var websocketconnections []*websocket.Conn
var Debug *log.Logger

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
    var iow io.Writer
    if os.Getenv("DEBUG") != "" {
        iow = os.Stdout
    } else {
        iow = ioutil.Discard
    }
    Debug = log.New(iow, "Debug: ", log.Ldate|log.Ltime|log.Lshortfile)
    http.Handle("/", http.FileServer(http.Dir("static")))
    http.Handle("/websocket", websocket.Handler(websocketOnConnect))
    startHttpServer()
    startSyslogServer()
    twiddleThumbs()
}

func startHttpServer() {
    staticListener, err := net.Listen("tcp", httpPort)
    for err != nil {
        Debug.Println("Creating http server failed: ", err)
        Debug.Println("Retrying in 5")
        time.Sleep(5 * time.Second)
        staticListener, err = net.Listen("tcp", httpPort)
    }
    go http.Serve(staticListener, nil);
    Debug.Println("Started http server")
}

func startSyslogServer() {
    channel := make(syslog.LogPartsChannel)
    handler := syslog.NewChannelHandler(channel)

    server := syslog.NewServer()
    server.SetFormat(syslog.RFC3164)
    server.SetHandler(handler)
    server.ListenUDP(syslogUdpPort)
    server.Boot()

    go func(channel syslog.LogPartsChannel) {
        for logParts := range channel {
            event := Event{
                Type: "Event",
            }
            if val, ok := logParts["hostname"]; ok {
                event.Hostname = val.(string);
            }
            if val, ok := logParts["tag"]; ok {
                event.Appname = val.(string);
            }
            if val, ok := logParts["app_name"]; ok {
                event.Appname = val.(string);
            }
            if val, ok := logParts["severity"]; ok {
                event.Severity = val.(int);
            }
            eventStringified, _ := json.Marshal(event)
            messageAllWebsockets(eventStringified)
            Debug.Println("syslog ", logParts)
        }
    }(channel)
}

func websocketOnConnect(ws *websocket.Conn) {
    Debug.Println("New connection added")
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
            Debug.Println("Failed to send message, removing ws connection")
            websocketconnections = append(websocketconnections[:i], websocketconnections[i+1:]...)
        }
    }
}

