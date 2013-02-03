package config

import (
    "os"
    "encoding/json"
)

const (
    DefaultConfig = "config.json"
)

type Config struct {
    Listen string
    LogFile string
    Testing bool
    TLS struct {
        CertFile string
        KeyFile string
    }
    Tokens []*Token
}

type Token struct {
    Id string
    Domain string
    Type string
    TTL int64
}

func (self *Config) LookupToken(tokenId string) (*Token, bool) {
    for _, t := range self.Tokens {
        if t.Id == tokenId {
            return t, true
        }
    }
    return nil, false
}

func Load() (*Config, error) {
    // Open config for reading
    f, err := os.Open(DefaultConfig)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    // Unmarshal config data
    decoder := json.NewDecoder(f)
    cfg := &Config{}
    err = decoder.Decode(cfg)
    if err != nil {
        return nil, err
    }

    return cfg, nil
}
