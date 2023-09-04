# kbcmd walk through: Creating new subscription and generating invoice

## Prerequisites
This walkthrough assumes that killbill is running
in localhost:8080 with username/password set to 'admin' and 'password'. If your
killbill installation has different values, you can set the values through exporting
the appropriate KB_* variables. (For details run `kbcmd -h`)

## Step 1: Create new tenant
```sh
# Set the api key and secret so future calls will use this.
# These exports ensure kbcmd uses them for future calls.
# Note: You can also pass these in explicitly to each command through flags.

# You can also pass these as parameters on each command. This may be useful if using a script to bootstrap an instance of KB
export KB_API_KEY=tenant1key
export KB_API_SECRET=tenant1secret

kbcmd tenants create tenant1 $KB_API_KEY $KB_API_SECRET
```
This command will print the output when it runs

```
EXTERNAL_KEY API_KEY    API_SECRET
tenant2      tenant2key tenant2secret
```

This is the default output of kbcmd. To change formatting option, you can specify -f option.

## Step 2: Register a Push Notification Callback (optional)
Register a sample Callback URL to receive Events from KB.

In this example, we explicitly setting the host, user and password. We also are setting the api key and api secret.

Note: The default transport scheme will use https and if you're running KB in docker along with not configuring TLS in KB, you will get an error because the client is expecting an https connection. To overcome this, use the transport_scheme parameter and set it `http`
```sh
kbcmd --host "127.0.0.1:8080" --user "admin" --password "password" --api_key "tenant1key" --api_secret "tenant1secret" --transport_scheme "http" tenants registerPushNotificationCallBack http://host.docker.internal:42722/events
```
Same as above and assuming the host, user, password, api key and api secret are configured through environment variables

Note: You may get your commands working since the default username and password along with the host information have default values set up kbcmd that match the default values set in the docker image.
```sh
kbcmd --transport_scheme "http" tenants registerPushNotificationCallBack http://host.docker.internal:42722/events
```

## Step 3: Create new Account
Let's create a new killbill account.

## Step 4: Upload catalog
Upload a sample catalog to the created tenant.
```sh
kbcmd catalog upload docs/samples/simple-catalog.xml
```

Now the catalog has been uploaded.

## Step 5: Create new Account
Let's create a new killbill account.

```sh
kbcmd acc create Name="John Doe" ExternalKey=johndoe Email=johndoe@gmail.com Company="Stark" Currency=USD
```

This command will print the created account
```sh
NAME     EXTERNAL_KEY ACCOUNT_ID                           EMAIL             BALANCE CURRENCY
John Doe johndoe      e4f47a6b-9975-4922-9c5a-baae7ef4d03c johndoe@gmail.com <nil>   USD
```

If you want to get help for specific command, just specify `-h` option. For ex.,
`kbcmd acc create -h` will print help for create command.

## Step 6: Configure stripe (optional)
This step configures stripe plugin. To do this, you need to get stripe account.
Visit https://stripe.com/ and create a new account for yourself. After that, you
can get the public/private key from Developers -> API Keys section.

### Step 6.1 Configure plugin for tenant
```bash
# Set your stripe keys here
STRIPE_PRIVATE_KEY=sk_test_YOUR_STRIPE_KEY
STRIPE_PUBLIC_KEY=pk_test_YOUR_STRIPE_KEY

# Configure stripe plugin with the keys that you got from stripe.com
kbcmd ten configure-stripe-plugin $STRIPE_PUBLIC_KEY $STRIPE_PRIVATE_KEY
```

### Step 6.2 Generate stripe card token
Stripe card token is anonymized credit card information. We will use this instead
of using credit card directly.
```bash
kbcmd stripe --stripe_key $STRIPE_PRIVATE_KEY new-card-token  Name="John Doe" Number=4242424242424242 ExpMonth=08 ExpYear=2025
```
store the card token in `CARD_TOKEN` variable.

### Step 6.3 Set default payment method
```bash
# Add the payment method and set to default
kbcmd acc payment-methods add johndoe killbill-stripe visa true false token=$CARD_TOKEN
```

## Step 7: Create new subscription
```sh
kbcmd subscriptions create ExternalKey=bundle1 Account=johndoe PlanName=simple-monthly
```
This will create a new bundle and print the output.
```
EXTERNAL_KEY BUNDLE_ID
bundle1      447ac056-182f-40a0-a672-8f7a4daa3851
```

## Step 8: Check invoices
Get list of invoices for account `johndoe`.
```bash
kbcmd invoices list johndoe
```
Since the subscription has fixed fees, you should see one invoice with $200.
```
AMOUNT BALANCE INVOICE_ID                           TARGET_DATE
200    <nil>   303b4d21-2809-4e8a-a553-665b313da860 2018-08-03
```

## Step 9: Generate upcoming invoice (Dry Run)
```sh
kbcmd invoices dry-run johndoe
```

# Invoking other APIs

kbcmd -h will print all the supported commands. You can also specify `-h` for a specific command to usage.
For ex., to print usage for `accounts create` command,

```sh
kbcmd accounts create -h
```
