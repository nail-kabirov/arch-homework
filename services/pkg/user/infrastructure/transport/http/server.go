package http

import (
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
	"arch-homework5/pkg/user/infrastructure/metrics"
)

const PathPrefix = "/api/v1/"

const (
	createUserEndpoint   = PathPrefix + "user"
	getUsersEndpoint     = PathPrefix + "users"
	specificUserEndpoint = PathPrefix + "user/{id}"
)

const (
	errorCodeUnknown               = 0
	errorCodeUserNotFound          = 1
	errorCodeUsernameAlreadyExists = 2
	errorCodeUsernameTooLong       = 3
)

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

func NewServer(userService *app.UserService, logger *logrus.Logger) *Server {
	return &Server{
		userService: userService,
		logger:      logger,
	}
}

type Server struct {
	userService *app.UserService
	logger      *logrus.Logger
}

func (s *Server) MakeHandler() http.Handler {
	router := mux.NewRouter()

	router.Methods(http.MethodPost).Path(createUserEndpoint).Handler(s.makeHandlerFunc(s.createUserHandler))
	router.Methods(http.MethodGet).Path(getUsersEndpoint).Handler(s.makeHandlerFunc(s.getUsersHandler))
	router.Methods(http.MethodGet).Path(specificUserEndpoint).Handler(s.makeHandlerFunc(s.getUserHandler))
	router.Methods(http.MethodPut).Path(specificUserEndpoint).Handler(s.makeHandlerFunc(s.updateUserHandler))
	router.Methods(http.MethodDelete).Path(specificUserEndpoint).Handler(s.makeHandlerFunc(s.deleteUserHandler))

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

func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) error {
	var info userInfo
	bytesBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytesBody, &info); err != nil {
		return err
	}

	userID, err := s.userService.Add(app.Username(info.Username), info.FirstName, info.LastName, app.Email(info.Email), app.Phone(info.Phone))
	if err != nil {
		return err
	}
	writeResponse(w, createdUserInfo{UserID: string(userID)})
	return nil
}

func (s *Server) getUsersHandler(w http.ResponseWriter, _ *http.Request) error {
	users, err := s.userService.FindAll()
	if err != nil {
		return err
	}

	userInfos := make([]userInfo, 0, len(users))
	for _, user := range users {
		userInfos = append(userInfos, toUserInfo(user))
	}

	writeResponse(w, userInfos)
	return nil
}

func (s *Server) getUserHandler(w http.ResponseWriter, r *http.Request) error {
	id, err := getIDFromRequest(r)
	if err != nil {
		return err
	}
	user, err := s.userService.Find(id)
	if err != nil {
		return err
	}
	writeResponse(w, toUserInfo(*user))
	return nil
}

func (s *Server) updateUserHandler(w http.ResponseWriter, r *http.Request) error {
	id, err := getIDFromRequest(r)
	if err != nil {
		return err
	}

	var info userInfoUpdate
	bytesBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytesBody, &info); err != nil {
		return err
	}
	var username *app.Username
	var email *app.Email
	var phone *app.Phone
	if info.Username != nil {
		usernameValue := app.Username(*info.Username)
		username = &usernameValue
	}
	if info.Email != nil {
		emailValue := app.Email(*info.Email)
		email = &emailValue
	}
	if info.Phone != nil {
		phoneValue := app.Phone(*info.Phone)
		phone = &phoneValue
	}

	err = s.userService.Update(id, username, info.FirstName, info.LastName, email, phone)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, http.StatusText(http.StatusOK))
	return nil
}

func (s *Server) deleteUserHandler(w http.ResponseWriter, r *http.Request) error {
	id, err := getIDFromRequest(r)
	if err != nil {
		return err
	}
	err = s.userService.Remove(id)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
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
	case app.ErrUsernameAlreadyExists:
		info.Code = errorCodeUsernameAlreadyExists
		w.WriteHeader(http.StatusBadRequest)
	case app.ErrUsernameTooLong:
		info.Code = errorCodeUsernameTooLong
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	js, _ := json.Marshal(info)
	_, _ = w.Write(js)
}

func toUserInfo(user app.User) userInfo {
	return userInfo{
		UserID:    string(user.UserID),
		Username:  string(user.Username),
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
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type userInfoUpdate struct {
	UserID    *string `json:"id"`
	Username  *string `json:"username"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
}

type createdUserInfo struct {
	UserID string `json:"id"`
}
