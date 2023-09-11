package accounts

import (
	"context"
	"reflect"
	"strconv"

	"github.com/go-openapi/strfmt"

	"github.com/killbill/kbcli/v3/kbcmd/cmdlib/args"

	"github.com/killbill/kbcli/v3/kbclient/payment_method"

	"github.com/killbill/kbcli/v3/kbclient/account"
	"github.com/killbill/kbcli/v3/kbcmd/cmdlib"
	"github.com/killbill/kbcli/v3/kbcmd/kblib"
	"github.com/killbill/kbcli/v3/kbmodel"
	"github.com/urfave/cli"
)

var paymentMethodFormatter = cmdlib.Formatter{
	Columns: []cmdlib.Column{
		{
			Name: "NAME",
			Path: "$.pluginName",
		},
		{
			Name: "ID",
			Path: "$.paymentMethodId",
		},
		{
			Name: "IS_DEFAULT",
			Path: "$.isDefault",
		},
	},
}

func listAccountPayments(ctx context.Context, o *cmdlib.Options) error {
	// Ensure the correct number of arguments are provided
	if len(o.Args) != 1 {
		return cmdlib.ErrorInvalidArgs
	}

	accIDOrKey := o.Args[0]

	// Fetch the account using the provided ID or key
	acc, err := kblib.GetAccountByKeyOrID(ctx, o.Client(), accIDOrKey)
	if err != nil {
		return err
	}

	// Always include payment attempts in the response
	var withAttempts = true
	resp, err := o.Client().Account.GetPaymentsForAccount(ctx, &account.GetPaymentsForAccountParams{
		AccountID:    acc.AccountID,
		WithAttempts: &withAttempts,
	})
	if err != nil {
		return err
	}

	// Print the list of payments
	o.Print(resp.Payload)
	return nil
}

func listAccountPaymentMethods(ctx context.Context, o *cmdlib.Options) error {
	// Ensure the correct number of arguments are provided
	if len(o.Args) != 1 {
		return cmdlib.ErrorInvalidArgs
	}

	accIDOrKey := o.Args[0]

	// Fetch the account using the provided ID or key
	acc, err := kblib.GetAccountByKeyOrID(ctx, o.Client(), accIDOrKey)
	if err != nil {
		return err
	}

	// Retrieve the payment methods associated with the account
	resp, err := o.Client().Account.GetPaymentMethodsForAccount(ctx, &account.GetPaymentMethodsForAccountParams{
		AccountID: acc.AccountID,
	})
	if err != nil {
		return err
	}

	// Print the list of payment methods
	o.Print(resp.Payload)
	return nil
}

func getAccountPaymentMethod(ctx context.Context, o *cmdlib.Options) error {
	// Ensure the correct number of arguments are provided
	if len(o.Args) != 1 {
		return cmdlib.ErrorInvalidArgs
	}

	paymentMethodID := o.Args[0]

	// Fetch the payment method using the provided ID
	resp, err := o.Client().PaymentMethod.GetPaymentMethod(ctx, &payment_method.GetPaymentMethodParams{
		PaymentMethodID: strfmt.UUID(paymentMethodID),
	})
	if err != nil {
		return err
	}

	// Print the retrieved payment method details
	o.Print(resp.Payload)
	return nil
}

func addAccountPaymentMethod(ctx context.Context, o *cmdlib.Options) error {
	if len(o.Args) < 4 {
		return cmdlib.ErrorInvalidArgs
	}

	accKey, method, externalKey := o.Args[0], o.Args[1], o.Args[2]

	isDefault, err := strconv.ParseBool(o.Args[3])
	if err != nil {
		return err
	}

	// Initialize payAllUnpaidInvoices with its zero value (false)
	var payAllUnpaidInvoices bool
	if len(o.Args) > 4 {
		payAllUnpaidInvoices, err = strconv.ParseBool(o.Args[4])
		if err != nil {
			return err
		}
	}

	acc, err := kblib.GetAccountByKeyOrID(ctx, o.Client(), accKey)
	if err != nil {
		return err
	}

	var pluginProperties []args.Input
	if len(o.Args) > 5 {
		pluginProperties, err = args.ParseArgs(o.Args[5:])
		if err != nil {
			return err
		}
	}

	// Construct the payment method and add properties
	pm := &kbmodel.PaymentMethod{
		ExternalKey: externalKey,
		PluginName:  method,
		PluginInfo:  &kbmodel.PaymentMethodPluginDetail{},
	}
	for _, pp := range pluginProperties {
		pm.PluginInfo.Properties = append(pm.PluginInfo.Properties, &kbmodel.PluginProperty{
			Key:   pp.Key,
			Value: pp.Value,
		})
	}

	resp, err := o.Client().Account.CreatePaymentMethod(ctx, &account.CreatePaymentMethodParams{
		AccountID:             acc.AccountID,
		Body:                  pm,
		IsDefault:             &isDefault,
		PayAllUnpaidInvoices:  &payAllUnpaidInvoices,
		ProcessLocationHeader: true,
	})
	if err != nil {
		return err
	}

	o.Print(resp.Payload)
	return err
}

func removeAccountPaymentMethod(ctx context.Context, o *cmdlib.Options) error {
	// Check for at least one argument
	if len(o.Args) < 1 {
		return cmdlib.ErrorInvalidArgs
	}

	paymentMethodID := o.Args[0]
	force := false // Default value

	// If a second argument is provided, attempt to parse it as a boolean for "force"
	if len(o.Args) > 1 {
		var err error
		force, err = strconv.ParseBool(o.Args[1])
		if err != nil {
			return err
		}
	}

	// Call the client method to delete the payment method
	_, err := o.Client().PaymentMethod.DeletePaymentMethod(ctx, &payment_method.DeletePaymentMethodParams{
		PaymentMethodID:        strfmt.UUID(paymentMethodID),
		ForceDefaultPmDeletion: &force,
	})

	return err
}

func refreshStripe(ctx context.Context, o *cmdlib.Options) error {
	if len(o.Args) < 1 {
		return cmdlib.ErrorInvalidArgs
	}

	pluginName := "killbill:payment-test-plugin"

	return refreshPaymentMethods(ctx, o, pluginName)
}

func refreshAllPaymentMethods(ctx context.Context, o *cmdlib.Options) error {
	if len(o.Args) < 1 {
		return cmdlib.ErrorInvalidArgs
	}

	return refreshPaymentMethods(ctx, o, "")
}

func refreshPaymentMethods(ctx context.Context, o *cmdlib.Options, pluginName string) error {
	if len(o.Args) < 1 {
		return cmdlib.ErrorInvalidArgs
	}

	accKey := o.Args[0]

	acc, err := kblib.GetAccountByKeyOrID(ctx, o.Client(), accKey)
	if err != nil {
		return err
	}
	params := &account.RefreshPaymentMethodsParams{
		AccountID: acc.AccountID,
	}
	if pluginName != "" {
		params.PluginName = &pluginName
	}
	_, err = o.Client().Account.RefreshPaymentMethods(ctx, params)

	return err
}

// updateDefaultPaymentMethod - Update default payment method for account
func updateDefaultPaymentMethod(ctx context.Context, o *cmdlib.Options) error {
	if len(o.Args) < 2 {
		return cmdlib.ErrorInvalidArgs
	}
	accIDOrName := o.Args[0]
	paymentMethodID := o.Args[1]
	payAllUnpaidInvoices := false

	if len(o.Args) > 2 && o.Args[2] != "" {
		var err error
		payAllUnpaidInvoices, err = strconv.ParseBool(o.Args[2])
		if err != nil {
			o.Print("error parsing PAY_ALL_UNPAID_INVOICES as a boolean value")
			return err
		}
	}

	acc, err := kblib.GetAccountByKeyOrID(ctx, o.Client(), accIDOrName)
	if err != nil {
		return err
	}

	var pluginPropertiesArgs []args.Input
	if len(o.Args) > 3 {
		pluginPropertiesArgs, err = args.ParseArgs(o.Args[3:])
		if err != nil {
			return err
		}
	}

	// Convert parsed arguments to the expected format for PluginProperty
	pluginProperties := make([]string, len(pluginPropertiesArgs))
	for i, pp := range pluginPropertiesArgs {
		pluginProperties[i] = pp.Key + "=" + pp.Value
	}

	_, err = o.Client().Account.SetDefaultPaymentMethod(ctx, &account.SetDefaultPaymentMethodParams{
		AccountID:             acc.AccountID,
		PaymentMethodID:       strfmt.UUID(paymentMethodID),
		ProcessLocationHeader: true,
		PayAllUnpaidInvoices:  &payAllUnpaidInvoices,
		PluginProperty:        pluginProperties,
	})

	return err
}

func registerAccountPaymentCommands(r *cmdlib.App) {
	// Payment method
	cmdlib.AddFormatter(reflect.TypeOf(&kbmodel.PaymentMethod{}), paymentMethodFormatter)

	// Payments
	r.Register("accounts", cli.Command{
		Name:    "payments",
		Aliases: []string{},
		Usage:   "Payments related commands",
	}, nil)

	// List payments
	r.Register("accounts.payments", cli.Command{
		Name:      "list",
		Aliases:   []string{"ls"},
		Usage:     "List payments for the given account",
		ArgsUsage: `ACCOUNT`,
	}, listAccountPayments)

	// payment methods
	r.Register("accounts", cli.Command{
		Name:    "payment-methods",
		Aliases: []string{"pm"},
		Usage:   "Payment method related commands",
	}, nil)

	// List payment methods
	r.Register("accounts.payment-methods", cli.Command{
		Name:      "list",
		Aliases:   []string{"ls"},
		Usage:     "List payment methods for the given account",
		ArgsUsage: `ACCOUNT`,
	}, listAccountPaymentMethods)

	// Get payment method by ID
	r.Register("accounts.payment-methods", cli.Command{
		Name:      "get",
		Usage:     "Get payment method for the given account",
		ArgsUsage: `PAYMENT_METHOD_ID`,
	}, getAccountPaymentMethod)

	// Add payment method
	r.Register("accounts.payment-methods", cli.Command{
		Name:  "add",
		Usage: "Add new payment method",
		ArgsUsage: `ACCOUNT PLUGIN_NAME EXTERNAL_KEY IS_DEFAULT PAY_ALL_UNPAID_INVOICES [Property1=Value1] ...

   For ex.,
      kbcmd accounts payment-methods add johndoe killbill-stripe visa true false token=tok_1CidZ7HGlIo9NLGOy7sPvbsz
		`,
	}, addAccountPaymentMethod)

	// Remove payment method
	r.Register("accounts.payment-methods", cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "Remove payment method",
		ArgsUsage: `PM_METHOD_ID [FORCE]`,
	}, removeAccountPaymentMethod)

	// Refresh all payment methods
	r.Register("accounts.payment-methods", cli.Command{
		Name:      "sync-payment-methods",
		Aliases:   []string{"sync"},
		Usage:     "Sync all the payment methods for the account",
		ArgsUsage: `ACCOUNT_ID `,
	}, refreshAllPaymentMethods)

	r.Register("accounts.payment-methods", cli.Command{
		Name:      "set-default",
		Aliases:   []string{"default"},
		Usage:     "Set default payment method for account",
		ArgsUsage: `ACCOUNT_ID_OR_NAME PAYMENT_METHOD_ID [PAY_ALL_UNPAID_INVOICES] [PLUGIN_PROPERTY]`,
		UsageText: `
		USAGE:
   		kbcmd set-default-payment-method ACCOUNT_ID_OR_NAME PAYMENT_METHOD_ID [PAY_ALL_UNPAID_INVOICES] [PLUGIN_PROPERTY]

		ARGUMENTS:
		ACCOUNT_ID_OR_NAME (Required)      : The account's unique ID or name.
		PAYMENT_METHOD_ID (Required)       : The unique ID of the payment method you want to set as default.
		PAY_ALL_UNPAID_INVOICES (Optional, Default: false) : Choose true to pay all unpaid invoices. E.g., true or false.
		PLUGIN_PROPERTY (Optional)         : List of plugin properties, if any. Multiple properties can be passed separated by a comma.

		NOTES:
		If successful, the command returns a status code of 204 and an empty body. Ensure the provided account ID or name and payment method ID are valid.
		Plugin properties are passed as key-value pairs separated by spaces. If a key or value contains spaces, wrap them in quotes. E.g., "key1=value1 key2=value2".
		`,
	}, updateDefaultPaymentMethod,
	)

}
