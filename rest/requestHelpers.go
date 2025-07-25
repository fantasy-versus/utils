package rest

import (
	"io"
	"net/http"

	"github.com/fantasy-versus/utils/decoder"
	"github.com/fantasy-versus/utils/log"
)

func DecodeRequestIntoStruct(w http.ResponseWriter, r *http.Request, dest interface{}) error {

	body2, err := io.ReadAll(r.Body)

	if err != nil {
		log.Errorf("Error with received json, seems to be invalid: no extra inf. %+v", err)
		ReturnRawError(w, "INVALID_DATA", "Review sent data", http.StatusForbidden)
		return err
	}
	defer r.Body.Close()

	err = decoder.JsonNumberDecode(body2, &dest)

	if err != nil {
		log.Errorf("Error with received json, cannot be decoded into NewWalletUserDTO. %+v", err)
		ReturnRawError(w, "JSON_ERROR", err.Error(), http.StatusBadRequest)
		return err
	}
	return nil
}
