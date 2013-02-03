package log

import (
    "io"
    "os"
    "fmt"
    "encoding/json"
    "crypto/md5"
    "time"
    "net/http"
)

type Event struct {
    Timestamp int64 `json:"timestamp"`
    Id string `json:"id"`
    Method string `json:"method"`
    Uri string `json:"uri"`
    Host string `json:"host"`
    RemoteAddr string `json:"remoteAddress"`
    UserAgent string `json:"userAgent"`
    Message string `json:"message"`
}

type Logger struct {
    file *os.File
}

func New(fname string) (*Logger, error) {
    var f *os.File

    // Log to stdout if no filename was provided
    if fname == "" {
        f = os.Stdout
    } else {
        var err error
        f, err = os.OpenFile(fname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
        if err != nil {
            return nil, err
        }
    }

    return &Logger{f}, nil
}

func (self *Logger) JSON(v interface{}) {
    json.NewEncoder(self.file).Encode(v)
}

func (self *Logger) Close() {
    self.file.Close()
}

func (self *Logger) RequestLogger(req *http.Request) *RequestLogger {
    // Generate a uniq request id
    h := md5.New()
    uniqId := fmt.Sprintf("%d %s", time.Now().UnixNano(), req.RemoteAddr)
    io.WriteString(h, uniqId)

    // Add static fields
    event := &Event{
        Id: fmt.Sprintf("%x", h.Sum(nil)),
        Method: req.Method,
        Uri: req.URL.RequestURI(),
        Host: req.Host,
        RemoteAddr: req.RemoteAddr,
        UserAgent: req.Header.Get("user-agent"),
    }

    return &RequestLogger{self, event}
}

type RequestLogger struct {
    *Logger
    event *Event
}

func (self *RequestLogger) Message(msg string) {
    e := self.event
    e.Timestamp = time.Now().Unix()
    e.Message = msg
    self.JSON(e)
}

func (self *RequestLogger) Error(err error) {
    self.Message(fmt.Sprintf("%v", err))
}
