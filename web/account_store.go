package web

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	errUserExists         = errors.New("用户名已存在")
	errUserNotFound       = errors.New("用户不存在")
	errBadCredentials     = errors.New("用户名或密码错误")
	errUserDisabled       = errors.New("用户已禁用")
	errLicenseUnavailable = errors.New("卡密不存在或已禁用")
	errLicenseExhausted   = errors.New("卡密次数已用完")
)

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Salt         string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Disabled     bool      `json:"disabled"`
	RunCount     int       `json:"runCount"`
}

type UserSummary struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Disabled  bool      `json:"disabled"`
	RunCount  int       `json:"runCount"`
}

type LicenseKey struct {
	Key       string     `json:"key"`
	Note      string     `json:"note"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"createdAt"`
	UsedBy    string     `json:"usedBy,omitempty"`
	UsedAt    *time.Time `json:"usedAt,omitempty"`
	Uses      int        `json:"uses"`
	MaxUses   int        `json:"maxUses"`
	Remaining int        `json:"remaining"`
}

type AppStats struct {
	UserCount    int `json:"userCount"`
	RunCount     int `json:"runCount"`
	TotalTasks   int `json:"totalTasks"`
	RunningTasks int `json:"runningTasks"`
	TotalLogs    int `json:"totalLogs"`
}

type accountStore struct {
	mu          sync.Mutex
	users       map[string]*User
	userByName  map[string]*User
	sessions    map[string]string
	licenses    map[string]*LicenseKey
	userCounter int
	totalRuns   int
}

func newAccountStore() *accountStore {
	store := &accountStore{
		users:      make(map[string]*User),
		userByName: make(map[string]*User),
		sessions:   make(map[string]string),
		licenses:   make(map[string]*LicenseKey),
	}
	store.addLicenseLocked(LicenseKey{Key: "YATORI-DEMO", Note: "默认测试卡密", Active: true, MaxUses: 0})
	return store
}

func (s *accountStore) register(username, password string) (UserSummary, error) {
	username = strings.TrimSpace(username)
	if username == "" || strings.TrimSpace(password) == "" {
		return UserSummary{}, errors.New("用户名和密码不能为空")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.userByName[username]; ok {
		return UserSummary{}, errUserExists
	}
	now := time.Now()
	s.userCounter++
	salt := randomToken()
	user := &User{
		ID:           fmt.Sprintf("user-%d", s.userCounter),
		Username:     username,
		PasswordHash: hashPassword(password, salt),
		Salt:         salt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.users[user.ID] = user
	s.userByName[user.Username] = user
	return summarizeUser(*user), nil
}

func (s *accountStore) login(username, password string) (string, UserSummary, error) {
	username = strings.TrimSpace(username)
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.userByName[username]
	if !ok || user.PasswordHash != hashPassword(password, user.Salt) {
		return "", UserSummary{}, errBadCredentials
	}
	if user.Disabled {
		return "", UserSummary{}, errUserDisabled
	}
	token := randomToken()
	s.sessions[token] = user.ID
	return token, summarizeUser(*user), nil
}

func (s *accountStore) logout(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

func (s *accountStore) userBySession(token string) (UserSummary, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.sessions[token]
	if !ok {
		return UserSummary{}, false
	}
	user, ok := s.users[id]
	if !ok || user.Disabled {
		return UserSummary{}, false
	}
	return summarizeUser(*user), true
}

func (s *accountStore) listUsers() []UserSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]UserSummary, 0, len(s.users))
	for _, user := range s.users {
		out = append(out, summarizeUser(*user))
	}
	return out
}

func (s *accountStore) upsertUser(id, username, password string, disabled *bool, createdAt *time.Time) (UserSummary, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return UserSummary{}, errors.New("用户名不能为空")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if id == "" {
		if _, ok := s.userByName[username]; ok {
			return UserSummary{}, errUserExists
		}
		s.userCounter++
		salt := randomToken()
		user := &User{
			ID:           fmt.Sprintf("user-%d", s.userCounter),
			Username:     username,
			PasswordHash: hashPassword(password, salt),
			Salt:         salt,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if user.PasswordHash == hashPassword("", salt) {
			return UserSummary{}, errors.New("密码不能为空")
		}
		if disabled != nil {
			user.Disabled = *disabled
		}
		if createdAt != nil {
			user.CreatedAt = *createdAt
		}
		s.users[user.ID] = user
		s.userByName[user.Username] = user
		return summarizeUser(*user), nil
	}
	user, ok := s.users[id]
	if !ok {
		return UserSummary{}, errUserNotFound
	}
	if existing, ok := s.userByName[username]; ok && existing.ID != id {
		return UserSummary{}, errUserExists
	}
	delete(s.userByName, user.Username)
	user.Username = username
	s.userByName[user.Username] = user
	if strings.TrimSpace(password) != "" {
		user.Salt = randomToken()
		user.PasswordHash = hashPassword(password, user.Salt)
	}
	if disabled != nil {
		user.Disabled = *disabled
	}
	if createdAt != nil {
		user.CreatedAt = *createdAt
	}
	user.UpdatedAt = now
	return summarizeUser(*user), nil
}

func (s *accountStore) deleteUser(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return false
	}
	delete(s.users, id)
	delete(s.userByName, user.Username)
	for token, userID := range s.sessions {
		if userID == id {
			delete(s.sessions, token)
		}
	}
	return true
}

func (s *accountStore) addLicense(key LicenseKey) (LicenseKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addLicenseLocked(key)
}

func (s *accountStore) addLicenseLocked(key LicenseKey) (LicenseKey, error) {
	key.Key = strings.TrimSpace(key.Key)
	if key.Key == "" {
		key.Key = randomReadableKey("YATORI")
	}
	if _, ok := s.licenses[key.Key]; ok {
		return LicenseKey{}, errors.New("卡密已存在")
	}
	if key.CreatedAt.IsZero() {
		key.CreatedAt = time.Now()
	}
	if key.MaxUses < 0 {
		key.MaxUses = 0
	}
	item := key
	s.licenses[item.Key] = &item
	return item, nil
}

func (s *accountStore) listLicenses() []LicenseKey {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]LicenseKey, 0, len(s.licenses))
	for _, item := range s.licenses {
		out = append(out, summarizeLicense(*item))
	}
	return out
}

func (s *accountStore) verifyLicense(userID, key string) (LicenseKey, error) {
	key = strings.TrimSpace(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.licenses[key]
	if !ok || !item.Active {
		return LicenseKey{}, errLicenseUnavailable
	}
	if item.MaxUses > 0 && item.Uses >= item.MaxUses {
		return LicenseKey{}, errLicenseExhausted
	}
	return summarizeLicense(*item), nil
}

func (s *accountStore) consumeLicense(userID, key string) (LicenseKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.licenses[strings.TrimSpace(key)]
	if !ok || !item.Active {
		return LicenseKey{}, errLicenseUnavailable
	}
	if item.MaxUses > 0 && item.Uses >= item.MaxUses {
		return LicenseKey{}, errLicenseExhausted
	}
	now := time.Now()
	item.UsedBy = userID
	item.UsedAt = &now
	item.Uses++
	return summarizeLicense(*item), nil
}

func (s *accountStore) deleteLicense(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.licenses[key]; !ok {
		return false
	}
	delete(s.licenses, key)
	return true
}

func (s *accountStore) recordRun(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalRuns++
	if user, ok := s.users[userID]; ok {
		user.RunCount++
		user.UpdatedAt = time.Now()
	}
}

func (s *accountStore) totalRunCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.totalRuns
}

func summarizeUser(user User) UserSummary {
	return UserSummary{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Disabled:  user.Disabled,
		RunCount:  user.RunCount,
	}
}

func summarizeLicense(key LicenseKey) LicenseKey {
	if key.MaxUses > 0 {
		key.Remaining = key.MaxUses - key.Uses
		if key.Remaining < 0 {
			key.Remaining = 0
		}
	} else {
		key.Remaining = -1
	}
	return key
}

func hashPassword(password, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + password))
	return base64.RawStdEncoding.EncodeToString(sum[:])
}

func randomToken() string {
	var data [24]byte
	if _, err := rand.Read(data[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(data[:])
}

func randomReadableKey(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "YATORI"
	}
	prefix = strings.ToUpper(prefix)
	prefix = strings.NewReplacer(" ", "", "-", "", "_", "").Replace(prefix)
	if prefix == "" {
		prefix = "YATORI"
	}
	token := strings.ToUpper(randomToken())
	token = strings.NewReplacer("-", "", "_", "").Replace(token)
	if len(token) > 20 {
		token = token[:20]
	}
	return prefix + "-" + token
}
