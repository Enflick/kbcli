package accounts

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/killbill/kbcli/v3/kbclient/account"
	"github.com/killbill/kbcli/v3/kbcmd/cmdlib"
	"github.com/killbill/kbcli/v3/kbcmd/kblib"
	"github.com/killbill/kbcli/v3/kbmodel"
	"github.com/urfave/cli"
)

func addStripeCustomerId(ctx context.Context, o *cmdlib.Options) error {
	//accounts stripe add-customerid cus_XXXXX
	if len(o.Args) != 2 {
		return cmdlib.ErrorInvalidArgs
	}
	accID := o.Args[0]
	stripeCustomerId := o.Args[1]

	o.Args = []string{accID, "STRIPE_CUSTOMER_ID", stripeCustomerId}
	return addCustomField(ctx, o)
}

func addStripeTokenToAccount(ctx context.Context, o *cmdlib.Options) error {
	if err := validateAndExtractArgs(o); err != nil {
		return err
	}

	accIDOrKey := o.Args[0]
	stripeToken := o.Args[1]
	stripeCustomerId := o.Args[2]
	overrideCustomerId, _ := strconv.ParseBool(o.Args[3])
	isNewDefault, _ := strconv.ParseBool(o.Args[4])

	acc, err := kblib.GetAccountByKeyOrID(ctx, o.Client(), accIDOrKey)
	if err != nil {
		return err
	}

	if stripeCustomerId != "" {
		if err := handleStripeCustomerID(ctx, o, acc, stripeCustomerId, overrideCustomerId); err != nil {
			return err
		}
	}

	return addStripePaymentMethod(ctx, o, acc, stripeToken, isNewDefault)
}

func validateAndExtractArgs(o *cmdlib.Options) error {
	switch len(o.Args) {
	case 2:
		// Only ACCOUNT_ID and STRIPE_TOKEN are provided.
		return nil
	case 4:
		// All arguments except IS_NEW_DEFAULT are provided. Validate the boolean for OVERRIDE_CUSTOMER_ID.
		_, err := strconv.ParseBool(o.Args[3])
		if err != nil {
			return errors.New("error parsing OVERRIDE_CUSTOMER_ID as a boolean value")
		}
		return nil
	case 5:
		// All arguments are provided. Validate the booleans for OVERRIDE_CUSTOMER_ID and IS_NEW_DEFAULT.
		_, err1 := strconv.ParseBool(o.Args[3])
		_, err2 := strconv.ParseBool(o.Args[4])
		if err1 != nil || err2 != nil {
			return errors.New("error parsing boolean values from arguments")
		}
		return nil
	default:
		return cmdlib.ErrorInvalidArgs
	}
}

func handleStripeCustomerID(ctx context.Context, o *cmdlib.Options, acc *kbmodel.Account, stripeCustomerId string, override bool) error {
	customField, err := getCustomFieldByNameOrID(ctx, o.Client(), acc.AccountID, "STRIPE_CUSTOMER_ID")
	if err != nil {
		if _, ok := err.(*account.GetAccountCustomFieldsBadRequest); ok {
			return createStripeCustomerId(ctx, o, acc, stripeCustomerId)
		}
		return err
	}
	if *customField.Value == stripeCustomerId {
		o.Print("Stripe customer id already exists and matches the one provided")
	} else {
		if override {
			return updateStripeCustomerId(ctx, o, acc, stripeCustomerId)
		}
	}
	return nil
}

func createStripeCustomerId(ctx context.Context, o *cmdlib.Options, acc *kbmodel.Account, stripeCustomerId string) error {
	o.Args = []string{acc.AccountID.String(), "STRIPE_CUSTOMER_ID", stripeCustomerId}
	if err := addCustomField(ctx, o); err != nil {
		o.Print("Error creating custom field to add the stripe customer id")
		return err
	}
	return nil
}

func updateStripeCustomerId(ctx context.Context, o *cmdlib.Options, acc *kbmodel.Account, stripeCustomerId string) error {
	o.Print("Stripe customer id already exists but does not match the one provided")
	o.Print("Updating the stripe customer id")
	o.Args = []string{acc.AccountID.String(), "STRIPE_CUSTOMER_ID", stripeCustomerId}
	if err := updateCustomField(ctx, o); err != nil {
		o.Print("Error updating custom field to add the stripe customer id")
		return err
	}
	return nil
}

func addStripePaymentMethod(ctx context.Context, o *cmdlib.Options, acc *kbmodel.Account, stripeToken string, isNewDefault bool) error {
	o.Args = []string{acc.AccountID.String(), "killbill-stripe", "", strconv.FormatBool(isNewDefault), fmt.Sprintf("token=%s", stripeToken)}
	if err := addAccountPaymentMethod(ctx, o); err != nil {
		o.Print("Error adding the stripe token as a payment method")
		return err
	}
	return nil
}

func registerAccountStripeCommands(r *cmdlib.App) {

	r.Register("accounts", cli.Command{
		Name:    "stripe",
		Aliases: []string{"strp"},
		Usage:   "Stripe related commands",
	}, nil)

	r.Register("accounts.stripe", cli.Command{
		Name:    "add-customerid",
		Aliases: []string{"set-id"},
		Usage:   "Add a customer stripe id to an account",
		ArgsUsage: `ACCOUNT VALUE

   For e.g.,
      accounts stripe add-customerid accountId cus_XXXXX`,
	}, addStripeCustomerId)

	// Refresh Stripe payment method
	r.Register("accounts.stripe.payment-methods", cli.Command{
		Name:      "sync-payment-methods",
		Aliases:   []string{"sync"},
		Usage:     "Sync the payment methods with stripe",
		ArgsUsage: `ACCOUNT_ID `,
	}, refreshStripe)

	r.Register("accounts.stripe.payment-methods", cli.Command{
		Name:      "save-token-payment-method",
		Aliases:   []string{"add-token"},
		Usage:     "Create a new payment method from a stripe token",
		ArgsUsage: `ACCOUNT_ID STRIPE_TOKEN [STRIPE_CUSTOMER_ID] [OVERRIDE_CUSTOMER_ID] [IS_NEW_DEFAULT]`,
		UsageText: `
This command creates a new payment method in the system using a Stripe token for a specific account.

- ACCOUNT_ID: The unique identifier or key of the account for which you want to save the payment method.
- STRIPE_TOKEN: The token generated by Stripe for the payment method.
- STRIPE_CUSTOMER_ID (optional, but required if OVERRIDE_CUSTOMER_ID is provided): The Stripe Customer ID you want to associate with this account.
- OVERRIDE_CUSTOMER_ID (optional, but required if STRIPE_CUSTOMER_ID is provided): A boolean flag (true/false) indicating whether to override the existing Stripe Customer ID if there's a mismatch between the provided ID and the one stored.
- IS_NEW_DEFAULT (optional): A boolean flag (true/false) indicating if the provided Stripe token should become the new default payment method for the account.

Examples:
1. To save a Stripe token as a new payment method without other options:
   > save-token-payment-method ACCOUNT_ID STRIPE_TOKEN

2. To save a Stripe token, associate a Stripe Customer ID, and override the existing Customer ID:
   > save-token-payment-method ACCOUNT_ID STRIPE_TOKEN STRIPE_CUSTOMER_ID OVERRIDE_CUSTOMER_ID

3. To save a Stripe token, associate a Stripe Customer ID, override the existing Customer ID, and set the token as the new default payment method:
   > save-token-payment-method ACCOUNT_ID STRIPE_TOKEN STRIPE_CUSTOMER_ID OVERRIDE_CUSTOMER_ID IS_NEW_DEFAULT
`,
	}, addStripeTokenToAccount)

}
