package v2

import (
	"net/http"

	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
)

// {
//   "upbat": {
//     "min": {
//       "battery": 87,
//       "value": 87,
//       "display": "87%",
//       "level": 75
//     }
//   },
//   "iob": {
//     "iob": 1.983,
//     "basaliob": -0.669,
//     "treatmentIob": 3.464
//   },
//   "cob": {
//     "cob": 12.39360495744192
//   },
//   "cage": {
//     "treatmentDate": 1722497240926,
//   },
//   "sage": {
//     "Sensor Start": {
//       "found": false
//     },
//     "Sensor Change": {
//       "treatmentDate": 1721978583000
//     },
//     "min": "Sensor Change"
//   }
// }

// PropertiesList lists properties using the Nightscout v1 query format.
func PropertiesList(dep nsutils.Dependencies) func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, identity *nsutils.Identity) {
		query, err := nsutils.ParseQuery(r.URL.Query())
		if err != nil {
			nsutils.WriteError(w, err)
			return
		}
		records, err := dep.Store.Search(r.Context(), "properties", query)
		if err != nil {
			nsutils.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		nsutils.WriteRecordsV1(w, records, nil)
	}
}
