package http

import (
	"arch-homework/pkg/common/app/uuid"
	"arch-homework/pkg/common/infrastructure/metrics"
	"arch-homework/pkg/common/jwtauth"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"arch-homework/pkg/order/app"
)

const PathPrefix = "/api/v1/"

const (
	createOrderEndpoint  = PathPrefix + "order"
	ordersEndpoint       = PathPrefix + "orders"
	specificOderEndpoint = PathPrefix + "order/{id}"
)

const (
	errorCodeUnknown       = 0
	errorCodeOrderNotFound = 1
	errorCodePaymentFailed = 1
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
		r, _ := regexp.Compile("^" + PathPrefix + "order/[a-f0-9-]+$")
		if r.MatchString(uri) {
			return specificOderEndpoint
		}
	}
	return uri
}

func NewServer(orderService *app.OrderService, tokenParser jwtauth.TokenParser, logger *logrus.Logger) *Server {
	return &Server{
		orderService: orderService,
		tokenParser:  tokenParser,
		logger:       logger,
	}
}

type Server struct {
	orderService *app.OrderService
	tokenParser  jwtauth.TokenParser
	logger       *logrus.Logger
}

func (s *Server) MakeHandler() http.Handler {
	router := mux.NewRouter()

	router.Methods(http.MethodPost).Path(createOrderEndpoint).Handler(s.makeHandlerFunc(s.createOrderHandler))
	router.Methods(http.MethodGet).Path(specificOderEndpoint).Handler(s.makeHandlerFunc(s.getOrderHandler))
	router.Methods(http.MethodGet).Path(ordersEndpoint).Handler(s.makeHandlerFunc(s.getOrdersHandler))

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

func (s *Server) getOrderHandler(w http.ResponseWriter, r *http.Request) error {
	tokenData, err := s.extractAuthorizationData(r)
	if err != nil {
		return err
	}
	orderID, err := getIDFromRequest(r)
	if err != nil {
		return err
	}

	order, err := s.orderService.Get(app.UserID(tokenData.UserID()), orderID)
	if err != nil {
		return err
	}
	writeResponse(w, toOrderInfo(*order))
	return nil
}

func (s *Server) getOrdersHandler(w http.ResponseWriter, r *http.Request) error {
	tokenData, err := s.extractAuthorizationData(r)
	if err != nil {
		return err
	}

	orders, err := s.orderService.FindAll(app.UserID(tokenData.UserID()))
	if err != nil {
		return err
	}
	orderInfos := make([]orderInfo, 0, len(orders))
	for _, order := range orders {
		orderInfos = append(orderInfos, toOrderInfo(order))
	}
	writeResponse(w, orderInfos)
	return nil
}

func (s *Server) createOrderHandler(w http.ResponseWriter, r *http.Request) error {
	tokenData, err := s.extractAuthorizationData(r)
	if err != nil {
		return err
	}

	var info createOrderInfo
	bytesBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytesBody, &info); err != nil {
		return err
	}

	orderID, err := s.orderService.Create(app.UserID(tokenData.UserID()), info.Price)
	if err != nil {
		return err
	}
	response := createOrderResponse{ID: string(orderID)}
	writeResponse(w, response)
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
	if err = uuid.ValidateUUID(tokenData.UserID()); err != nil {
		return nil, errors.WithStack(err)
	}
	return tokenData, nil
}

func getIDFromRequest(r *http.Request) (app.OrderID, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return "", errors.New("id param required")
	}
	if err := uuid.ValidateUUID(id); err != nil {
		return "", err
	}
	return app.OrderID(id), nil
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
	case app.ErrOrderNotFound:
		info.Code = errorCodeOrderNotFound
		w.WriteHeader(http.StatusNotFound)
	case app.ErrPaymentFailed:
		info.Code = errorCodePaymentFailed
		w.WriteHeader(http.StatusBadRequest)
	case errForbidden:
		w.WriteHeader(http.StatusForbidden)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	js, _ := json.Marshal(info)
	_, _ = w.Write(js)
}

func toOrderInfo(order app.Order) orderInfo {
	return orderInfo{
		ID:           string(order.ID),
		Price:        order.Price.Value(),
		CreationDate: order.CreationDate.Format(time.RFC3339),
	}
}

type errorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type orderInfo struct {
	ID           string  `json:"id"`
	Price        float64 `json:"price"`
	CreationDate string  `json:"creationDate"`
}

type createOrderInfo struct {
	Price float64 `json:"price"`
}

type createOrderResponse struct {
	ID string `json:"id"`
}
