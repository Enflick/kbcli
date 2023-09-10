package commands

import (
	"github.com/go-openapi/runtime"
	"github.com/killbill/kbcli/v3/kbcmd/cmdlib"
	"github.com/killbill/kbcli/v3/kbcmd/commands/accounts"
	"github.com/killbill/kbcli/v3/kbcmd/commands/subscriptions"
)

// RegisterAll registers all commands
func RegisterAll(r *cmdlib.App) {
	accounts.RegisterAccountCommands(r)
	subscriptions.RegisterCommands(r)

	registerAuditLogCommands(r)
	registerBundleCommands(r)
	registerCatalogCommands(r)
	registerCustomFieldFunctions(r)
	registerInvoicesCommands(r)
	registerStripeCommands(r)
	registerTagDefinitionCommands(r)
	registerTenantCommands(r)
	registerAdminCommands(r)
	registerNodesInfoCommands(r)

	// Dev
	registerDevCommands(r)
}

type HttpResponseHandler struct {
	HttpResponse runtime.ClientResponse
}

func (h *HttpResponseHandler) IsSuccess() bool {
	return h.HttpResponse.Code() >= 200 && h.HttpResponse.Code() < 300
}

func (h *HttpResponseHandler) IsRedirect() bool {
	return h.HttpResponse.Code() >= 300 && h.HttpResponse.Code() < 400
}

func (h *HttpResponseHandler) IsClientError() bool {
	return h.HttpResponse.Code() >= 400 && h.HttpResponse.Code() < 500
}

func (h *HttpResponseHandler) IsServerError() bool {
	return h.HttpResponse.Code() >= 500
}

func (h *HttpResponseHandler) IsCode(code int) bool {
	return h.HttpResponse.Code() == code
}
