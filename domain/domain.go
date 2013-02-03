package domain

import (
    "fmt"
    "github.com/prasmussen/gandi-api/domain"
    "github.com/prasmussen/gandi-api/domain/zone"
    "github.com/prasmussen/gandi-api/domain/zone/record"
    "github.com/prasmussen/gandi-api/domain/zone/version"
    "github.com/prasmussen/gandi-api/client"
    "github.com/prasmussen/gandi/config"
)

func New(testing bool) *Domain {
    // Load the gandi config file, which contains the api keys
    cfg := config.Load("")
    var c *client.Client
    if testing {
        c = client.New(cfg.ApiTestKey, client.Testing)
    } else {
        c = client.New(cfg.ApiProdKey, client.Production)
    }

    return &Domain{
        domain: domain.New(c),
        zone: zone.New(c),
        record: record.New(c),
        version: version.New(c),
    }
}

type Domain struct {
    domain *domain.Domain
    zone *zone.Zone
    record *record.Record
    version *version.Version
}

type Record struct {
    Name string `json:"name"`
    Value string `json:"value"`
    Type string `json:"type"`
    TTL int64 `json:"ttl"`
}

func (self *Domain) ListRecords(domainName, recordType string) ([]*Record, error) {
    list := make([]*Record, 0)

    // Find zone used on domain
    domainInfo, err := self.domain.Info(domainName)
    if err != nil {
        return nil, err
    }

    // Get list of records from active zone
    records, err := self.record.List(domainInfo.ZoneId, 0)
    if err != nil {
        return list, err
    }

    for _, r := range records {
        if r.Type != recordType {
            continue
        }

        list = append(list, &Record{
            Name: r.Name,
            Value: r.Value,
            Type: r.Type,
            TTL: r.Ttl,
        })
    }

    return list, nil
}

func (self *Domain) findRecords(zoneId, zoneVersion int64, recordName, recordType string) ([]int64, error){
    ids := make([]int64, 0)

    // Get list of records from zone
    records, err := self.record.List(zoneId, zoneVersion)
    if err != nil {
        return ids, err
    }

    // Loop over each record and find records matching name and type
    for _, record := range records {
        if record.Name == recordName && record.Type == recordType {
            ids = append(ids, record.Id)
        }
    }

    return ids, nil
}

func (self *Domain) StartTransaction(domainName string) (*Transaction, error) {
    trans := &Transaction{domain: self}

    // Find zone used on domain
    domainInfo, err := self.domain.Info(domainName)
    if err != nil {
        return nil, err
    }
    trans.zoneId = domainInfo.ZoneId

    // Create new version of zone with the same records as the active version
    trans.zoneVersion, err = self.version.New(trans.zoneId, 0)
    if err != nil {
        return nil, err
    }

    return trans, nil
}

type Transaction struct {
    domain *Domain
    zoneId int64
    zoneVersion int64
}

func (self *Transaction) DeleteRecords(recordName, recordType string) error {
    // Find all records with matching record name and type
    recordIds, err := self.domain.findRecords(self.zoneId, self.zoneVersion, recordName, recordType)
    if err != nil {
        return err
    }

    // Delete records
    for _, recordId := range recordIds {
        ok, err := self.domain.record.Delete(self.zoneId, self.zoneVersion, recordId)
        if err != nil {
            return err
        }

        if !ok {
            return fmt.Errorf("Could not delete record %d on zone %d version %d for unknown reasons", recordId, self.zoneId, self.zoneVersion)
        }
    }

    return nil
}

func (self *Transaction) AddRecord(recordName, recordType, recordValue string, ttl int64) error {
    // Add record to zone
    _, err := self.domain.record.Add(record.RecordAdd{
        Zone: self.zoneId,
        Version:self.zoneVersion,
        Name: recordName,
        Type: recordType,
        Value: recordValue,
        Ttl: ttl,
    })

    return err
}

func (self *Transaction) Commit() error {
    // Set version active
    ok, err := self.domain.version.Set(self.zoneId, self.zoneVersion)
    if err != nil {
        return err
    } else if !ok {
        return fmt.Errorf("Failed to set version %d active on zone %d for unknown reasons", self.zoneVersion, self.zoneId)
    }
    return nil
}
