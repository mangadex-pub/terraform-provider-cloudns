package cloudns

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDnsRecord(t *testing.T) {
	testZone := os.Getenv(EnvVarAcceptanceTestsZone)
	testRecord := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record" {
  name  = "something"
  zone  = "%s"
  type  = "A"
  value = "1.2.3.4"
  ttl   = "600"
}
`, testZone)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testRecord,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "name", "something"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "zone", testZone),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "type", "A"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "value", "1.2.3.4"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "ttl", "600"),
				),
			},
		},
	})
}
