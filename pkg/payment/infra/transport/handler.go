package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/transport"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/query"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/payment/domain"
	"net/http"
)

type route struct {
	Name    string
	Method  string
	Pattern string
	Handler func(*service.PaymentService, query.PaymentQueryService, http.ResponseWriter, *http.Request)
}

func getRoutes() []route {
	return []route{
		{
			"getPayment",
			http.MethodGet,
			"/payment/{orderID}",
			getPaymentHandler,
		},
		{
			"health",
			http.MethodGet,
			"/healthz",
			healthCheckHandler,
		},
	}
}

func getPaymentHandler(_ *service.PaymentService, srv query.PaymentQueryService, w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUID(mux.Vars(r)["orderID"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := srv.GetPayment(orderID)
	if errors.Is(err, query.ErrPaymentNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	textStatus, err := getTextStatus(data.Status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := struct {
		OrderID     uuid.UUID `json:"order_id"`
		Status      string    `json:"status"`
		TotalAmount int       `json:"total_amount"`
	}{
		data.OrderID,
		textStatus,
		data.TotalAmount,
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resultJSON)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func healthCheckHandler(_ *service.PaymentService, _ query.PaymentQueryService, w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{"OK"})
}

func getTextStatus(status domain.PaymentStatus) (string, error) {
	switch status {
	case domain.PaymentStatusAuthorized:
		return "authorized", nil
	case domain.PaymentStatusCancelled:
		return "cancelled", nil
	case domain.PaymentStatusCompleted:
		return "completed", nil
	case domain.PaymentStatusRejected:
		return "rejected", nil
	default:
		return "", errors.New(fmt.Sprintf("unknown status %v", status))
	}
}

func parseUUID(str string) (uuid.UUID, error) {
	return uuid.Parse(str)
}

func getHandlerFunc(
	paymentService *service.PaymentService,
	paymentQueryService query.PaymentQueryService,
	f func(*service.PaymentService, query.PaymentQueryService, http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(paymentService, paymentQueryService, w, r)
	}
}

func NewHTTPHandler(paymentService *service.PaymentService, queryService query.PaymentQueryService, logger log.Logger) (http.Handler, error) {
	router := mux.NewRouter()

	for _, route := range getRoutes() {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			HandlerFunc(getHandlerFunc(paymentService, queryService, route.Handler))
	}

	router.Use(transport.NewLoggingMiddleware(logger))
	return router, nil
}
