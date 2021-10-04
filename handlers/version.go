package handlers

import (
	"github.com/AlexandrGurkin/vm_agent/models"
	"github.com/AlexandrGurkin/vm_agent/restapi/operations/version"
	"github.com/go-openapi/runtime/middleware"
)

type VersionHandler struct {
}

func (vh VersionHandler) Handle(_ version.GetVersionParams) middleware.Responder {
	return version.NewGetVersionOK().WithPayload(&models.ResponseVersion{
		Version:   "1",
		Branch:    "2",
		Commit:    "3",
		BuildTime: "4",
	})
}
