package transport

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/klwxsrx/arch-course-project/pkg/cart/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/transport"
	"net/http"
)

type route struct {
	Name    string
	Method  string
	Pattern string
	Handler func(*service.CartService, http.ResponseWriter, *http.Request)
}

func getRoutes() []route {
	return []route{
		{
			"getCart",
			http.MethodGet,
			"/web/cart",
			getCartHandler,
		},
		{
			"addToCart",
			http.MethodPut,
			"/web/cart",
			addToCartHandler,
		},
		{
			"deleteFromCart",
			http.MethodDelete,
			"/web/cart/{productID}",
			deleteFromCartHandler,
		},
		{
			"checkout",
			http.MethodPost,
			"/web/cart/checkout",
			checkoutHandler,
		},
		{
			"health",
			http.MethodGet,
			"/healthz",
			healthCheckHandler,
		},
	}
}

func getCartHandler(srv *service.CartService, w http.ResponseWriter, r *http.Request) {
	authUserID, err := parseAuthUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	cart, err := srv.GetCart(authUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type productQuantity struct {
		ID       uuid.UUID `json:"id"`
		Quantity int       `json:"quantity"`
	}

	result := make([]productQuantity, 0, len(cart.Products))
	for _, item := range cart.Products {
		result = append(result, productQuantity{
			ID:       item.ID,
			Quantity: item.Quantity,
		})
	}

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func addToCartHandler(srv *service.CartService, w http.ResponseWriter, r *http.Request) {
	authUserID, err := parseAuthUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var cartBody struct {
		ID       uuid.UUID `json:"id"`
		Quantity int       `json:"quantity"`
	}

	err = json.NewDecoder(r.Body).Decode(&cartBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = srv.AddProduct(authUserID, cartBody.ID, cartBody.Quantity)
	switch {
	case errors.Is(err, service.ErrInvalidQuantity) || errors.Is(err, service.ErrInvalidProduct):
		w.WriteHeader(http.StatusBadRequest)
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteFromCartHandler(srv *service.CartService, w http.ResponseWriter, r *http.Request) {
	authUserID, err := parseAuthUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	productID, err := parseUUID(mux.Vars(r)["productID"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = srv.DeleteProduct(authUserID, productID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func checkoutHandler(srv *service.CartService, w http.ResponseWriter, r *http.Request) {
	authUserID, err := parseAuthUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var checkoutBody struct {
		AddressID uuid.UUID `json:"address_id"`
	}
	err = json.NewDecoder(r.Body).Decode(&checkoutBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderID, err := srv.Checkout(authUserID, checkoutBody.AddressID)
	switch {
	case errors.Is(err, service.ErrEmptyCartCheckout):
		w.WriteHeader(http.StatusNotFound)
		return
	case err != nil:
		w.WriteHeader(http.StatusInternalServerError)
		return
	case err == nil:
		_ = json.NewEncoder(w).Encode(orderID)
	}
}

func healthCheckHandler(_ *service.CartService, w http.ResponseWriter, _ *http.Request) {
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

func getHandlerFunc(
	cartService *service.CartService,
	f func(*service.CartService, http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(cartService, w, r)
	}
}

func NewHTTPHandler(cartService *service.CartService, logger log.Logger) (http.Handler, error) {
	router := mux.NewRouter()

	for _, route := range getRoutes() {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			HandlerFunc(getHandlerFunc(cartService, route.Handler))
	}

	router.Use(transport.NewLoggingMiddleware(logger))
	return router, nil
}
