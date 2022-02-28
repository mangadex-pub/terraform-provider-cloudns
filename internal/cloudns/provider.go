package cloudns

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sta-travel/cloudns-go"
)

const EnvVarAuthId = "CLOUDNS_AUTH_ID"
const EnvVarSubAuthId = "CLOUDNS_SUB_AUTH_ID"
const EnvVarPassword = "CLOUDNS_PASSWORD"

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New() func() *schema.Provider {
	return func() *schema.Provider {

		providerSchema := map[string]*schema.Schema{
			"auth_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc(EnvVarAuthId, nil),
				Description: fmt.Sprintf("When using api users, this is the `auth-id`. It is read from the environment variable `%s` if not passed explicitly. Mutually exclusive with `sub_auth_id`.", EnvVarAuthId),
			},
			"sub_auth_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc(EnvVarSubAuthId, nil),
				Description: fmt.Sprintf("When using api sub-users, this is the `sub-auth-id`. It is read from the environment variable `%s` if not passed explicitly. Mutually exclusive with `auth_id`.", EnvVarSubAuthId),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc(EnvVarPassword, nil),
				Description: fmt.Sprintf("This is the password associated with your auth-id or sub-auth-id. It is read from the environment variable `%s` if not passed explicitly.", EnvVarPassword),
			},
		}

		p := &schema.Provider{
			Schema:         providerSchema,
			DataSourcesMap: map[string]*schema.Resource{},
			ResourcesMap: map[string]*schema.Resource{
				"cloudns_dns_record": resourceDnsRecord(),
			},
		}

		p.ConfigureContextFunc = configure()

		return p
	}
}

type ClientConfig struct {
	apiAccess cloudns.Apiaccess
}

func configure() func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		authId := d.Get("auth_id").(int)
		subAuthId := d.Get("sub_auth_id").(int)
		password := d.Get("password").(string)

		if len(password) == 0 {
			return nil, diag.Errorf("Expected password to be defined but it wasn't")
		}

		if (authId != 0) == (subAuthId != 0) {
			var golangSucks = "not defined"
			if authId != 0 {
				golangSucks = "defined"
			}
			return nil, diag.Errorf("Exactly one of auth_id or sub_auth_id must be set, but both were %s", golangSucks)
		}

		if authId != 0 {
			return ClientConfig{
				apiAccess: cloudns.Apiaccess{
					Authid:       authId,
					Authpassword: password,
				},
			}, nil
		} else {
			return ClientConfig{
				apiAccess: cloudns.Apiaccess{
					Subauthid:    subAuthId,
					Authpassword: password,
				},
			}, nil
		}

	}
}
