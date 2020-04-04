package dcron

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/docker/distribution/uuid"
)

// AuthBackend Authentication server layer
type AuthBackend interface {
	HandleLogin(w http.ResponseWriter, r *http.Request)
	AuthMiddleware(next http.Handler) http.Handler
}

type session struct {
	Token   string
	Expires time.Time
}

func (s session) isExpired() bool {
	return s.Expires.Unix() < time.Now().Unix()
}

type sessionStore struct {
	sync.RWMutex
	Sessions map[string]session
}

func (s *sessionStore) Add(session session) {
	s.Lock()
	defer s.Unlock()
	// clean expired sessions first
	for key, session := range s.Sessions {
		if session.isExpired() {
			delete(s.Sessions, key)
		}
	}
	s.Sessions[session.Token] = session
}

func (s *sessionStore) Remove(token string) {
	s.Lock()
	delete(s.Sessions, token)
	s.Unlock()
}

func (s *sessionStore) Check(token string) bool {
	s.RLock()
	ses, ok := s.Sessions[token]
	s.RUnlock()
	if ok && ses.isExpired() {
		s.Remove(token)
		return false
	}
	return ok
}

type credentials struct {
	Password string `json:"password"`
}

type authBackend struct {
	Credentials credentials
	Sessions    *sessionStore
	Expiration  time.Duration
}

func (a *authBackend) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var u credentials
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	valid := u.Password == a.Credentials.Password
	if !valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	s := session{uuid.Generate().String(), time.Now().Add(a.Expiration)}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    s.Token,
		Expires:  s.Expires,
		Path:     "/",
		HttpOnly: true,
	})
	a.Sessions.Add(s)
	fmt.Fprintf(w, "Success\n")
}

func (a *authBackend) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session")
		if err != nil || c == nil || !a.Sessions.Check(c.Value) {
			http.Error(w, "Permission denied", 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewAuthBackend creates new authentication backend
func NewAuthBackend(password string, expiration time.Duration) AuthBackend {
	store := &sessionStore{Sessions: make(map[string]session)}
	return &authBackend{credentials{password}, store, expiration}
}
