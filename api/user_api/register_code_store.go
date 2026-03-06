package user_api

import (
	"strings"
	"sync"
	"time"
)

type registerCodeEntry struct {
	Code     string
	ExpireAt time.Time
	SentAt   time.Time
}

var registerCodeStore sync.Map

func normalizeRegisterEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func getRegisterCodeEntry(email string) (registerCodeEntry, bool) {
	normalized := normalizeRegisterEmail(email)
	if normalized == "" {
		return registerCodeEntry{}, false
	}
	raw, ok := registerCodeStore.Load(normalized)
	if !ok {
		return registerCodeEntry{}, false
	}
	entry, ok := raw.(registerCodeEntry)
	if !ok {
		registerCodeStore.Delete(normalized)
		return registerCodeEntry{}, false
	}
	if time.Now().After(entry.ExpireAt) {
		registerCodeStore.Delete(normalized)
		return registerCodeEntry{}, false
	}
	return entry, true
}

func setRegisterCodeEntry(email string, code string, ttl time.Duration) registerCodeEntry {
	entry := registerCodeEntry{
		Code:     code,
		ExpireAt: time.Now().Add(ttl),
		SentAt:   time.Now(),
	}
	registerCodeStore.Store(normalizeRegisterEmail(email), entry)
	return entry
}

func clearRegisterCodeEntry(email string) {
	registerCodeStore.Delete(normalizeRegisterEmail(email))
}
