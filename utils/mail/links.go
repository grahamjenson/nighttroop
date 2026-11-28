package mail

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/mail"
	"sync"
	"time"
)

// maxAddressLength is the longest address an email standard allows
const maxAddressLength = 254

// Address is whether email is one email address, as it is typed, and no longer than an email standard allows
func Address(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email && len(email) <= maxAddressLength
}

// Links are the link last sent to the email of each owner, such as a site, by its ID, which proves the email is theirs.
// An owner has at most one, and is sent one each every, so an owner cannot be used to send emails to an address that
// did not ask for them. They are kept in memory, so a restart of the server means sending the link again.
type Links struct {
	lasts, every time.Duration
	lock         sync.Mutex
	byOwner      map[string]link
}

// link is a link sent: the SHA-256 of its key, the email it was sent to, as its owner's row has it, and when
type link struct {
	keyHash  [sha256.Size]byte
	emailEnc string
	sentAt   time.Time
}

// NewLinks are links that each work for lasts, and are sent to an owner at most once each every
func NewLinks(lasts, every time.Duration) *Links {
	return &Links{lasts: lasts, every: every, byOwner: map[string]link{}}
}

// Wait is how long until owner may be sent another link, 0 for now
func (l *Links) Wait(owner string) time.Duration {
	l.lock.Lock()
	defer l.lock.Unlock()
	last, sent := l.byOwner[owner]
	if !sent {
		return 0
	}
	return max(0, l.every-time.Since(last.sentAt))
}

// Claim takes owner's turn to be sent a link, before the email is sent and in one step, so requests at once send one
// email. If wait is not 0, owner is sent none for that long. If it is 0, send the email and keep its link with Sent,
// or call notSent if the email was not sent, which puts back what was there, so another can be sent at once.
func (l *Links) Claim(owner string) (wait time.Duration, notSent func()) {
	l.lock.Lock()
	defer l.lock.Unlock()
	before, sent := l.byOwner[owner]
	if wait = l.every - time.Since(before.sentAt); sent && wait > 0 {
		return wait, nil
	}
	l.byOwner[owner] = link{sentAt: time.Now()}
	return 0, func() {
		l.lock.Lock()
		defer l.lock.Unlock()
		if sent {
			l.byOwner[owner] = before
		} else {
			delete(l.byOwner, owner)
		}
	}
}

// Key is a new link's key, as it goes in the link
func Key() string {
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	return hex.EncodeToString(key)
}

// Sent keeps the link with key, sent for owner to the email emailEnc, in place of the one before
func (l *Links) Sent(owner, key, emailEnc string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.byOwner[owner] = link{keyHash: sha256.Sum256([]byte(key)), emailEnc: emailEnc, sentAt: time.Now()}
}

// Forget stops owner's link working, though when it was sent still counts, so another is not sent at once
func (l *Links) Forget(owner string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if last, ok := l.byOwner[owner]; ok {
		l.byOwner[owner] = link{sentAt: last.sentAt}
	}
}

// Take is the email owner's link was sent to, if key is its key and it still works, which it then no longer does
func (l *Links) Take(owner, key string) (emailEnc string, ok bool) {
	keyHash := sha256.Sum256([]byte(key))
	l.lock.Lock()
	defer l.lock.Unlock()
	last, sent := l.byOwner[owner]
	if !sent || last.emailEnc == "" || time.Since(last.sentAt) >= l.lasts || subtle.ConstantTimeCompare(keyHash[:], last.keyHash[:]) != 1 {
		return "", false
	}
	l.byOwner[owner] = link{sentAt: last.sentAt}
	return last.emailEnc, true
}
