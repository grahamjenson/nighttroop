package nightscoutapi

import (
	"cmp"

	"nighttroop.t1d.nz/server/nightscoutapi/nsutils"
	"nighttroop.t1d.nz/socketio"
)

// alarmReceivers is the room of /alarm sockets that can read, which the site's alarms will go to
const alarmReceivers = "AlarmReceivers"

// alarmSubscribeMessage is an /alarm subscribe: an access token, or as the web app sends, a JWT or API secret
type alarmSubscribeMessage struct {
	AccessToken string `json:"accessToken"`
	JWTToken    string `json:"jwtToken"`
	Secret      string `json:"secret"`
}

// onAlarmConnect is API v3's alarms on /alarm: a socket subscribes with its access token, or as the web app does
// with its API secret or JWT, and joins AlarmReceivers when it can read. The site has no alarms to send yet.
func (ws *Websocket) onAlarmConnect(socket *socketio.Socket) {
	onCredentials(ws, socket, "subscribe", func(message alarmSubscribeMessage) any {
		return ws.subscribeToAlarms(socket, message.AccessToken, message.JWTToken, message.Secret)
	})
}

// subscribeToAlarms answers as Nightscout does: any credential given must be good, and the web app, which may
// have none yet, is told what it may do. Acks are ignored until the site has alarms.
func (ws *Websocket) subscribeToAlarms(socket *socketio.Socket, accessToken, jwtToken, secret string) map[string]any {
	token := cmp.Or(accessToken, jwtToken)
	if secret == "null" {
		secret = ""
	}
	_, identity := ws.identity(socket.Request(), token, secret)
	if identity == nil && token+secret != "" {
		return map[string]any{"success": false, "message": "Missing or bad accessToken"}
	}
	read := identity != nil && nsutils.HasPermission(*identity, "api:*:read")
	if read {
		socket.Join(alarmReceivers)
	} else {
		socket.Leave(alarmReceivers)
	}
	reply := map[string]any{"success": true, "message": "Subscribed for alarms"}
	if accessToken == "" {
		reply["read"], reply["ack"] = read, false
	}
	return reply
}
