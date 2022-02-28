package cloudns

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/sta-travel/cloudns-go"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDnsARecord(t *testing.T) {
	testZone := os.Getenv(EnvVarAcceptanceTestsZone)

	initialRecordValue := "1.2.3.4"
	updatedRecordValue := "5.6.7.8"

	initialRecord := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record" {
  name  = "something"
  zone  = "%s"
  type  = "A"
  value = "%s"
  ttl   = "600"
}
`, testZone, initialRecordValue)

	updatedRecord := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record" {
  name  = "something"
  zone  = "%s"
  type  = "A"
  value = "%s"
  ttl   = "600"
}
`, testZone, updatedRecordValue)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: initialRecord,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "name", "something"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "zone", testZone),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "type", "A"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "value", initialRecordValue),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "ttl", "600"),
				),
			},
			{
				Config: updatedRecord,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "name", "something"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "zone", testZone),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "type", "A"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "value", updatedRecordValue),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "ttl", "600"),
				),
			},
		},
		CheckDestroy: CheckDestroyedRecords(testZone),
	})
}

func TestAccDnsCNAMERecord(t *testing.T) {
	testZone := os.Getenv(EnvVarAcceptanceTestsZone)

	initialRecordValue := fmt.Sprintf("target-init.%s", testZone)
	updatedRecordValue := fmt.Sprintf("target-updated.%s", testZone)

	initialRecord := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record" {
  name  = "something"
  zone  = "%s"
  type  = "CNAME"
  value = "%s"
  ttl   = "600"
}
`, testZone, initialRecordValue)

	updatedRecord := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record" {
  name  = "something"
  zone  = "%s"
  type  = "CNAME"
  value = "%s"
  ttl   = "600"
}
`, testZone, updatedRecordValue)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: initialRecord,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "name", "something"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "zone", testZone),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "type", "CNAME"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "value", initialRecordValue),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "ttl", "600"),
				),
			},
			{
				Config: updatedRecord,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "name", "something"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "zone", testZone),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "type", "CNAME"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "value", updatedRecordValue),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "ttl", "600"),
				),
			},
		},
		CheckDestroy: CheckDestroyedRecords(testZone),
	})
}

func CheckDestroyedRecords(zone string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		provider := testAccProvider
		apiAccess := provider.Meta().(ClientConfig).apiAccess
		records, err := cloudns.Zone{
			Ztype:  "master",
			Domain: zone,
		}.List(&apiAccess)

		if err != nil {
			return err
		}

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "cloudns_dns_record" {
				continue
			}

			fmt.Printf("Checking that cloudns_dns_record#%s was properly deleted\n", rs.Primary.ID)

			for _, record := range records {
				existingRecordId := record.ID
				if rs.Primary.ID == existingRecordId {
					return fmt.Errorf(
						"record %s (%s.%s %d in %s %s) still exists",
						record.ID,
						record.Host,
						record.Domain,
						record.TTL,
						record.Rtype,
						record.Record,
					)
				}
			}
		}

		return nil
	}
}
