package responder

import (
    "encoding/json"
    "net/http"
    "fmt"
)

type SuccessResponse struct {
    Success bool `json:"success"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}

type Responder struct {
    res http.ResponseWriter
}

func New(res http.ResponseWriter) *Responder {
    return &Responder{res}
}

func (self *Responder) JSON(v interface{}) {
    self.res.Header().Set("Content-Type", "application/json")
    json.NewEncoder(self.res).Encode(v)
}

func (self *Responder) Success() {
    self.JSON(&SuccessResponse{
        Success: true,
    })
}

func (self *Responder) Error(err error) {
    self.JSON(&ErrorResponse{
        Error: fmt.Sprintf("%v", err),
    })
}

