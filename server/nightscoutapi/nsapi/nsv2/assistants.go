package nsv2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/utils/apiutils"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/httputils"
)

const (
	unknownIntent = "I'm sorry, I don't know what you're asking for."
	unknownValue  = "That value is unknown at the moment. Please see your Nightscout site for more details."
	welcome       = "What would you like to check on Nightscout?"
)

// assistantMetrics are the names Nightscout's plugins give their metrics, for a MetricNow intent, by what is said
var assistantMetrics = map[string]string{"bg": "bg", "blood glucose": "bg", "number": "bg", "delta": "delta",
	"blood glucose delta": "delta", "iob": "iob", "insulin on board": "iob", "cob": "cob", "carbs on board": "cob",
	"carbohydrates on board": "cob", "uploader battery": "upbat"}

// assistantDirections are how Nightscout's virtual assistants say each direction
var assistantDirections = map[string]string{"DoubleUp": "rapidly rising", "SingleUp": "rising", "FortyFiveUp": "slightly rising",
	"Flat": "holding", "FortyFiveDown": "slightly falling", "SingleDown": "falling", "DoubleDown": "rapidly falling"}

// assistantAnswer is what a virtual assistant says to an intent: its card's title and the speech, and whether the
// intent was known
type assistantAnswer struct {
	title, speech string
	known         bool
}

// answerIntent answers Nightscout's virtual assistant intents, NSStatus, UploaderBattery and MetricNow of a metric,
// from the site's Nightscout properties, saying glucose in the settings' units
func answerIntent(r *http.Request, intent, metric string) assistantAnswer {
	siteContext := httputils.GetSiteContext(r.Context())
	now := time.Now()
	properties := Properties(r.Context(), siteContext.SiteDB.Queries(), now)
	units := *siteContext.GetSettings().GetUnits()
	// Whole mg/dL, or mmol/L to one decimal
	decimals := 0
	if units == consts.UnitsMMOLL {
		decimals = 1
	}
	glucose := func(mgdl float64) string { return fmt.Sprintf("%.*f", decimals, units.FromMGDL(mgdl)) }
	bg := unknownValue
	if bucket, ok := properties["bgnow"].(Bucket); ok {
		direction := "unknown"
		if d, ok := properties["direction"].(Direction); ok && assistantDirections[d.Value] != "" {
			direction = assistantDirections[d.Value]
		}
		bg = fmt.Sprintf("%s and %s as of %s.", glucose(bucket.Last), direction, timeAgo(now, bucket.Mills))
	}
	iob := "You have no insulin on board"
	if i := properties["iob"].(IOB); i.Display != nil {
		iob = fmt.Sprintf("You have %s units of insulin on board", *i.Display)
	}

	delta, cob, battery := unknownValue, "You have no carbohydrates on board", unknownValue
	if d, ok := properties["delta"].(Delta); ok {
		delta = fmt.Sprintf("Your delta was %+.*f between %s and %s.", decimals, units.FromMGDL(d.Mgdl), timeAgo(now, d.Times["recent"]), timeAgo(now, d.Times["previous"]))
	}
	if c := properties["cob"].(COB); c.Display != nil {
		cob = fmt.Sprintf("You have %v carbohydrates on board", *c.Display)
	}
	if b := properties["upbat"].(UptimeBattery); b.Display != nil {
		battery = "Your uploader battery is at " + *b.Display
	}
	answers := map[string]assistantAnswer{"bg": {"Current BG", bg, true}, "delta": {"Blood Glucose Delta", delta, true},
		"iob": {"Current IOB", iob, true}, "cob": {"Current COB", cob, true}, "upbat": {"Uploader Battery", battery, true}}

	switch intent {
	case "NSStatus":
		return assistantAnswer{"Full Status", bg + " " + iob + ".", true}
	case "UploaderBattery":
		return answers["upbat"]
	}
	if answer, ok := answers[assistantMetrics[strings.ToLower(metric)]]; ok && intent == "MetricNow" {
		return answer
	}
	return assistantAnswer{"Unknown Intent", unknownIntent, false}
}

// timeAgo is how long ago a time was, as Nightscout's virtual assistants say it
func timeAgo(now time.Time, mills int64) string {
	switch ago := now.Sub(time.UnixMilli(mills)); {
	case ago < 45*time.Second:
		return "a few seconds ago"
	case ago < 90*time.Second:
		return "a minute ago"
	case ago < 45*time.Minute:
		return fmt.Sprintf("%.0f minutes ago", ago.Minutes())
	case ago < 90*time.Minute:
		return "an hour ago"
	default:
		return fmt.Sprintf("%.0f hours ago", ago.Hours())
	}
}

// alexaRequest is the part of an Alexa skill's request Nightscout's skill reads: its type, intent and metric slot.
// JSON's names match its fields', whatever their case.
type alexaRequest struct {
	Request struct {
		Type   string
		Intent struct {
			Name  string
			Slots struct {
				Metric struct {
					Resolutions struct{ ResolutionsPerAuthority []alexaResolution }
				}
			}
		}
	}
}

// alexaResolution is what a slot was heard as: its status code, and when matched the values it matched
type alexaResolution struct {
	Status struct{ Code string }
	Values []struct{ Value struct{ Name string } }
}

// Alexa is Nightscout's /api/v1/alexa, which its Alexa skill calls: a launch is welcomed, a session ending is
// answered with nothing, and an intent with what Nightscout's skill says
func Alexa() func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, _ *nsutils.Identity) {
		var request alexaRequest
		json.NewDecoder(r.Body).Decode(&request)
		answer, reprompt, end := assistantAnswer{"Welcome to Nightscout", welcome, true}, welcome, false
		switch request.Request.Type {
		case "SessionEndedRequest":
			apiutils.WriteJSON(w, http.StatusOK, "")
			return
		case "IntentRequest":
			intent, metric := request.Request.Intent, ""
			if p := intent.Slots.Metric.Resolutions.ResolutionsPerAuthority; len(p) > 0 && p[0].Status.Code == "ER_SUCCESS_MATCH" && len(p[0].Values) > 0 {
				metric = p[0].Values[0].Value.Name
			}
			answer, reprompt, end = answerIntent(r, intent.Name, metric), "", true
		}
		speech := func(text string) map[string]string { return map[string]string{"type": "PlainText", "text": text} }
		apiutils.WriteJSON(w, http.StatusOK, map[string]any{"version": "1.0", "response": map[string]any{
			"outputSpeech": speech(answer.speech), "card": map[string]string{"type": "Simple", "title": answer.title, "content": answer.speech},
			"reprompt": map[string]any{"outputSpeech": speech(reprompt)}, "shouldEndSession": end}})
	}
}

// googleHomeRequest is the part of a Dialogflow fulfilment request Nightscout's Google Home action reads
type googleHomeRequest struct {
	QueryResult struct {
		Intent     struct{ DisplayName string }
		Parameters struct{ Metric string }
	}
}

// GoogleHome is Nightscout's /api/v1/googlehome, which its Google Home action calls through Dialogflow: an intent is
// answered with what Nightscout's action says, and an unknown one keeps listening
func GoogleHome() func(http.ResponseWriter, *http.Request, *nsutils.Identity) {
	return func(w http.ResponseWriter, r *http.Request, _ *nsutils.Identity) {
		var request googleHomeRequest
		json.NewDecoder(r.Body).Decode(&request)
		answer := answerIntent(r, request.QueryResult.Intent.DisplayName, request.QueryResult.Parameters.Metric)
		apiutils.WriteJSON(w, http.StatusOK, map[string]any{"payload": map[string]any{"google": map[string]any{
			"expectUserResponse": !answer.known,
			"richResponse":       map[string]any{"items": []any{map[string]any{"simpleResponse": map[string]string{"textToSpeech": answer.speech}}}}}}})
	}
}
