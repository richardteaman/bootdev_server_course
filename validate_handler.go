package main

import (
	"encoding/json"
	"net/http"
)

func validateHandler(w http.ResponseWriter, r *http.Request) {
	type parametrs struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parametrs{}
	err := decoder.Decode(&params)
	if err != nil {
		type errReturnVals struct {
			// the key will be the name of struct field unless you give it an explicit JSON tag

			Err string `json:"error"`
		}

		respBody := errReturnVals{
			Err: "Something went wrong",
		}
		dat, _ := json.Marshal(respBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(dat)
		return
	}
	if len(params.Body) > 140 {
		type errReturnVals struct {
			Err string `json:"error"`
		}

		respBody := errReturnVals{
			Err: "Chirp is too long",
		}
		dat, _ := json.Marshal(respBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	type returnVals struct {
		Valid bool `json:"valid"`
	}
	respBody := returnVals{
		Valid: true,
	}
	dat, _ := json.Marshal(respBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}
