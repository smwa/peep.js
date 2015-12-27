package main

import "log"
import "os"
import "io"
import "io/ioutil"
import "net"
import "net/http"
import "time"
import "strconv"
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
    Intensity float64
}

func init() {
    var iow io.Writer
    if os.Getenv("DEBUG") != "" {
        iow = os.Stdout
    } else {
        iow = ioutil.Discard
    }
    Debug = log.New(iow, "Debug: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
    go startHttpServer()
    go startSyslogServer()
    select{}
}

func startHttpServer() {
    http.Handle("/", http.FileServer(http.Dir("static")))
    http.Handle("/websocket", websocket.Handler(websocketOnConnect))
    staticListener, err := net.Listen("tcp", httpPort)
    for err != nil {
        Debug.Println("Creating http server failed: ", err)
        Debug.Println("Retrying in 5")
        time.Sleep(5 * time.Second)
        staticListener, err = net.Listen("tcp", httpPort)
    }
    Debug.Println("Started http server")
    http.Serve(staticListener, nil);
}

func startSyslogServer() {
    channel := make(syslog.LogPartsChannel)
    handler := syslog.NewChannelHandler(channel)

    server := syslog.NewServer()
    server.SetFormat(syslog.RFC3164)
    server.SetHandler(handler)
    server.ListenUDP(syslogUdpPort)
    server.Boot()
    processLogparts(channel)
}

func processLogparts(channel syslog.LogPartsChannel) {
    for logParts := range channel {
        Debug.Println("syslog ", logParts)
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
        //state
        if event.Appname == "cpu_state" || event.Appname == "memory_state" {
            event.Intensity, _ = strconv.ParseFloat(logParts["content"].(string), 64)
        }
        eventStringified, _ := json.Marshal(event)
        messageAllWebsockets(eventStringified)
    }
}


func websocketOnConnect(ws *websocket.Conn) {
    Debug.Println("New connection added")
    websocketconnections = append(websocketconnections, ws)
    defer func() {
        if err := ws.Close(); err != nil {
            Debug.Println("Websocket could not be closed", err.Error())
        }
    }()
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

