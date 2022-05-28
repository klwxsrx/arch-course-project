package service

import (
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/app/persistence"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/domain"
)

var (
	ErrDeliveryIsNotScheduled = errors.New("delivery doesn't have scheduled status")
)

type DeliveryService struct {
	ufw    persistence.UnitOfWork
	logger log.Logger
}

func (s *DeliveryService) Schedule(orderID uuid.UUID, addressID uuid.UUID) error {
	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		_, err := p.DeliveryRepository().GetByID(orderID)
		if err == nil {
			return nil // already created
		}
		if !errors.Is(err, domain.ErrItemNotFound) {
			return err
		}

		delivery := &domain.Delivery{
			OrderID: orderID,
			Status:  domain.DeliveryStatusScheduled,
			Address: s.getAddress(addressID),
		}

		return p.DeliveryRepository().Store(delivery)
	})
	if err != nil {
		s.logger.WithError(err).With(log.Fields{
			"orderID":   orderID,
			"addressID": addressID,
		}).Error("failed to schedule delivery")
	}
	return err
}

func (s *DeliveryService) CancelSchedule(orderID uuid.UUID) error {
	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		delivery, err := p.DeliveryRepository().GetByID(orderID)
		if err != nil {
			return err
		}

		if delivery.Status == domain.DeliveryStatusCancelled {
			return nil
		}
		if delivery.Status != domain.DeliveryStatusScheduled {
			return ErrDeliveryIsNotScheduled
		}

		delivery.Status = domain.DeliveryStatusCancelled
		return p.DeliveryRepository().Store(delivery)
	})
	if err != nil && !errors.Is(err, ErrDeliveryIsNotScheduled) {
		s.logger.WithError(err).With(log.Fields{
			"orderID": orderID,
		}).Error("failed to delete delivery schedule")
	}
	return err
}

func (s *DeliveryService) getAddress(addressID uuid.UUID) string {
	return "Санкт-Петербург, пр. Тореза, дом 30, подъезд 1, кв. 10" // TODO: store address from database
}

func NewDeliveryService(ufw persistence.UnitOfWork, logger log.Logger) *DeliveryService {
	return &DeliveryService{
		ufw:    ufw,
		logger: logger,
	}
}
