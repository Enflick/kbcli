# Kill Bill go client library and kill bill command line
This repository contains killbill go client library (kbclient)
and killbill command line tool (kbcmd)

## Versions

| KB Version | KBCli Version |
|------------|---------------|
| 0.20.x     | 1.x.y         |
| 0.22.x     | 2.x.y         |
| 0.24.x     | 3.x.y         |



## Kill bill go client library
Kill bill go client library is a go package that can be used to connect to Kill Bill server.

### Install
```bash
go get -u github.com/killbill/kbcli/v3
```

### Creating new client
```go
    trp := httptransport.New("127.0.0.1:8080", "", nil)
    // Add text/xml producer which is not handled by openapi runtime.
    trp.Producers["text/xml"] = runtime.TextProducer()
    // Set this to true to dump http messages
    trp.Debug = false
    // Authentication
    authWriter := runtime.ClientAuthInfoWriterFunc(func(r runtime.ClientRequest, _ strfmt.Registry) error {
        encoded := base64.StdEncoding.EncodeToString([]byte("admin"/*username*/ + ":" + "password" /**password*/))
        if err := r.SetHeaderParam("Authorization", "Basic "+encoded); err != nil {
            return err
        }
        if err := r.SetHeaderParam("X-KillBill-ApiKey", apiKey); err != nil {
            return err
        }
        if err := r.SetHeaderParam("X-KillBill-ApiSecret", apiSecret); err != nil {
            return err
        }
        return nil
    })
    client := kbclient.New(trp, strfmt.Default, authWriter, kbclient.KillbillDefaults{})
```

Look at the [complete example here](examples/listaccounts/main.go).
For more examples, look at [kbcmd tool](kbcmd/README.md).

### Wrapper client

We also provide a client wrapper (a higher level and easier to use api) that is built (not generated) on top of the generated code.
See the package `killbill` package under the `wrapper` directory.

There is a suite of test `client_test.go` that shows how to use it.

## Go-Swagger Integration and Client Generation

We've integrated go-swagger into our build process to allow for easy generation of client libraries based on our API's Swagger definitions.

### How to Use

1. **Finding and Setting up Go-Swagger**

   Run the following command to automatically search for the `go-swagger` directory:
   ```
   make find-swagger
   ```

   **Logic Behind Directory Search:**
   - The script starts by checking the immediate parent directory of the Makefile.
   - If not found, it proceeds to the grandparent directory.
   - If the `go-swagger` directory is still not located, it will iterate over each subdirectory of the grandparent directory to find it.
   
   **Assumptions and Directory Structure:**
   The script assumes a directory structure similar to `/path/to/base/github.com/your-org/your-repo/`. The `base` directory (e.g., `github.com` or `git.mena.technology`) is dynamically determined based on the original location of the Makefile. This ensures the correct formation of the `git clone` command and the destination path for cloning.

   If the directory isn't found in the above locations, the script will prompt you to clone the `go-swagger` repository. It creates the `git clone` command based on the directory structure, ensuring that the cloned repo is placed in the correct location relative to the original Makefile.

2. **Generating Client Libraries**

   To generate the client libraries, use:
   ```
   make generate-extensions
   ```

   This will generate the client based on the `kbswagger.yaml` definition.

3. **Post-Generation Cleanup**

   After generating the client, it's recommended to run, and will be ran for you:
   ```
   make clean-module
   ```
   This command will execute `go mod tidy -compat=$go_version_in_go.mod`, ensuring that your module information is cleaned and up-to-date. An informational message will be displayed before this command runs.

### Manual Client code generation

We use a forked version of the `go-swagger` client hosted under https://github.com/killbill/go-swagger.
Every so often we rebase our fork from upstream to keep it up-to-date. Given a version of a `swagger` binary, the
client can be regenerated using:

`swagger  generate client -f kbswagger.yaml -m kbmodel -c kbclient --default-scheme=http`

### Generating dev extensions
The original project proposed separating out the clock test API but this is moved to always be available since Kill Bill will return an error when it's not in test mode. The API, however, is always available.

## Kill bill command line tool (kbcmd)
kbcmd is a command line tool that uses the go client library. This tool can do many of the
kill bill operations. More details are [available here in README](kbcmd/README.md).