package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/persistence"
	"github.com/klwxsrx/arch-course-project/pkg/payment/domain"
)

type PaymentService struct {
	ufw    persistence.UnitOfWork
	logger log.Logger
}

func (s *PaymentService) AuthorizePayment(orderID uuid.UUID, totalAmount int) error {
	// TODO: authorize payment from payment gateway

	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		payment, err := p.PaymentRepository().GetByID(orderID)
		if err != nil && !errors.Is(err, domain.ErrPaymentNotFound) {
			return fmt.Errorf("failed to get payment: %w", err)
		}
		if err == nil {
			return nil
		}

		payment = &domain.Payment{
			OrderID:     orderID,
			TotalAmount: totalAmount,
			Status:      domain.PaymentStatusAuthorized,
		}

		err = p.PaymentRepository().Store(payment)
		if err != nil {
			return fmt.Errorf("failed to store payment: %w", err)
		}

		err = p.OrderAPI().NotifyPaymentAuthorized(orderID)
		if err != nil {
			return fmt.Errorf("failed to notify payment authorized: %w", err)
		}
		return nil
	})
	if err != nil {
		s.logger.WithError(err).With(log.Fields{
			"orderID":     orderID,
			"totalAmount": totalAmount,
		}).Error("failed to authorize payment")
		return err
	}

	s.logger.With(log.Fields{
		"orderID":     orderID,
		"totalAmount": totalAmount,
	}).Info("payment authorized")
	return nil
}

func (s *PaymentService) CompletePayment(orderID uuid.UUID) error {
	// TODO: complete payment from payment gateway

	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		payment, err := p.PaymentRepository().GetByID(orderID)
		if errors.Is(err, domain.ErrPaymentNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get payment: %w", err)
		}
		if payment.Status != domain.PaymentStatusAuthorized {
			return nil
		}

		payment.Status = domain.PaymentStatusCompleted
		err = p.PaymentRepository().Store(payment)
		if err != nil {
			return fmt.Errorf("failed to store completed payment: %w", err)
		}

		err = p.OrderAPI().NotifyPaymentCompleted(orderID)
		if err != nil {
			return fmt.Errorf("failed to notify payment completion: %w", err)
		}
		return nil
	})
	if err != nil {
		s.logger.WithError(err).With(log.Fields{
			"orderID": orderID,
		}).Error("failed to complete payment")
		return err
	}

	s.logger.With(log.Fields{
		"orderID": orderID,
	}).Info("payment completed")
	return nil
}

func (s *PaymentService) CancelPayment(orderID uuid.UUID) error {
	// TODO: cancel payment from payment gateway

	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		payment, err := p.PaymentRepository().GetByID(orderID)
		if errors.Is(err, domain.ErrPaymentNotFound) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to get payment: %w", err)
		}

		if payment.Status != domain.PaymentStatusAuthorized {
			return nil
		}

		payment.Status = domain.PaymentStatusCancelled
		err = p.PaymentRepository().Store(payment)
		if err != nil {
			return fmt.Errorf("failed to store cancelled payment: %w", err)
		}
		return nil
	})
	if err != nil {
		s.logger.WithError(err).With(log.Fields{
			"orderID": orderID,
		}).Error("failed to cancel payment")
		return err
	}

	s.logger.With(log.Fields{
		"orderID": orderID,
	}).Info("payment cancelled")
	return nil
}

func NewPaymentService(ufw persistence.UnitOfWork, logger log.Logger) *PaymentService {
	return &PaymentService{ufw: ufw, logger: logger}
}
