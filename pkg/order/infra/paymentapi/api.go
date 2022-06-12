package paymentapi

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/event"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/async"
)

const PaymentEventTopicName = "payment_event"

type apiClient struct {
	eventDispatcher event.Dispatcher
}

func (a *apiClient) AuthorizeOrder(orderID uuid.UUID, totalAmount int) error {
	createPayment := struct {
		OrderID     uuid.UUID `json:"order_id"`
		TotalAmount int       `json:"total_amount"`
	}{
		orderID, totalAmount,
	}

	createPaymentJSON, err := json.Marshal(createPayment)
	if err != nil {
		return errors.New("failed to encode json request")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "authorize_payment",
		TopicName: PaymentEventTopicName,
		Key:       orderID.String(),
		Body:      createPaymentJSON,
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func (a *apiClient) CompleteTransaction(orderID uuid.UUID) error {
	orderIDJSON, err := json.Marshal(orderID.String())
	if err != nil {
		return errors.New("failed to encode uuid to json")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "complete_payment",
		TopicName: PaymentEventTopicName,
		Key:       orderID.String(),
		Body:      orderIDJSON,
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func (a *apiClient) CancelPayment(orderID uuid.UUID) error {
	orderIDJSON, err := json.Marshal(orderID.String())
	if err != nil {
		return errors.New("failed to encode uuid to json")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "cancel_payment",
		TopicName: PaymentEventTopicName,
		Key:       orderID.String(),
		Body:      orderIDJSON,
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func New(messageDispatcher event.Dispatcher) async.PaymentAPI {
	return &apiClient{eventDispatcher: messageDispatcher}
}
