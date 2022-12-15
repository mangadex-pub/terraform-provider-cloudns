package cloudns

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/matschundbrei/cloudns-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDnsARecord(t *testing.T) {
	testZone := os.Getenv(EnvVarAcceptanceTestsZone)

	initialRecordValue := "1.2.3.4"
	updatedRecordValue := "5.6.7.8"

	genUuid, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	testUuid := genUuid.String()

	initialRecord := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record" {
  name  = "%s"
  zone  = "%s"
  type  = "A"
  value = "%s"
  ttl   = "600"
}
`, testUuid, testZone, initialRecordValue)

	updatedRecord := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record" {
  name  = "%s"
  zone  = "%s"
  type  = "A"
  value = "%s"
  ttl   = "600"
}
`, testUuid, testZone, updatedRecordValue)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: initialRecord,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "name", testUuid),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "zone", testZone),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "type", "A"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "value", initialRecordValue),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "ttl", "600"),
				),
			},
			{
				Config: updatedRecord,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "name", testUuid),
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

func TestAccDnsARecordMultiMatch(t *testing.T) {
	testZone := os.Getenv(EnvVarAcceptanceTestsZone)

	r1value := "1.2.3.4"
	r2value := "5.6.7.8"

	genUuid, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	testUuid := genUuid.String()

	r1res := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record-1" {
  name  = "%s"
  zone  = "%s"
  type  = "A"
  value = "%s"
  ttl   = "600"
}
`, testUuid, testZone, r1value)

	r2res := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record-2" {
  name  = "%s"
  zone  = "%s"
  type  = "A"
  value = "%s"
  ttl   = "600"
}
`, testUuid, testZone, r2value)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s\n%s", r1res, r2res),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record-1", "value", r1value),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record-2", "value", r2value),
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

	genUuid, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	testUuid := genUuid.String()

	initialRecord := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record" {
  name  = "%s"
  zone  = "%s"
  type  = "CNAME"
  value = "%s"
  ttl   = "600"
}
`, testUuid, testZone, initialRecordValue)

	updatedRecord := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record" {
  name  = "%s"
  zone  = "%s"
  type  = "CNAME"
  value = "%s"
  ttl   = "600"
}
`, testUuid, testZone, updatedRecordValue)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: initialRecord,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "name", testUuid),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "zone", testZone),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "type", "CNAME"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "value", initialRecordValue),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "ttl", "600"),
				),
			},
			{
				Config: updatedRecord,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "name", testUuid),
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

func TestAccDnsMXRecord(t *testing.T) {
	testZone := os.Getenv(EnvVarAcceptanceTestsZone)

	initialRecordValue := fmt.Sprintf("target-init.%s", testZone)
	updatedRecordValue := fmt.Sprintf("target-updated.%s", testZone)

	genUuid, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	testUuid := genUuid.String()

	initialRecord := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record" {
  name     = "%s"
  zone     = "%s"
  type     = "MX"
  value    = "%s"
  ttl      = "600"
  priority = "10"
}
`, testUuid, testZone, initialRecordValue)

	updatedRecord := fmt.Sprintf(`
resource "cloudns_dns_record" "some-record" {
  name     = "%s"
  zone     = "%s"
  type     = "MX"
  value    = "%s"
  ttl      = "600"
  priority = "0"
}
`, testUuid, testZone, updatedRecordValue)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: initialRecord,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "name", testUuid),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "zone", testZone),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "type", "MX"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "value", initialRecordValue),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "ttl", "600"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "priority", "10"),
				),
			},
			{
				Config: updatedRecord,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "name", testUuid),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "zone", testZone),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "type", "MX"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "value", updatedRecordValue),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "ttl", "600"),
					resource.TestCheckResourceAttr("cloudns_dns_record.some-record", "priority", "0"),
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
