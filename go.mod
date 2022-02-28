module gitlab.org/mangadex-pub/terraform-provider-cloudns

go 1.15

require (
	github.com/hashicorp/terraform-plugin-docs v0.5.1
	github.com/hashicorp/terraform-plugin-log v0.2.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.10.1
	github.com/sta-travel/cloudns-go v0.0.0-20200217151626-28ac622cf3b2
)

replace github.com/sta-travel/cloudns-go v0.0.0-20200217151626-28ac622cf3b2 => github.com/matschundbrei/cloudns-go v0.0.0-20200217151626-28ac622cf3b2
