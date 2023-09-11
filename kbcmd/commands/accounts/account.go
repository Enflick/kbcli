package accounts

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/killbill/kbcli/v3/kbcmd/cmdlib/args"

	"github.com/killbill/kbcli/v3/kbclient/account"
	"github.com/killbill/kbcli/v3/kbcmd/cmdlib"
	"github.com/killbill/kbcli/v3/kbcmd/kblib"
	"github.com/killbill/kbcli/v3/kbmodel"
	"github.com/urfave/cli"
)

var accountFormatter = cmdlib.Formatter{
	Columns: []cmdlib.Column{
		{
			Name: "NAME",
			Path: "$.name",
		},
		{
			Name: "EXTERNAL_KEY",
			Path: "$.externalKey",
		},
		{
			Name: "ACCOUNT_ID",
			Path: "$.accountId",
		},
		{
			Name: "EMAIL",
			Path: "$.email",
		},
		{
			Name: "BALANCE",
			Path: "$.accountBalance",
		},
		{
			Name: "CURRENCY",
			Path: "$.currency",
		},
	},
}

var (
	createAccountPropertyList args.Properties
	updateAccountPropertyList args.Properties
)

func listAccounts(ctx context.Context, o *cmdlib.Options) error {
	var err error
	resp, err := o.Client().Account.GetAccounts(ctx, &account.GetAccountsParams{})
	if err != nil {
		return err
	}
	o.Print(resp.Payload)
	return nil
}

// getAccount - get account information command
func getAccount(ctx context.Context, o *cmdlib.Options) error {
	if len(o.Args) != 1 {
		return cmdlib.ErrorInvalidArgs
	}

	acc, err := kblib.GetAccountByKeyOrIDWithBalanceAndCBA(ctx, o.Client(), o.Args[0])
	if err == nil {
		o.Print(acc)
	}

	return err
}

func createAccount(ctx context.Context, o *cmdlib.Options) error {
	accToCreate := &kbmodel.Account{}
	err := args.LoadProperties(accToCreate, createAccountPropertyList, o.Args)
	if err != nil {
		return err
	}

	accCreated, err := o.Client().Account.CreateAccount(ctx, &account.CreateAccountParams{
		Body:                  accToCreate,
		ProcessLocationHeader: true,
	})
	if err != nil {
		return err
	}
	o.Print(accCreated.Payload)

	return nil
}

func updateAccount(ctx context.Context, o *cmdlib.Options) error {
	if len(o.Args) < 2 {
		return cmdlib.ErrorInvalidArgs
	}
	key := o.Args[0]

	acc, err := kblib.GetAccountByKeyOrID(ctx, o.Client(), key)
	if err != nil {
		return err
	}
	err = args.LoadProperties(acc, updateAccountPropertyList, o.Args[1:])
	if err != nil {
		return err
	}

	_, err = o.Client().Account.UpdateAccount(ctx, &account.UpdateAccountParams{
		AccountID: acc.AccountID,
		Body:      acc,
	})
	if err != nil {
		return err
	}

	o.Print(acc)
	return nil
}

// getAccount - get account information command
func closeAccount(ctx context.Context, o *cmdlib.Options) error {
	if len(o.Args) < 1 {
		return cmdlib.ErrorInvalidArgs
	}

	accountId := o.Args[0]
	var namedArgs []args.Input
	var err error
	if len(o.Args) > 1 {
		namedArgs, err = args.ParseArgs(o.Args[1:])
		if err != nil {
			return err
		}
	}

	cancelAllSubscriptions := getBoolArg(namedArgs, "cancelAllSubscriptions", false)
	writeOffUnpaidInvoices := getBoolArg(namedArgs, "writeOffUnpaidInvoices", false)
	itemAdjustUnpaidInvoices := getBoolArg(namedArgs, "itemAdjustUnpaidInvoices", false)
	removeFutureNotifications := getBoolArg(namedArgs, "removeFutureNotifications", false)

	_, err = o.Client().Account.CloseAccount(ctx, &account.CloseAccountParams{
		AccountID:                 strfmt.UUID(accountId),
		CancelAllSubscriptions:    &cancelAllSubscriptions,
		WriteOffUnpaidInvoices:    &writeOffUnpaidInvoices,
		ItemAdjustUnpaidInvoices:  &itemAdjustUnpaidInvoices,
		RemoveFutureNotifications: &removeFutureNotifications,
	})
	if err != nil {
		return err
	}
	o.Print("Account successfully closed.")
	return nil
}

// Helper function to retrieve a boolean named argument value from a list of named arguments.
func getBoolArg(namedArgs []args.Input, key string, defaultVal bool) bool {
	for _, arg := range namedArgs {
		if arg.Key == key {
			value, err := strconv.ParseBool(arg.Value)
			if err == nil {
				return value
			}
		}
	}
	return defaultVal
}

var accountNoContentFormatter = cmdlib.Formatter{
	Columns: []cmdlib.Column{
		{
			Name: "Account Operation",
			Path: "",
		},
	},
	CustomFn: cmdlib.CustomFormatter(func(v interface{}, fo cmdlib.FormatOptions) cmdlib.Output {
		return cmdlib.Output{
			Title:   "Account Operation",
			Columns: []string{"Formatted Value"},
			Rows: []cmdlib.OutputRow{
				{
					Values:   []string{"Success"},
					Children: nil, // No child outputs in this example
				},
			},
		}
	}),
}

// RegisterAccountCommands registers all account commands.
func RegisterAccountCommands(r *cmdlib.App) {
	// Register formatters
	cmdlib.AddFormatter(reflect.TypeOf(&kbmodel.Account{}), accountFormatter)

	// Register top level command
	r.Register("", cli.Command{
		Name:    "accounts",
		Aliases: []string{"acc"},
		Usage:   "Account related commands",
	}, nil)

	// Get account
	r.Register("accounts", cli.Command{
		Name:        "get",
		Usage:       "Get account information",
		ArgsUsage:   "ACCOUNT",
		Description: "ACCOUNT can be the account id or external key. An optional + may be prefixed to external key for disambiguation",
	}, getAccount)

	// List all accounts
	r.Register("accounts", cli.Command{
		Name:      "list",
		Usage:     "List all accounts",
		ArgsUsage: "",
	}, listAccounts)

	// Create account
	createAccountPropertyList = args.GetProperties(&kbmodel.Account{})
	createAccountPropertyList.Get("ReferenceTime").Default = time.Now().Format(time.RFC3339)
	createAccountPropertyList.Get("TimeZone").Default = "UTC"
	createAccountPropertyList.Get("Currency").Default = string(kbmodel.AccountCurrencyUSD)
	createAccountPropertyList.Sort(true, true)

	createAccountsUsage := fmt.Sprintf(`%s

		For ex.,:
				kbcmd accounts create ExternalKey=prem1 Name="Prem Ramanathan" Email=prem@prem.com Currency=USD
				`,
		args.GenerateUsageString(&kbmodel.Account{}, createAccountPropertyList))

	r.Register("accounts", cli.Command{
		Name:        "create",
		Usage:       "Create new account",
		ArgsUsage:   createAccountsUsage,
		Description: "Creates new account",
		UsageText: `
		USAGE:
		kbcmd accounts create [options]

			OPTIONS:
			Name=STRING (Required)              : Set the name. E.g., Name="Bob Lazar"
			Email=STRING (Required)             : Set the email address. E.g., Email="bob@myhibrid.com"

			ExternalKey=STRING (Optional, Default: Same as accountId) : Set the unique external key for ID mapping and ensuring idempotency. E.g., ExternalKey="Ext123"
			AccountID=UUID (System Default: Generated UUID) : Set the account ID. E.g., AccountID=123e4567-e89b-12d3-a456-426614174000
			
			Address1=STRING (Optional, Suggested for fraud detection)         : Set the primary address. Knowing the user's address can help in cross-referencing with other data and detecting unusual behavior. E.g., Address1="123 Main St"
			Address2=STRING (Optional)         : Set the secondary address. E.g., Address2="Apt 4B"
			City=STRING (Optional, Suggested for fraud detection)             : Set the city. Geographical data can help in identifying patterns and inconsistencies in user behavior. E.g., City="Los Angeles"
			Company=STRING (Optional, Suggested for fraud detection)          : Set the company name. Associating an account with a legitimate company can reduce the risk profile. E.g., Company="SpaceX"
			Country=STRING (Optional, Suggested for fraud detection)          : Set the country. Different countries have different risk profiles. E.g., Country="USA"
			Locale=STRING (Optional)           : Set the locale. E.g., Locale="en_US"
			Phone=STRING (Optional, Suggested for fraud detection)            : Set the phone number. A verified phone number can be a strong indicator of a genuine user. E.g., Phone="1234567890"
			PostalCode=STRING (Optional, Suggested for fraud detection)       : Set the postal code. This can help in verifying the authenticity of the address provided. E.g., PostalCode="90001"
			State=STRING (Optional, Suggested for fraud detection)            : Set the state. Along with city and country, this provides a more complete geographical profile. E.g., State="CA"
			
			BillCycleDayLocal=INTEGER (Default: 0) : Set the bill cycle day. E.g., BillCycleDayLocal=5
			Currency=STRING (Optional, Default: USD) : Set the currency. E.g., Currency="USD"
			FirstNameLength=INTEGER (Optional) : Set the length of the first name. E.g., FirstNameLength=3
			IsMigrated={True|False} (Optional) : Specify if the account is migrated. E.g., IsMigrated=True
			IsPaymentDelegatedToParent={True|False} (Default: false) : Specify if the payment is delegated to parent.
			ParentAccountID=UUID (Optional)   : Set the parent account ID. E.g., ParentAccountID=123e4567-e89b-12d3-a456-426614174001
			PaymentMethodID=UUID (Optional, Suggested for fraud detection)   : Set the payment method ID. Monitoring the payment methods used can help in detecting fraudulent activities. E.g., PaymentMethodID=123e4567-e89b-12d3-a456-426614174002
			ReferenceTime=DATETIME (System Default: Current time in specified timezone) : Set the reference time. E.g., ReferenceTime="2023-09-10T15:31:59-05:00"
			TimeZone=STRING (Default: UTC, Suggested for fraud detection)     : Set the time zone. Unusual time zone changes can be a red flag. E.g., TimeZone="UTC"

			Notes=STRING (Optional)            : Add any notes. E.g., Notes="VIP customer"

			NOTES:
			The 'auditLogs' parameter is not included in this command. Use a separate command or mechanism to manage or import audit logs.

			External Key:
			When creating a new resource, Kill Bill allocates a unique ID, and you can also associate your own unique external key. This external key is used for:
			1. ID Mapping - Create a mapping between the Kill Bill generated ID and a known key for the resource.
			2. Idempotency - Ensure that each external key is unique. If an API call times out, it can be safely retried with the same external key. This ensures the idempotency of the call, since the external key is unique per tenant and across all resources.

		`,
	}, createAccount)

	// Update account
	updateAccountPropertyList = args.GetProperties(&kbmodel.Account{})
	// Following properties can't change
	updateAccountPropertyList.Remove("ExternalKey")
	updateAccountPropertyList.Remove("AccountID")
	updateAccountPropertyList.Remove("Currency")
	updateAccountPropertyList.Remove("BillCycleDayLocal")
	updateAccountPropertyList.Remove("TimeZone")
	updateAccountPropertyList.Remove("AccountBalance")
	updateAccountPropertyList.Sort(true, true)

	updateAccountsUsage := fmt.Sprintf(`ACCOUNT %s

        For ex.,:
                kbcmd accounts create ExternalKey=prem1 Name="Prem Ramanathan" Email=prem@prem.com Currency=USD
                `,
		args.GenerateUsageString(&kbmodel.Account{}, updateAccountPropertyList))

	r.Register("accounts", cli.Command{
		Name:        "update",
		Usage:       "Updates existing account",
		ArgsUsage:   updateAccountsUsage,
		Description: "Updates existing account",
		UsageText: `
		NAME:
   		kbcmd accounts update - Updates existing account attributes

		USAGE:
		kbcmd accounts update ACCOUNT_KEY_OR_ID
				[Name=STRING]
				[FirstNameLength=INTEGER]
				[Email=STRING]
				[BillCycleDayLocal=INTEGER (Update only once)]
				[ParentAccountID=UUID]
				[IsPaymentDelegatedToParent={True|False}]
				[PaymentMethodID=UUID]
				[Address1=STRING]
				[Address2=STRING]
				[PostalCode=STRING]
				[Company=STRING]
				[City=STRING]
				[State=STRING]
				[Country=STRING]
				[Locale=STRING]
				[Phone=STRING]
				[Notes=STRING]
				[IsMigrated={True|False}]
				[AccountBalance=NUMBER]
				[TreatNullAsReset={True|False}]

		DESCRIPTION:
		Updates selected attributes of an existing account. Certain fields such as externalKey, currency, timeZone, and referenceTime can only be set upon creation and are not updatable. The billCycleDayLocal can be updated but only once.
		`,
	}, updateAccount)

	// close an account
	r.Register("accounts", cli.Command{
		Name:        "close",
		Description: "Close an account",
		ArgsUsage:   "ACCOUNT_ID_HERE --cancelAllSubscriptions=True --writeOffUnpaidInvoices=True",
		UsageText: `
		NAME:
		kbcmd accounts close - Close an account

		USAGE:
		kbcmd accounts close ACCOUNT_ID
				[cancelAllSubscriptions={True|False}]    Default: False
				[writeOffUnpaidInvoices={True|False}]    Default: False
				[itemAdjustUnpaidInvoices={True|False}]  Default: False
				[removeFutureNotifications={True|False}] Default: False

		DESCRIPTION:
		Closes an account, ensuring no other state change will occur on this Account. 
		This endpoint is not for account deletion but brings the account to a stable state. 
		Depending on the value of the query parameters, it can potentially cancel all 
		active subscriptions, write-off unpaid invoices, adjust unpaid invoices, 
		or remove future notifications.

		EXAMPLE:
		kbcmd accounts close 1234-5678-9101-1121 cancelAllSubscriptions=True writeOffUnpaidInvoices=True
		`,
	}, closeAccount)

	registerAccountPaymentCommands(r)
	registerAccountTagCommands(r)
	registerAccountCustomFieldCommands(r)
	registerAccountStripeCommands(r)
}
