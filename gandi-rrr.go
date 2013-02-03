package main

import (
    "os"
    "os/signal"
    "net/http"
    "encoding/json"
    "flag"
    "fmt"
    "github.com/prasmussen/gandi-rrr/config"
    "github.com/prasmussen/gandi-rrr/responder"
    "github.com/prasmussen/gandi-rrr/log"
    api "github.com/prasmussen/gandi-rrr/domain"
)

const (
    VersionNumber = "1.0.0"
)

type Request struct {
    Name string
    Value string
}

func (self *Request) IsValid() bool {
    return self.Name != "" && self.Value != ""
}

type RecordListResponse struct {
    Domain string `json:"domain"`
    Records []*api.Record `json:"records"`
}

func handleGET(token *config.Token, domain *api.Domain, l *log.RequestLogger, r *responder.Responder, recordName string) {
    // Get list of records for specified type and domain
    records, err := domain.ListRecords(token.Domain, token.Type)
    if err != nil {
        r.Error(err)
        l.Error(err)
        return
    }

    // Return list of all records if no specific record was asked for
    if recordName == "" {
        r.JSON(&RecordListResponse{
            Domain: token.Domain,
            Records: records,
        })
        l.Message(fmt.Sprintf("Returning %d records of type '%s' for domain '%s'", len(records), token.Domain, token.Type))
        return
    }

    // Return specific record
    for _, record := range records {
        if record.Name == recordName {
            r.JSON(record)
            l.Message(fmt.Sprintf("Returning record '%s' of type '%s' for domain '%s'", recordName, token.Domain, token.Type))
            return
        }
    }

    // Record not found
    err = fmt.Errorf("Record with name '%s' not found", recordName)
    r.Error(err)
    l.Error(err)
}

func handlePOST(token *config.Token, domain *api.Domain, l *log.RequestLogger, r *responder.Responder, request *Request) {
    // Start domain update transaction
    trans, err := domain.StartTransaction(token.Domain)
    if err != nil {
        r.Error(err)
        l.Error(err)
        return
    }

    // Delete all existing records with same name and type
    if err := trans.DeleteRecords(request.Name, token.Type); err != nil {
        r.Error(err)
        l.Error(err)
        return
    }

    // Add new record to domain
    if err := trans.AddRecord(request.Name, token.Type, request.Value, token.TTL); err != nil {
        r.Error(err)
        l.Error(err)
        return
    }

    // Commit changes to domain
    if err := trans.Commit(); err != nil {
        r.Error(err)
        l.Error(err)
        return
    }

    // Send success response
    r.Success()
    l.Message(fmt.Sprintf("Added [%s] '%s' - '%s' to domain '%s'", token.Type, request.Name, request.Value, token.Domain))
}

func handleDELETE(token *config.Token, domain *api.Domain, l *log.RequestLogger, r *responder.Responder, recordName string) {
    // Start domain update transaction
    trans, err := domain.StartTransaction(token.Domain)
    if err != nil {
        r.Error(err)
        l.Error(err)
        return
    }

    // Delete all existing records with same name and type
    if err := trans.DeleteRecords(recordName, token.Type); err != nil {
        r.Error(err)
        l.Error(err)
        return
    }

    // Commit changes to domain
    if err := trans.Commit(); err != nil {
        r.Error(err)
        l.Error(err)
        return
    }

    // Send success response
    r.Success()
    l.Message(fmt.Sprintf("Removed all [%s] '%s' records from domain '%s'", token.Type, recordName, token.Domain))
}

func rootHandler(cfg *config.Config, domain *api.Domain, logger *log.Logger) func(res http.ResponseWriter, req *http.Request) {
    return func(res http.ResponseWriter, req *http.Request) {
        l := logger.RequestLogger(req)
        r := responder.New(res)
        method := req.Method

        l.Message("Request start")
        defer l.Message("Request end")

        // Only accept GET, POST and delete requests
        if method != "GET" && method != "POST" && method != "DELETE"{
            err := fmt.Errorf("Invalid request method")
            r.Error(err)
            l.Error(err)
            return
        }

        // Get token from header
        tokenId := req.Header.Get("Token")

        // Lookup domain and record type from config
        token, ok := cfg.LookupToken(tokenId)
        if !ok {
            err := fmt.Errorf("Invalid token")
            r.Error(err)
            l.Error(err)
            return
        }

        recordName := req.URL.Path[1:]
        if req.Method == "GET" {
            handleGET(token, domain, l, r, recordName)
            return
        } else if req.Method == "DELETE" {
            handleDELETE(token, domain, l, r, recordName)
            return
        }

        // Unmarshal json payload
        request := &Request{}
        decoder := json.NewDecoder(req.Body)
        if err := decoder.Decode(request); err != nil {
            r.Error(err)
            l.Error(err)
            return
        }

        // Check that request is valid
        if !request.IsValid() {
            err := fmt.Errorf("Invalid request")
            r.Error(err)
            l.Error(err)
            return
        }

        handlePOST(token, domain, l, r, request)
    }
}

func main() {
    printVersion := flag.Bool("v", false, "Print version")
    flag.Parse()

    // Print version and exit if -v flag is set
    if *printVersion {
        fmt.Printf("gandi-rrr v%s\n", VersionNumber)
        return
    }

    // Load config
    cfg, err := config.Load()
    if err != nil {
        fmt.Println(err)
        return
    }

    // Create logger
    logger, err := log.New(cfg.LogFile)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Handle os interrupt and kill signals
    signals := make(chan os.Signal, 1)
    signal.Notify(signals)
    go func() {
        for sig := range signals {
            if sig == os.Interrupt || sig == os.Kill {
                fmt.Println("Shutting down...")
                logger.Close()
                os.Exit(0)
            }
        }
    }()

    // Instantiate domain api
    domain := api.New(cfg.Testing)

    // Handle requests
    http.HandleFunc("/", rootHandler(cfg, domain, logger))

    // Start web server
    if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
        fmt.Printf("Listening for https connections on %s\n", cfg.Listen)
        err = http.ListenAndServeTLS(cfg.Listen, cfg.TLS.CertFile, cfg.TLS.KeyFile, nil)
    } else {
        fmt.Printf("Listening for http connections on %s\n", cfg.Listen)
        err = http.ListenAndServe(cfg.Listen, nil)
    }
    fmt.Println(err)
}
