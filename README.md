# Terraform Provider ClouDNS

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.1.x (older versions may work but are entirely untested)
- [Go](https://golang.org/doc/install) >= 1.19
- [ClouDNS](https://cloudns.net) API credentials and a pre-existing DNS zone manageable by the user/sub-user associated with said credentials

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```sh
$ go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules). Please see the Go documentation for the most up to date information about using Go
modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

Ensure that you have an API user/sub-user on ClouDNS (requires a paid subscription with reseller access).

> Note that using a sub-user which you delegate a specific zone to is a **much** safer approach and should always be your first choice

Once that is done, you must pre-create the zones you will want to manage on ClouDNS side (technically they are manageable through the API)

## Limitations

The following features are knowingly not part of the provider's initial implementation:

- Zone management
- Complex records (anything using optional parameters [here](https://www.cloudns.net/wiki/article/58/))

The main reason is that I find writing Golang extremely unpleasant and would rather avoid doing more of it than strictly necessary.

Please feel free to contribute features however; we will happily review and merge them.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ CLOUDNS_SUB_AUTH_ID=1234 CLOUDNS_PASSWORD=verysecret CLOUDNS_ACCEPTANCE_TESTS_ZONE=some-test-zone.net make testacc
```
