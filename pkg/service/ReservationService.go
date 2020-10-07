package service

import (
	"gitlab.com/BPeters58/spring-go"
)

func init() {
	spring.Container.Singleton(func() ReservationService {
		return ReservationService{}
	})
}

type ReservationService struct{}

func (m ReservationService) Test() string {
	return "ReservationService works"
}
