package nsutils

import (
	"net/http"
)

// WriteRecords serializes records using the public Nightscout response shape.
func WriteRecords(w http.ResponseWriter, records []Record, fields []string) {
	response := make([]map[string]any, 0, len(records))
	for _, record := range records {
		if len(fields) == 0 {
			response = append(response, record.ToMap(false))
		} else {
			response = append(response, SelectFields(record, fields))
		}
	}
	WriteJSON(w, http.StatusOK, response)
}
