package transport

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/klwxsrx/arch-course-project/pkg/catalog/app/query"
	"github.com/klwxsrx/arch-course-project/pkg/catalog/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/transport"
	"net/http"
)

type route struct {
	Name    string
	Method  string
	Pattern string
	Handler func(*service.ProductService, query.ProductService, http.ResponseWriter, *http.Request)
}

func getRoutes() []route {
	return []route{
		{
			"getProductsWeb",
			http.MethodGet,
			"/web/products",
			getProductsHandler,
		},
		{
			"getProductsByIDs",
			http.MethodGet,
			"/products",
			getProductsByIDsHandler,
		},
		{
			"addProduct",
			http.MethodPut,
			"/products",
			addProductHandler,
		},
		{
			"updateProduct",
			http.MethodPatch,
			"/product/{productID}",
			updateProductHandler,
		},
		{
			"health",
			http.MethodGet,
			"/healthz",
			healthCheckHandler,
		},
	}
}

type productJSONSchema struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int    `json:"price"`
}

type productWithIDJSONSchema struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int       `json:"price"`
}

func getProductsHandler(_ *service.ProductService, service query.ProductService, w http.ResponseWriter, _ *http.Request) {
	products, err := service.ListAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := make([]productWithIDJSONSchema, 0, len(products))
	for _, product := range products {
		result = append(result, productWithIDJSONSchema{
			ID:          product.ID,
			Title:       product.Title,
			Description: product.Description,
			Price:       product.Price,
		})
	}

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getProductsByIDsHandler(_ *service.ProductService, service query.ProductService, w http.ResponseWriter, r *http.Request) {
	var productIDs []uuid.UUID
	err := json.NewDecoder(r.Body).Decode(&productIDs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	products, err := service.GetByIDs(productIDs)
	if errors.Is(err, query.ErrProductByIDNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := make([]productWithIDJSONSchema, 0, len(products))
	for _, product := range products {
		result = append(result, productWithIDJSONSchema{
			ID:          product.ID,
			Title:       product.Title,
			Description: product.Description,
			Price:       product.Price,
		})
	}

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func addProductHandler(srv *service.ProductService, _ query.ProductService, w http.ResponseWriter, r *http.Request) {
	var productBody productJSONSchema
	err := json.NewDecoder(r.Body).Decode(&productBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	productID, err := srv.Add(productBody.Title, productBody.Description, productBody.Price)
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	case service.ErrInvalidProperty:
		w.WriteHeader(http.StatusBadRequest)
	case nil:
		_ = json.NewEncoder(w).Encode(productID)
	}
}

func updateProductHandler(srv *service.ProductService, _ query.ProductService, w http.ResponseWriter, r *http.Request) {
	productID, err := parseUUID(mux.Vars(r)["productID"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var productBody productJSONSchema
	err = json.NewDecoder(r.Body).Decode(&productBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = srv.Update(productID, productBody.Title, productBody.Description, productBody.Price)
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	case service.ErrInvalidProperty:
		w.WriteHeader(http.StatusBadRequest)
	case service.ErrProductNotExists:
		w.WriteHeader(http.StatusNotFound)
	case nil:
		w.WriteHeader(http.StatusNoContent)
	}
}

func healthCheckHandler(_ *service.ProductService, _ query.ProductService, w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode(struct {
		Status string `json:"status"`
	}{"OK"})
}

func parseUUID(str string) (uuid.UUID, error) {
	return uuid.Parse(str)
}

func getHandlerFunc(
	service *service.ProductService,
	query query.ProductService,
	f func(*service.ProductService, query.ProductService, http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		f(service, query, w, r)
	}
}

func NewHTTPHandler(service *service.ProductService, query query.ProductService, logger log.Logger) (http.Handler, error) {
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
