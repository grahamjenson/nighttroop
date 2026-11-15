// Package can holds every rule on what an identity may do, so a reviewer can read them all in one place. A handler
// asks this package, and does not compare roles itself.
//
// Site identity roles (consts.IdentityRole):
//   - admin: administers the site (identities, settings, connectors, push credentials, logs). Has no Nightscout token.
//   - user: a person who reads and changes the site's data.
//   - user_readonly: a person who only reads the site's data.
//   - clinician: a person who sees only the printable report. Has no Nightscout token.
//   - device_pwd_phone: the PWD (person with diabetes) phone. Reads and changes all data, and pairs AAPS clients.
//   - device_read: a follower app that only reads the site's data.
//   - device_read_write: a Caregiver app. Reads all data and changes only treatments.
//
// Control identity roles (consts.ControlRole):
//   - admin: sees and changes the server's sites, admins, rate limits and push.
//   - read: sees what an admin sees, and changes nothing.
//
// Names read as a sentence after "can":
//   - can.Identity<Verb><What> is a yes or no for a site identity, e.g. can.IdentityLoginToWebsite(role).
//   - can.Control<Verb><What> is a yes or no for a control identity, e.g. can.ControlChangeServer(role).
//   - A list that a query or a client needs is a noun, e.g. can.DeviceTokenRoles and can.NightscoutPermissions.
//
// A check takes the role. A check takes the whole identity only when it must refuse a missing identity (signed out).
//
// The files group the rules by context:
//   - login.go: who signs in to the website, and who must use TOTP.
//   - site_stores.go: who may search, get, create, update and delete each store of the site API.
//   - site_admin.go: what only a site admin may do outside the stores.
//   - site_report.go: the clinician, who sees only the printable report.
//   - nightscout.go: Nightscout tokens, Nightscout permissions and AAPS device pairing.
//   - device_tokens.go: who gets the device tokens of Loop and Trio.
//   - control.go: what a control admin may do that a control reader may not.
//
// permissions_test.go has the full matrix of every role and every check.
package can
