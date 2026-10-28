package nswebsockets

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/socketio"
	"nighttroop.t1d.nz/utils/alarms"
	"nighttroop.t1d.nz/utils/httputils"
)

const (
	// defaultSilence is how long an ack without a silence time snoozes for, Nightscout's shortest choice
	defaultSilence = 30 * time.Minute
	// maxSilence is the longest an ack snoozes for, so one cannot silence an alarm for good
	maxSilence = 24 * time.Hour
)

// alarmRoom is the room of an identity's /alarm sockets that can read, which its alarms go to
func alarmRoom(identityID int64) string {
	return fmt.Sprintf("AlarmReceivers-%d", identityID)
}

// alarmSubscribeMessage is an /alarm subscribe: an access token, or as the web app sends, a JWT or API secret
type alarmSubscribeMessage struct {
	AccessToken string `json:"accessToken"`
	JWTToken    string `json:"jwtToken"`
	Secret      string `json:"secret"`
}

// onAlarmConnect is API v3's alarms on /alarm: a socket subscribes with its access token, or as the web app does
// with its API secret or JWT, and when it can read gets its identity's alarms, and can ack them to snooze them
func (ws *Websocket) onAlarmConnect(socket *socketio.Socket) {
	var identityID *int64
	onCredentials(ws, socket, "subscribe", func(message alarmSubscribeMessage) any {
		reply, subscribed := ws.subscribeToAlarms(socket, message.AccessToken, message.JWTToken, message.Secret)
		if identityID != nil {
			socket.Leave(alarmRoom(*identityID))
		}
		if identityID = subscribed; identityID != nil {
			socket.Join(alarmRoom(*identityID))
		}
		return reply
	})
	socket.On("ack", func(args []json.RawMessage, ack func(...any)) {
		if identityID != nil {
			ws.snoozeAlarms(socket, *identityID, args)
		}
	})
}

// subscribeToAlarms answers as Nightscout does: any credential given must be good, and the web app, which may
// have none yet, is told what it may do. It is the identity subscribed when it can read.
func (ws *Websocket) subscribeToAlarms(socket *socketio.Socket, accessToken, jwtToken, secret string) (map[string]any, *int64) {
	token := cmp.Or(accessToken, jwtToken)
	if secret == "null" {
		secret = ""
	}
	ctx, identity := ws.identity(socket.Request(), token, secret)
	if identity == nil && token+secret != "" {
		return map[string]any{"success": false, "message": "Missing or bad accessToken"}, nil
	}
	read := identity != nil && nsutils.HasPermission(*identity, "api:*:read")
	reply := map[string]any{"success": true, "message": "Subscribed for alarms"}
	if accessToken == "" {
		reply["read"], reply["ack"] = read, read
	}
	if !read {
		return reply, nil
	}
	return reply, &httputils.GetSiteContext(ctx).IdentityModel.ID
}

// snoozeAlarms is Nightscout's ack of a level, group and silence time in milliseconds: it snoozes the identity's
// alarms in that group up to that level, which the alarm checker stops alerting on its next check
func (ws *Websocket) snoozeAlarms(socket *socketio.Socket, identityID int64, args []json.RawMessage) {
	var level alarms.Level
	var group string
	var silenceMs int64
	for i, into := range []any{&level, &group, &silenceMs} {
		if i < len(args) {
			json.Unmarshal(args[i], into)
		}
	}
	silence := min(cmp.Or(time.Duration(max(silenceMs, 0))*time.Millisecond, defaultSilence), maxSilence)
	siteID := httputils.GetSiteContext(socket.Request().Context()).SiteID
	for _, t := range alarms.Types {
		if t.Level() <= level && alarmGroup(t) == group {
			snooze := livedb.SnoozeAlarmParams{SiteID: siteID, IdentityID: identityID, Type: t, Until: time.Now().Add(silence).UnixMilli()}
			if err := ws.controlDB.Live.Queries.SnoozeAlarm(context.Background(), snooze); err != nil {
				slog.Error("cannot snooze alarm", "error", err)
			}
		}
	}
}

// alarmNotification is a Nightscout notification, as its /alarm sockets get them
type alarmNotification struct {
	Level     alarms.Level `json:"level"`
	Title     string       `json:"title"`
	Message   string       `json:"message"`
	Group     string       `json:"group"`
	Timestamp int64        `json:"timestamp"`
	Clear     bool         `json:"clear,omitempty"`
}

// alarmTitles are each alarm's title and message, as Nightscout's simplealarms and timeago say them
var alarmTitles = map[alarms.Type][2]string{
	alarms.UrgentLow:  {"Urgent LOW", "Glucose is urgently low"},
	alarms.Low:        {"Low warning", "Glucose has been low"},
	alarms.UrgentHigh: {"Urgent HIGH", "Glucose is urgently high"},
	alarms.High:       {"High warning", "Glucose has been high"},
	alarms.NoData:     {"Stale data, check rig?", "No glucose readings are coming in"},
}

// alarmGroup is Nightscout's group for an alarm: Time Ago for no data, and default for the glucose alarms
func alarmGroup(t alarms.Type) string {
	if t == alarms.NoData {
		return "Time Ago"
	}
	return "default"
}

// sendAlarms sends an identity's checked alarms to its /alarm sockets as Nightscout does: an alarm, or urgent_alarm
// when urgent, for each that starts or repeats, and clear_alarm once none is alerting. Nightscout's clients show one
// alarm, so when the loudest alerting one changes without notifying, as when an urgent alarm clears or is snoozed
// while the one it quieted still alerts, they are cleared and sent the one alerting now.
func (ws *Websocket) sendAlarms(namespace *socketio.Namespace, checked alarms.Checked) {
	// The sockets were last sent the loudest of the alarms alerting before the check
	room, loudest, showing := alarmRoom(checked.IdentityID), loudestOf(checked.Alerting), loudestOf(checked.WasAlerting)

	sounded := false
	for _, notification := range checked.Notifications {
		if notification.Kind != alarms.Cleared {
			sendAlarm(namespace, room, notification.Type, notification.At)
			sounded = true
		}
	}
	clear := func() {
		namespace.EmitTo(room, "clear_alarm", alarmNotification{Title: "All Clear", Message: "No alarm is sounding",
			Group: "default", Timestamp: time.Now().UnixMilli(), Clear: true})
	}
	switch {
	case loudest == "":
		clear()
	case !sounded && loudest != showing:
		clear()
		sendAlarm(namespace, room, loudest, time.Now().UnixMilli())
	}
}

// loudestOf is the loudest alarm, the first of the most urgent, or "" when there is none
func loudestOf(alerting []alarms.Type) alarms.Type {
	var loudest alarms.Type
	for _, t := range alerting {
		if loudest == "" || t.Level() > loudest.Level() {
			loudest = t
		}
	}
	return loudest
}

// sendAlarm sends one alarm to a room's /alarm sockets: an alarm, or urgent_alarm when urgent
func sendAlarm(namespace *socketio.Namespace, room string, t alarms.Type, at int64) {
	event := "alarm"
	if t.Level() == alarms.Urgent {
		event = "urgent_alarm"
	}
	title := alarmTitles[t]
	namespace.EmitTo(room, event, alarmNotification{Level: t.Level(), Title: title[0], Message: title[1], Group: alarmGroup(t), Timestamp: at})
}
