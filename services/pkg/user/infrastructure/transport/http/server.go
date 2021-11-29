package http

import (
	"arch-homework5/pkg/common/jwtauth"
	"arch-homework5/pkg/common/metrics"
	"arch-homework5/pkg/common/uuid"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"arch-homework5/pkg/user/app"
)

const PathPrefix = "/api/v1/"

const (
	currentUserEndpoint  = PathPrefix + "user"
	specificUserEndpoint = PathPrefix + "user/{id}"
)

const (
	errorCodeUnknown      = 0
	errorCodeUserNotFound = 1
)

const authTokenHeader = "X-Auth-Token"

var errForbidden = errors.New("access forbidden")

func NewEndpointLabelCollector() metrics.EndpointLabelCollector {
	return endpointLabelCollector{}
}

type endpointLabelCollector struct {
}

func (e endpointLabelCollector) EndpointLabelForURI(uri string) string {
	if strings.HasPrefix(uri, PathPrefix) {
		r, _ := regexp.Compile("^" + PathPrefix + "user/[a-f0-9-]+$")
		if r.MatchString(uri) {
			return specificUserEndpoint
		}
	}
	return uri
}

func NewServer(userService *app.UserService, tokenParser jwtauth.TokenParser, logger *logrus.Logger) *Server {
	return &Server{
		userService: userService,
		tokenParser: tokenParser,
		logger:      logger,
	}
}

type Server struct {
	userService *app.UserService
	tokenParser jwtauth.TokenParser
	logger      *logrus.Logger
}

func (s *Server) MakeHandler() http.Handler {
	router := mux.NewRouter()

	router.Methods(http.MethodGet).Path(currentUserEndpoint).Handler(s.makeHandlerFunc(s.getCurrentUserIDHandler))
	router.Methods(http.MethodGet).Path(specificUserEndpoint).Handler(s.makeHandlerFunc(s.getUserHandler))
	router.Methods(http.MethodPut).Path(specificUserEndpoint).Handler(s.makeHandlerFunc(s.updateUserHandler))

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

func (s *Server) getCurrentUserIDHandler(w http.ResponseWriter, r *http.Request) error {
	tokenData, err := s.extractAuthorizationData(r)
	if err != nil {
		return err
	}

	writeResponse(w, currentUserInfo{UserID: tokenData.UserID()})
	return nil
}

func (s *Server) getUserHandler(w http.ResponseWriter, r *http.Request) error {
	tokenData, err := s.extractAuthorizationData(r)
	if err != nil {
		return err
	}
	id, err := getIDFromRequest(r)
	if err != nil {
		return err
	}
	if string(id) != tokenData.UserID() {
		return errForbidden
	}

	user, err := s.userService.Get(id)
	if err != nil {
		return err
	}
	writeResponse(w, toUserInfo(*user, tokenData))
	return nil
}

func (s *Server) updateUserHandler(w http.ResponseWriter, r *http.Request) error {
	tokenData, err := s.extractAuthorizationData(r)
	if err != nil {
		return err
	}
	id, err := getIDFromRequest(r)
	if err != nil {
		return err
	}
	if string(id) != tokenData.UserID() {
		return errForbidden
	}

	var info userInfoUpdate
	bytesBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytesBody, &info); err != nil {
		return err
	}
	var email *app.Email
	var phone *app.Phone

	if info.Email != nil {
		emailValue := app.Email(*info.Email)
		email = &emailValue
	}
	if info.Phone != nil {
		phoneValue := app.Phone(*info.Phone)
		phone = &phoneValue
	}

	err = s.userService.Update(id, info.FirstName, info.LastName, email, phone)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, http.StatusText(http.StatusOK))
	return nil
}

func (s *Server) extractAuthorizationData(r *http.Request) (jwtauth.TokenData, error) {
	token := r.Header.Get(authTokenHeader)
	if token == "" {
		return nil, errForbidden
	}
	tokenData, err := s.tokenParser.ParseToken(token)
	if err != nil {
		return nil, errors.Wrap(errForbidden, err.Error())
	}
	return tokenData, nil
}

func getIDFromRequest(r *http.Request) (app.UserID, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return "", errors.New("id param required")
	}
	if err := uuid.ValidateUUID(id); err != nil {
		return "", err
	}
	return app.UserID(id), nil
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
	case errForbidden:
		w.WriteHeader(http.StatusForbidden)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	js, _ := json.Marshal(info)
	_, _ = w.Write(js)
}

func toUserInfo(user app.User, tokenData jwtauth.TokenData) userInfo {
	return userInfo{
		UserID:    string(user.UserID),
		Login:     tokenData.UserLogin(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     string(user.Email),
		Phone:     string(user.Phone),
	}
}

type errorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type userInfo struct {
	UserID    string `json:"id"`
	Login     string `json:"login"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type userInfoUpdate struct {
	UserID    *string `json:"id"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
}

type currentUserInfo struct {
	UserID string `json:"id"`
}
