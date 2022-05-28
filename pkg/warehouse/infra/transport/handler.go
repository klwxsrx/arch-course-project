package transport

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/transport"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/domain"
	"net/http"
)

type route struct {
	Name    string
	Method  string
	Pattern string
	Handler func(*service.WarehouseService, http.ResponseWriter, *http.Request)
}

func getRoutes() []route {
	return []route{
		{
			"getAvailableItemsQuantity",
			http.MethodGet,
			"/warehouse/items/available",
			getAvailableItemsQuantityHandler,
		},
		{
			"addItems",
			http.MethodPut,
			"/warehouse/items",
			addItemsHandler,
		},
		{
			"reserveOrderItems",
			http.MethodPost,
			"/warehouse/order/{orderID}/reserve",
			reserveOrderItemsHandler,
		},
		{
			"deleteOrderItemsReservation",
			http.MethodDelete,
			"/warehouse/order/{orderID}/reserve",
			deleteOrderItemsReservationHandler,
		},
		{
			"health",
			http.MethodGet,
			"/healthz",
			healthCheckHandler,
		},
	}
}

func getAvailableItemsQuantityHandler(srv *service.WarehouseService, w http.ResponseWriter, r *http.Request) {
	var itemIDs []uuid.UUID
	err := json.NewDecoder(r.Body).Decode(&itemIDs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	itemsQuantity, err := srv.GetAvailableItemsQuantity(itemIDs)
	if errors.Is(err, service.ErrItemNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type resultQuantity struct {
		ItemID   uuid.UUID `json:"item_id"`
		Quantity int       `json:"quantity"`
	}

	resultItems := make([]resultQuantity, 0, len(itemsQuantity))
	for _, item := range itemsQuantity {
		resultItems = append(resultItems, resultQuantity{
			ItemID:   item.ItemID,
			Quantity: item.Quantity,
		})
	}

	err = json.NewEncoder(w).Encode(resultItems)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func addItemsHandler(srv *service.WarehouseService, w http.ResponseWriter, r *http.Request) {
	idempotenceKey, err := parseIdempotenceKey(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body := struct {
		ItemID   uuid.UUID `json:"item_id"`
		Quantity int       `json:"quantity"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = srv.AddItems(idempotenceKey, body.ItemID, body.Quantity)
	switch err {
	case service.ErrInvalidQuantity:
		w.WriteHeader(http.StatusBadRequest)
	case service.ErrItemAlreadyAdded:
		w.WriteHeader(http.StatusConflict)
	case nil:
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func reserveOrderItemsHandler(srv *service.WarehouseService, w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUID(mux.Vars(r)["orderID"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var body []struct {
		ItemID   uuid.UUID `json:"item_id"`
		Quantity int       `json:"quantity"`
	}

	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	items := make([]domain.ItemQuantity, 0, len(body))
	for _, bodyItem := range body {
		items = append(items, domain.ItemQuantity{
			ItemID:   bodyItem.ItemID,
			Quantity: bodyItem.Quantity,
		})
	}

	err = srv.ReserveOrderItems(orderID, items)
	switch err {
	case service.ErrInvalidQuantity:
		w.WriteHeader(http.StatusBadRequest)
	case service.ErrOrderOperationsAlreadyExist:
		w.WriteHeader(http.StatusConflict)
	case service.ErrItemsOutOfStock:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case service.ErrItemNotFound:
		w.WriteHeader(http.StatusNotFound)
	case nil:
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func deleteOrderItemsReservationHandler(srv *service.WarehouseService, w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUID(mux.Vars(r)["orderID"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = srv.DeleteOrderItemsReservation(orderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func healthCheckHandler(_ *service.WarehouseService, w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{"OK"})
}

func parseIdempotenceKey(r *http.Request) (string, error) {
	key := r.Header.Get("X-Idempotence-Key")
	if key == "" {
		return "", errors.New("idempotence key not found")
	}
	return key, nil
}

func parseUUID(str string) (uuid.UUID, error) {
	return uuid.Parse(str)
}

func getHandlerFunc(
	warehouseService *service.WarehouseService,
	f func(*service.WarehouseService, http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(warehouseService, w, r)
	}
}

func NewHTTPHandler(warehouseService *service.WarehouseService, logger log.Logger) (http.Handler, error) {
	router := mux.NewRouter()

	for _, route := range getRoutes() {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			HandlerFunc(getHandlerFunc(warehouseService, route.Handler))
	}

	router.Use(transport.NewLoggingMiddleware(logger))
	return router, nil
}
