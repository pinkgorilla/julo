package http

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	w.Header().Add("content-type", "application/json")
	if data == nil {
		return
	}

	if e := json.NewEncoder(w).Encode(data); e != nil {
		log.Print(e)
	}
}

func WriteErrorJSON(w http.ResponseWriter, status int, err error) {
	var res Response
	res.Status = "failed"
	res.Data = map[string]interface{}{"error": err.Error()}

	bs, _ := json.Marshal(res)
	log.Println(string(bs))
	WriteJSON(w, status, res)
}
