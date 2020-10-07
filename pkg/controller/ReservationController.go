package controller

import (
	"gitlab.com/BPeters58/spring-go"
	"gitlab.com/BPeters58/spring-go/pkg/model"
	"gitlab.com/BPeters58/spring-go/pkg/service"
)

func init() {
	spring.AddController(ReservationController{})
}

type ReservationController struct {
	ReservationService service.ReservationService
}

func (rc ReservationController) GetReservation() spring.Handler {
	return spring.Handler{
		Route:       "/",
		Handler:     rc.getReservationImpl,
		Method:      "POST",
		RequestBody: model.Test{},
		Produces:    "application/json",
		Consumes:    "application/json",
	}
}

func (rc ReservationController) getReservationImpl() string {
	return "Your reservation"
}
