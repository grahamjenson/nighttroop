package nswebsockets

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/socketio"
	"nighttroop.t1d.nz/utils/alarms"
	"nighttroop.t1d.nz/utils/httputils"
)

// alarmRoom is the room of an identity's /alarm sockets that can read, which its alarms go to
func alarmRoom(identityID int64) string {
	return fmt.Sprintf("AlarmReceivers-%d", identityID)
}

// announcementRoom is the room of the site's /alarm sockets that can read, which its announcements go to
const announcementRoom = "AnnouncementReceivers"

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
	onCredentials(ws, socket, "subscribe", func(message alarmSubscribeMessage) (any, bool) {
		reply, subscribed := ws.subscribeToAlarms(socket, message.AccessToken, message.JWTToken, message.Secret)
		if identityID != nil {
			socket.Leave(alarmRoom(*identityID))
			socket.Leave(announcementRoom)
		}
		if identityID = subscribed; identityID != nil {
			socket.Join(alarmRoom(*identityID))
			socket.Join(announcementRoom)
		}
		return reply, identityID != nil
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
	ctx, identity := ws.identity(socket, token, secret)
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

// snoozeAlarms is Nightscout's ack of a level, group and silence time in milliseconds, its arguments
func (ws *Websocket) snoozeAlarms(socket *socketio.Socket, identityID int64, args []json.RawMessage) {
	var level alarms.Level
	var group string
	var silenceMs int64
	for i, into := range []any{&level, &group, &silenceMs} {
		if i < len(args) {
			json.Unmarshal(args[i], into)
		}
	}
	nsutils.SnoozeAlarms(httputils.GetSiteContext(socket.Request().Context()), identityID, level, group, silenceMs)
}

// alarmNotification is a Nightscout notification, as its /alarm sockets get them
type alarmNotification struct {
	Level     alarms.Level `json:"level"`
	Title     string       `json:"title"`
	Message   string       `json:"message"`
	Group     string       `json:"group"`
	Timestamp int64        `json:"timestamp"`
	Clear     bool         `json:"clear,omitempty"`
	// IsAnnouncement marks an announcement, as Nightscout's treatment notify does
	IsAnnouncement bool `json:"isAnnouncement,omitempty"`
}

// sendAnnouncement sends a new announcement to the site's /alarm sockets, as Nightscout does for an Announcement
// treatment: its notes are the message, which AAPS shows as a notification
func sendAnnouncement(namespace *socketio.Namespace, event events.Event) {
	bundle, ok := event.Model.(*sitedb.EventGroupsBundle)
	if !ok || event.Action != events.Created || bundle.NotificationEvent == nil || bundle.NotificationEvent.NotificationType != "announcement" {
		return
	}
	n := bundle.NotificationEvent
	message := ""
	if n.Content != nil {
		message = *n.Content
	}
	namespace.EmitTo(announcementRoom, "announcement", alarmNotification{Title: n.Title, Message: message, Group: "Announcement",
		Timestamp: n.OccurredAt, IsAnnouncement: true})
}

// notifyWithin is how recent a new treatment must be to be notified, as in Nightscout's treatment notify
const notifyWithin = 10 * time.Minute

// unnotifiedTypes are the treatments Nightscout's treatment notify sends no notification for, the announcements
// having their own
var unnotifiedTypes = []string{"", string(nightscoutmodels.Announcement), string(nightscoutmodels.TempBasal), string(nightscoutmodels.TemporaryOverride)}

// treatmentLines are the fields a treatment notification's message has, with their labels and units
var treatmentLines = [][3]string{{"glucose", "BG", ""}, {"reason", "Reason", ""}, {"carbs", "Carbs", "g"}, {"insulin", "Insulin", "U"},
	{"duration", "Duration", " min"}, {"percent", "Percent", "%"}, {"absolute", "Value", ""}, {"enteredBy", "Entered By", ""}, {"notes", "Notes", ""}}

// sendTreatmentNotification sends a new treatment to the site's /alarm sockets as Nightscout's treatment notify does,
// when it is recent and not an announcement: an info notification titled with its event type, saying what it was
func sendTreatmentNotification(namespace *socketio.Namespace, record nsutils.Document) {
	fields := record.Fields(false)
	eventType, _ := fields["eventType"].(string)
	ago := time.Since(time.UnixMilli(record.OccurredAt()))
	if slices.Contains(unnotifiedTypes, eventType) || ago < 0 || ago >= notifyWithin {
		return
	}
	message := ""
	for _, line := range treatmentLines {
		if value := fields[line[0]]; value != nil && value != "" && value != 0.0 {
			message += fmt.Sprintf("\n%s: %v%s", line[1], value, line[2])
			if glucoseType, ok := fields["glucoseType"].(string); ok && line[0] == "glucose" {
				message += " (" + glucoseType + ")"
			}
		}
	}
	if eventType == string(nightscoutmodels.TemporaryTarget) && fields["duration"] == 0.0 {
		eventType, message = eventType+" Cancel", "Canceled"
	}
	namespace.EmitTo(announcementRoom, "notification", alarmNotification{Title: eventType, Message: message, Group: "default",
		Timestamp: record.OccurredAt()})
}

// alarmTitles are each alarm's title and message, as Nightscout's simplealarms and timeago say them
var alarmTitles = map[alarms.Type][2]string{
	alarms.UrgentLow:     {"Urgent LOW", "Glucose is urgently low"},
	alarms.Low:           {"Low warning", "Glucose has been low"},
	alarms.PredictedLow:  {"Low predicted", "Glucose is heading low"},
	alarms.UrgentHigh:    {"Urgent HIGH", "Glucose is urgently high"},
	alarms.High:          {"High warning", "Glucose has been high"},
	alarms.FastRise:      {"Rising fast", "Glucose is rising fast"},
	alarms.FastDrop:      {"Dropping fast", "Glucose is dropping fast"},
	alarms.NoData:        {"Stale data, check rig?", "No glucose readings are coming in"},
	alarms.DeviceBattery: {"Device battery low", "A device's battery is low, or it has stopped reporting"},
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
	namespace.EmitTo(room, event, alarmNotification{Level: t.Level(), Title: title[0], Message: title[1], Group: nsutils.AlarmGroup(t), Timestamp: at})
}
