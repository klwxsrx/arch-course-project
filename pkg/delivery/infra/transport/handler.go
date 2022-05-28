package transport

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/transport"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/app/query"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/domain"
	"net/http"
)

type route struct {
	Name    string
	Method  string
	Pattern string
	Handler func(*service.DeliveryService, query.Service, http.ResponseWriter, *http.Request)
}

func getRoutes() []route {
	return []route{
		{
			"getDelivery",
			http.MethodGet,
			"/delivery/{orderID}",
			getDeliveryHandler,
		},
		{
			"scheduleDelivery",
			http.MethodPost,
			"/delivery/{orderID}/schedule",
			scheduleDeliveryHandler,
		},
		{
			"cancelDeliverySchedule",
			http.MethodDelete,
			"/delivery/{orderID}/schedule",
			cancelDeliveryScheduleHandler,
		},
		{
			"health",
			http.MethodGet,
			"/healthz",
			healthCheckHandler,
		},
	}
}

func getDeliveryHandler(_ *service.DeliveryService, qs query.Service, w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUID(mux.Vars(r)["orderID"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	del, err := qs.GetByID(orderID)
	if errors.Is(err, query.ErrDeliveryNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var status string
	switch del.Status {
	case domain.DeliveryStatusScheduled:
		status = "scheduled"
	case domain.DeliveryStatusAwaitingDelivery:
		status = "awaiting_delivery"
	case domain.DeliveryStatusProcessing:
		status = "processing"
	case domain.DeliveryStatusDelivered:
		status = "delivered"
	case domain.DeliveryStatusCancelled:
		status = "cancelled"
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(struct {
		OrderID uuid.UUID `json:"order_id"`
		Status  string    `json:"status"`
		Address string    `json:"address"`
	}{
		del.OrderID,
		status,
		del.Address,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func scheduleDeliveryHandler(srv *service.DeliveryService, _ query.Service, w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUID(mux.Vars(r)["orderID"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	addressBody := struct {
		AddressID uuid.UUID `json:"address_id"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&addressBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = srv.Schedule(orderID, addressBody.AddressID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func cancelDeliveryScheduleHandler(srv *service.DeliveryService, _ query.Service, w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUID(mux.Vars(r)["orderID"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = srv.CancelSchedule(orderID)
	if errors.Is(err, service.ErrDeliveryIsNotScheduled) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func healthCheckHandler(_ *service.DeliveryService, _ query.Service, w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{"OK"})
}

func parseUUID(str string) (uuid.UUID, error) {
	return uuid.Parse(str)
}

func getHandlerFunc(
	service *service.DeliveryService,
	query query.Service,
	f func(*service.DeliveryService, query.Service, http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(service, query, w, r)
	}
}

func NewHTTPHandler(service *service.DeliveryService, query query.Service, logger log.Logger) (http.Handler, error) {
	router := mux.NewRouter()

	for _, route := range getRoutes() {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			HandlerFunc(getHandlerFunc(service, query, route.Handler))
	}

	router.Use(transport.NewLoggingMiddleware(logger))
	return router, nil
}
