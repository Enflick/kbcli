package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/killbill/kbcli/v3/kbmodel"

	"github.com/go-openapi/strfmt"
	"github.com/killbill/kbcli/v3/kbclient/admin"
	"github.com/killbill/kbcli/v3/kbcmd/cmdlib"
	"github.com/urfave/cli"
)

type adminFormatter struct {
	Message string `json:"message"`
}

func getQueues(ctx context.Context, o *cmdlib.Options) error {

	params := &admin.GetQueueEntriesParams{
		WithBusEvents:     BoolPtr(true),
		WithHistory:       BoolPtr(true),
		WithInProcessing:  BoolPtr(true),
		WithNotifications: BoolPtr(true),
	}

	for _, arg := range o.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return errors.New("invalid argument format")
		}
		argName := parts[0]
		argValue := parts[1]

		switch argName {
		case "accountId":
			uid := strfmt.UUID(argValue)
			params.AccountID = &uid
		case "queueName":
			params.QueueName = &argValue
		case "serviceName":
			params.ServiceName = &argValue
		case "withHistory":
			val := parseBool(argValue)
			params.WithHistory = &val
		case "minDate":
			params.MinDate = &argValue
		case "maxDate":
			params.MaxDate = &argValue
		case "withInProcessing":
			val := parseBool(argValue)
			params.WithInProcessing = &val
		case "withBusEvents":
			val := parseBool(argValue)
			params.WithBusEvents = &val
		case "withNotifications":
			val := parseBool(argValue)
			params.WithNotifications = &val
		default:
			return errors.New("unknown argument: " + argName)
		}
	}

	resp, err := o.Client().Admin.GetQueueEntries(ctx, params)
	if err != nil {
		return err
	}
	if resp.IsSuccess() {
		result := adminFormatter{}
		result.Message = "Successfully retrieved queue entries"
		o.Print(&result)
		bodyBytes, err := io.ReadAll(resp.HttpResponse.Body())
		if err != nil {
			return err
		}
		o.Print(string(bodyBytes))
	}

	return nil
}

// BoolPtr is a utility function to return a pointer to a bool value.
func BoolPtr(b bool) *bool {
	return &b
}

// parseBool is a utility function to parse a string into a bool value.
func parseBool(s string) bool {
	return s == "true"
}

func putInRotation(ctx context.Context, o *cmdlib.Options) error {
	var err error

	if len(o.Args) != 0 {
		return cmdlib.ErrorInvalidArgs
	}

	resp, err := o.Client().Admin.PutInRotation(ctx, &admin.PutInRotationParams{})
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		result := adminFormatter{}
		result.Message = "failed to put instance in rotation"
		o.Print(&result)
		return err
	}
	result := adminFormatter{Message: "Successfully put instance in rotation"}
	o.Print(&result)
	return nil
}

func pullFromRotation(ctx context.Context, o *cmdlib.Options) error {
	var err error
	if len(o.Args) != 0 {
		return cmdlib.ErrorInvalidArgs
	}

	resp, err := o.Client().Admin.PutOutOfRotation(ctx, &admin.PutOutOfRotationParams{})
	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		result := adminFormatter{}
		result.Message = "failed to remove instance from rotation"
		o.Print(&result)
		return err
	}
	result := adminFormatter{}
	result.Message = "Successfully removed instance from rotation"
	o.Print(&result)
	return nil
}

func invalidateTenantCache(ctx context.Context, o *cmdlib.Options) error {
	var err error
	if len(o.Args) != 0 {
		return cmdlib.ErrorInvalidArgs
	}

	resp, err := o.Client().Admin.InvalidatesCacheByTenant(ctx, &admin.InvalidatesCacheByTenantParams{})
	if err != nil {
		return err
	}
	o.Print(resp.Code())
	if !resp.IsSuccess() {
		result := adminFormatter{}
		result.Message = "failed to invalidate tenant cache"
		o.Print(&result)
		return err
	}
	result := adminFormatter{}
	result.Message = "Successfully invalidated tenant cache"
	o.Print(&result)
	return nil
}

func invalidateAllCachesForInstance(ctx context.Context, o *cmdlib.Options) error {
	var err error
	if len(o.Args) != 0 {
		return cmdlib.ErrorInvalidArgs
	}

	resp, err := o.Client().Admin.InvalidatesCache(ctx, &admin.InvalidatesCacheParams{})
	if err != nil {
		return err
	}
	o.Print(resp.Code())
	if !resp.IsSuccess() {
		result := adminFormatter{}
		result.Message = "failed to invalidate tenant cache"
		o.Print(&result)
		return err
	}
	result := adminFormatter{}
	result.Message = "successfully invalidated tenant cache"
	o.Print(&result)
	return nil
}

func invalidateAccountCache(ctx context.Context, o *cmdlib.Options) error {
	var err error
	if len(o.Args) != 1 {
		return cmdlib.ErrorInvalidArgs
	}
	accountIdString := o.Args[0]
	accountId := strfmt.UUID(accountIdString)

	resp, err := o.Client().Admin.InvalidatesCacheByAccount(ctx, &admin.InvalidatesCacheByAccountParams{
		AccountID: accountId,
	})
	if err != nil {
		return err
	}
	o.Print(resp.Code())
	if !resp.IsSuccess() {
		result := adminFormatter{}
		result.Message = "failed to invalidate account cache"
		o.Print(&result)
		return err
	}
	result := adminFormatter{}
	result.Message = "successfully invalidated account cache"
	o.Print(&result)
	return nil
}

func generateInvoicesForParkedAccounts(ctx context.Context, o *cmdlib.Options) error {
	params := &admin.TriggerInvoiceGenerationForParkedAccountsParams{
		Limit:  Int64Ptr(100), // Default value for Limit
		Offset: Int64Ptr(0),   // Default value for Offset
	}

	for _, arg := range o.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			argName := parts[0]
			argValue := parts[1]

			switch argName {
			case "limit":
				val, err := strconv.ParseInt(argValue, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid value for limit: %s", argValue)
				}
				params.Limit = &val
			case "offset":
				val, err := strconv.ParseInt(argValue, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid value for offset: %s", argValue)
				}
				params.Offset = &val
			default:
				return fmt.Errorf("unknown argument: %s", argName)
			}
		} else {
			// Assuming positional arguments
			if params.Limit == nil {
				val, err := strconv.ParseInt(parts[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid value for positional limit: %s", parts[0])
				}
				params.Limit = &val
			} else if params.Offset == nil {
				val, err := strconv.ParseInt(parts[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid value for positional offset: %s", parts[0])
				}
				params.Offset = &val
			}
		}
	}

	resp, err := o.Client().Admin.TriggerInvoiceGenerationForParkedAccounts(ctx, params)
	if err != nil {
		return err
	}
	o.Print(resp.Code())
	if !resp.IsSuccess() {
		result := adminFormatter{}
		result.Message = "failed to trigger invoice generation for parked accounts"
		o.Print(&result)
		return err
	}
	result := adminFormatter{}
	result.Message = "successfully triggered invoice generation for parked accounts"
	o.Print(&result)
	return nil
}

// Int64Ptr is a utility function to return a pointer to an int64 value.
func Int64Ptr(i int64) *int64 {
	return &i
}

func updatePaymentTransactionState(ctx context.Context, o *cmdlib.Options) error {
	if len(o.Args) < 5 {
		return cmdlib.ErrorInvalidArgs
	}

	var lastSuccessfulPaymentState, currentPaymentStateName, transactionStatus, paymentIdString, paymentTransactionIdString string

	for _, arg := range o.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			argName := parts[0]
			argValue := parts[1]

			switch argName {
			case "lastSuccessfulPaymentState":
				lastSuccessfulPaymentState = argValue
			case "currentPaymentStateName":
				currentPaymentStateName = argValue
			case "transactionStatus":
				transactionStatus = argValue
			case "paymentId":
				paymentIdString = argValue
			case "paymentTransactionId":
				paymentTransactionIdString = argValue
			default:
				return fmt.Errorf("unknown argument: %s", argName)
			}
		} else {
			// Assuming positional arguments
			if lastSuccessfulPaymentState == "" {
				lastSuccessfulPaymentState = parts[0]
			} else if currentPaymentStateName == "" {
				currentPaymentStateName = parts[0]
			} else if transactionStatus == "" {
				transactionStatus = parts[0]
			} else if paymentIdString == "" {
				paymentIdString = parts[0]
			} else if paymentTransactionIdString == "" {
				paymentTransactionIdString = parts[0]
			}
		}
	}

	paymentId := strfmt.UUID(paymentIdString)
	paymentTransactionId := strfmt.UUID(paymentTransactionIdString)

	_, err := o.Client().Admin.UpdatePaymentTransactionState(ctx, &admin.UpdatePaymentTransactionStateParams{
		Body: &kbmodel.AdminPayment{
			LastSuccessPaymentState: lastSuccessfulPaymentState,
			CurrentPaymentStateName: currentPaymentStateName,
			TransactionStatus:       transactionStatus,
		},
		PaymentID:            paymentId,
		PaymentTransactionID: paymentTransactionId,
	})
	if err != nil {
		return err
	}
	result := adminFormatter{}
	result.Message = "successfully updated payment transaction state"
	o.Print(&result)
	return nil
}

var simpleSuccessOrFailFormatter = cmdlib.Formatter{
	Columns: []cmdlib.Column{
		{
			Name: "Result",
			Path: "$.message",
		},
	},
}

func registerAdminCommands(r *cmdlib.App) {
	cmdlib.AddFormatter(reflect.TypeOf(&admin.GetQueueEntriesOK{}), simpleSuccessOrFailFormatter)
	cmdlib.AddFormatter(reflect.TypeOf(&adminFormatter{}), simpleSuccessOrFailFormatter)
	// Register top level command
	r.Register("", cli.Command{
		Name:  "admin",
		Usage: "development/debugging related commands",
	}, nil)

	r.Register("admin", cli.Command{
		Name: "get-queues",
		Usage: `
- accountId=<UUID>          : Filter by account ID. (Default: all accounts)
- queueName=<string>        : Filter by queue name. (Default: all queues)
- serviceName=<string>      : Filter by service name. (Default: all services)
- withHistory=<true|false>  : If true, include history. (Default: true)
- minDate=<string>          : Specify the earliest date for history. (Default: from the beginning)
- maxDate=<string>          : Specify the latest date for history. (Default: current date)
- withInProcessing=<true|false> : If true, include entries in processing. (Default: true)
- withBusEvents=<true|false>    : If true, include bus events. (Default: true)
- withNotifications=<true|false>: If true, include notifications. (Default: true)

Usage Example:
getQueues --accountId=12345-6789-abcd-efgh --queueName=myQueue --withHistory=false

Note: Arguments are optional and can be used in any combination.
		`,
	}, getQueues)

	// Instance Rotation
	r.Register("admin", cli.Command{
		Name:      "put-in-rotation",
		Usage:     "Put a server instance in rotation",
		ArgsUsage: ``,
	}, putInRotation)

	r.Register("admin", cli.Command{
		Name:      "take-from-rotation",
		Usage:     "Pull a server instance from rotation",
		ArgsUsage: ``,
	}, pullFromRotation)

	r.Register("admin", cli.Command{
		Name:      "invalidate-tenant-cache",
		Usage:     "Invalidates the tenant cache for the tenant to which the API Key and Secret belong",
		ArgsUsage: ``,
	}, invalidateTenantCache)

	r.Register("admin", cli.Command{
		Name:      "invalidate-account-cache",
		Usage:     "Invalidates the account cache for the given account in the tenant to which the API Key and Secret belong",
		ArgsUsage: ``,
	}, invalidateAccountCache)

	r.Register("admin", cli.Command{
		Name:      "invalidate-instance-cache",
		Usage:     "Invalidates all the cache for the given instance",
		ArgsUsage: ``,
	}, invalidateAllCachesForInstance)

	r.Register("admin", cli.Command{
		Name:  "trigger-invoices-for-parked-accounts",
		Usage: "Triggers the invoicing for all parked accounts. Should be ran after the issue that caused the account to be parked is resolved",
		ArgsUsage: `
		Positional or Named Arguments:
- limit=<int64>                          : Number of results to retrieve. (Default: 100)
- offset=<int64>                         : Starting point for results retrieval.

Usage Example (Positional):
generateInvoicesForParkedAccounts 50 10

Usage Example (Named):
generateInvoicesForParkedAccounts --limit=50 --offset=10

Note: Arguments are optional and can be used in any combination.
		`,
	}, generateInvoicesForParkedAccounts)

	r.Register("admin", cli.Command{
		Name:  "update-payment-transaction-state",
		Usage: "Provides a way to fix the payment state data for a given Payment, if that data becomes corrupted. This could happen, for example, if a call to a third party payment gateway times out, leaving the system in an unknown state.",
		ArgsUsage: `
Positional or Named Arguments:
- lastSuccessfulPaymentState=<string>    : Previous successful payment state.
- currentPaymentStateName=<string>       : Current payment state name.
- transactionStatus=<string>             : Transaction status.
- paymentId=<UUID>                       : Payment ID.
- paymentTransactionId=<UUID>            : Payment transaction ID.

Usage Example (Positional):
updatePaymentTransactionState state1 state2 success 12345-6789-abcd-efgh 67890-abcd-efgh-12345

Usage Example (Named):
updatePaymentTransactionState --lastSuccessfulPaymentState=state1 --currentPaymentStateName=state2 --transactionStatus=success --paymentId=12345-6789-abcd-efgh --paymentTransactionId=67890-abcd-efgh-12345

		`,
	}, updatePaymentTransactionState)
}
