package transport

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/transport"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/query"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/order/domain"
	"net/http"
)

const healthEndpoint = "/healthz"

type createOrderItemData struct {
	ID        uuid.UUID `json:"id"`
	ItemPrice int       `json:"item_price"`
	Quantity  int       `json:"quantity"`
}

type createOrderData struct {
	UserID    uuid.UUID             `json:"user_id"`
	AddressID uuid.UUID             `json:"address_id"`
	Items     []createOrderItemData `json:"items"`
}

type route struct {
	Name    string
	Method  string
	Pattern string
	Handler func(*service.OrderService, query.Service, http.ResponseWriter, *http.Request)
}

func getRoutes() []route {
	return []route{
		{
			"createOrder",
			http.MethodPut,
			"/orders",
			createOrderHandler,
		},
		{
			"getOrder",
			http.MethodGet,
			"/web/order/{orderID}",
			getOrderHandler,
		},
		{
			"health",
			http.MethodGet,
			healthEndpoint,
			healthCheckHandler,
		},
	}
}

func createOrderHandler(srv *service.OrderService, _ query.Service, w http.ResponseWriter, r *http.Request) {
	idempotenceKey, err := parseIdempotenceKey(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var createOrder createOrderData
	err = json.NewDecoder(r.Body).Decode(&createOrder)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderItems := make([]domain.OrderItem, 0, len(createOrder.Items))
	for _, item := range createOrder.Items {
		orderItems = append(orderItems, domain.OrderItem{
			ID:        item.ID,
			ItemPrice: item.ItemPrice,
			Quantity:  item.Quantity,
		})
	}

	orderID, err := srv.Create(idempotenceKey, createOrder.UserID, createOrder.AddressID, orderItems)
	if errors.Is(err, service.ErrOrderAlreadyCreated) {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if errors.Is(err, service.ErrEmptyOrder) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(orderID)
}

func getOrderHandler(_ *service.OrderService, qs query.Service, w http.ResponseWriter, r *http.Request) {
	authUserID, err := parseAuthUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orderID, err := parseUUID(mux.Vars(r)["orderID"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order, err := qs.GetOrderData(orderID)
	if errors.Is(err, query.ErrOrderNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if order.UserID != authUserID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type orderItemJSONSchema struct {
		ID        uuid.UUID `json:"id"`
		ItemPrice int       `json:"price"`
		Quantity  int       `json:"quantity"`
	}
	type orderJSONSchema struct {
		ID          uuid.UUID             `json:"id"`
		UserID      uuid.UUID             `json:"user_id"`
		AddressID   uuid.UUID             `json:"address_id"`
		Items       []orderItemJSONSchema `json:"items"`
		Status      string                `json:"status"`
		TotalAmount int                   `json:"total_amount"`
	}

	var orderStatus string
	switch order.Status {
	case domain.OrderStatusCreated, domain.OrderStatusPaymentAuthorized, domain.OrderStatusItemsReserved, domain.OrderStatusDeliveryScheduled:
		orderStatus = "processing"
	case domain.OrderStatusSentToDelivery:
		orderStatus = "sent_to_delivery"
	case domain.OrderStatusDelivered:
		orderStatus = "delivered"
	case domain.OrderStatusCancelled:
		orderStatus = "cancelled"
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	orderItems := make([]orderItemJSONSchema, 0, len(order.Items))
	for _, item := range order.Items {
		orderItems = append(orderItems, orderItemJSONSchema{
			ID:        item.ID,
			ItemPrice: item.ItemPrice,
			Quantity:  item.Quantity,
		})
	}

	err = json.NewEncoder(w).Encode(orderJSONSchema{
		ID:          order.ID,
		UserID:      order.UserID,
		AddressID:   order.AddressID,
		Items:       orderItems,
		Status:      orderStatus,
		TotalAmount: order.TotalAmount,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func healthCheckHandler(_ *service.OrderService, _ query.Service, w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{"OK"})
}

func parseAuthUserID(r *http.Request) (uuid.UUID, error) {
	id := r.Header.Get("X-Auth-User-ID")
	return uuid.Parse(id)
}

func parseUUID(str string) (uuid.UUID, error) {
	return uuid.Parse(str)
}

func parseIdempotenceKey(r *http.Request) (string, error) {
	key := r.Header.Get("X-Idempotence-Key")
	if key == "" {
		return "", errors.New("idempotence key not found")
	}
	return key, nil
}

func getHandlerFunc(
	orderService *service.OrderService,
	queryService query.Service,
	f func(*service.OrderService, query.Service, http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(orderService, queryService, w, r)
	}
}

func NewHTTPHandler(orderService *service.OrderService, queryService query.Service, logger log.Logger) (http.Handler, error) {
	router := mux.NewRouter()

	for _, route := range getRoutes() {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			HandlerFunc(getHandlerFunc(orderService, queryService, route.Handler))
	}

	router.Use(transport.NewLoggingMiddleware(logger, []string{healthEndpoint}))
	return router, nil
}
