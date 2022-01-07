package http

import (
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"arch-homework/pkg/auth/app"
	"arch-homework/pkg/common/app/uuid"
	"arch-homework/pkg/common/jwtauth"
)

const PathPrefix = "/api/v1/"

const (
	registerUserEndpoint = PathPrefix + "register"
	authEndpoint         = PathPrefix + "auth"
	loginEndpoint        = PathPrefix + "login"
	logoutEndpoint       = PathPrefix + "logout"
)

const (
	errorCodeUnknown               = 0
	errorCodeUserNotFound          = 1
	errorCodeUsernameAlreadyExists = 2
	errorCodeUsernameTooLong       = 3
	errorCodeInvalidPassword       = 4
)

const sessionCookieName = "session_id"
const sessionLifetime = time.Minute * 30
const authTokenHeader = "X-Auth-Token"

var errUnauthorized = errors.New("not authorized")

func NewServer(userService *app.UserService, sessionRepo app.SessionRepository, tokenGenerator jwtauth.TokenGenerator, logger *logrus.Logger) *Server {
	return &Server{
		userService:    userService,
		sessionRepo:    sessionRepo,
		tokenGenerator: tokenGenerator,
		logger:         logger,
	}
}

type Server struct {
	userService    *app.UserService
	sessionRepo    app.SessionRepository
	tokenGenerator jwtauth.TokenGenerator
	logger         *logrus.Logger
}

func (s *Server) MakeHandler() http.Handler {
	router := mux.NewRouter()

	router.Methods(http.MethodPost).Path(registerUserEndpoint).Handler(s.makeHandlerFunc(s.registerUserHandler))
	router.Methods(http.MethodPost).Path(loginEndpoint).Handler(s.makeHandlerFunc(s.loginHandler))
	router.Methods(http.MethodPost).Path(logoutEndpoint).Handler(s.makeHandlerFunc(s.logoutHandler))
	router.Path(authEndpoint).Handler(s.makeHandlerFunc(s.authHandler))

	return router
}

func (s *Server) makeHandlerFunc(fn func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		fields := logrus.Fields{
			"method": r.Method,
			"host":   r.Host,
			"path":   r.URL.Path,
		}
		if r.URL.RawQuery != "" {
			fields["query"] = r.URL.RawQuery
		}
		if r.PostForm != nil {
			fields["post"] = r.PostForm
		}

		err := fn(w, r)

		if err != nil {
			writeErrorResponse(w, err)

			fields["err"] = err
			s.logger.WithFields(fields).Error()
		} else {
			s.logger.WithFields(fields).Info("call")
		}
	}
}

func (s *Server) registerUserHandler(w http.ResponseWriter, r *http.Request) error {
	var info userAuthData
	bytesBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytesBody, &info); err != nil {
		return err
	}

	userID, err := s.userService.Add(info.Login, info.Password)
	if err != nil {
		return err
	}
	writeResponse(w, createdUserInfo{UserID: string(userID)})
	return nil
}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) error {
	var info userAuthData
	bytesBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytesBody, &info); err != nil {
		return err
	}

	user, err := s.userService.FindUserByLoginAndPassword(info.Login, info.Password)
	if err != nil {
		return err
	}
	session := app.Session{
		ID:        app.SessionID(uuid.GenerateNew()),
		UserID:    user.UserID,
		ValidTill: time.Now().Add(sessionLifetime),
	}
	err = s.sessionRepo.Store(&session)
	if err != nil {
		return err
	}

	setSessionCookie(w, &session.ID)
	w.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) error {
	if sessionID, err := getSessionIDFromRequest(r); err == nil {
		err = s.sessionRepo.Remove(sessionID)
		if err != nil {
			return err
		}
	}

	setSessionCookie(w, nil)
	w.WriteHeader(http.StatusOK)
	return nil
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) error {
	sessionID, err := getSessionIDFromRequest(r)
	if err != nil {
		return errUnauthorized
	}
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		if errors.Cause(err) == app.ErrSessionNotFound {
			return errUnauthorized
		}
		return err
	}
	user, err := s.userService.FindUserByID(session.UserID)
	if err != nil {
		return err
	}
	session.ValidTill = time.Now().Add(sessionLifetime)
	_ = s.sessionRepo.Store(session)

	token, err := s.tokenGenerator.GenerateToken(string(user.UserID), string(user.Login))
	if err != nil {
		return err
	}

	w.Header().Set(authTokenHeader, token)
	w.WriteHeader(http.StatusOK)
	return nil
}

func getSessionIDFromRequest(r *http.Request) (app.SessionID, error) {
	sessionID, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}
	if err := uuid.ValidateUUID(sessionID.Value); err != nil {
		return "", err
	}
	return app.SessionID(sessionID.Value), nil
}

func setSessionCookie(w http.ResponseWriter, sessionID *app.SessionID) {
	c := &http.Cookie{
		Name:     sessionCookieName,
		Path:     "/",
		HttpOnly: true,
	}
	if sessionID != nil {
		c.Value = string(*sessionID)
	} else {
		// delete cookie
		c.MaxAge = -1
	}

	http.SetCookie(w, c)
}

func writeResponse(w http.ResponseWriter, response interface{}) {
	js, err := json.Marshal(response)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(js)
}

func writeErrorResponse(w http.ResponseWriter, err error) {
	info := errorInfo{Code: errorCodeUnknown, Message: err.Error()}
	switch errors.Cause(err) {
	case app.ErrUserNotFound:
		info.Code = errorCodeUserNotFound
		w.WriteHeader(http.StatusNotFound)
	case app.ErrLoginAlreadyExists:
		info.Code = errorCodeUsernameAlreadyExists
		w.WriteHeader(http.StatusBadRequest)
	case app.ErrLoginTooLong:
		info.Code = errorCodeUsernameTooLong
		w.WriteHeader(http.StatusBadRequest)
	case app.ErrInvalidPassword:
		info.Code = errorCodeInvalidPassword
		w.WriteHeader(http.StatusBadRequest)
	case errUnauthorized:
		w.WriteHeader(http.StatusUnauthorized)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	js, _ := json.Marshal(info)
	_, _ = w.Write(js)
}

type userAuthData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type createdUserInfo struct {
	UserID string `json:"id"`
}

type errorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
