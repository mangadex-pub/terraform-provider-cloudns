package cloudns

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sta-travel/cloudns-go"
)

func resourceDnsRecord() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "A simple DNS record.",

		CreateContext: resourceDnsRecordCreate,
		ReadContext:   resourceDnsRecordRead,
		UpdateContext: resourceDnsRecordUpdate,
		DeleteContext: resourceDnsRecordDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceImportStateContext,
		},

		// Naming **does not** follow the scheme used by ClouDNS, due to how comically misleading and unclear it is
		// see: https://www.cloudns.net/wiki/article/58/ for the relevant "vanilla" schema on ClouDNS side
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the record (eg: `[something].cloudns.net 600 in A 1.2.3.4`)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
			},
			"zone": {
				Description: "The zone on which to add the record (eg: `something.[cloudns.net] 600 in A 1.2.3.4`)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"ttl": {
				Description: "The TTL to assign to the record (eg: `something.cloudns.net [600] in A 1.2.3.4`)",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    false,
			},
			"type": {
				Description: "The type of record (eg: `something.cloudns.net 600 in [A] 1.2.3.4`)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
			},
			"value": {
				Description: "Value of the record (eg: `something.cloudns.net 600 in A [1.2.3.4]`)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
			},
		},
	}
}

func resourceDnsRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clientConfig := meta.(ClientConfig)
	recordToCreate := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("CREATE %s.%s %d in %s %s", recordToCreate.Host, recordToCreate.Domain, recordToCreate.TTL, recordToCreate.Rtype, recordToCreate.Record))

	clientConfig.rateLimiter.Take()
	recordCreated, err := recordToCreate.Create(&clientConfig.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(recordCreated.ID)

	return resourceDnsRecordRead(ctx, d, meta)
}

func resourceDnsRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Id() == "" {
		d.SetId("")
		return nil
	}

	config := meta.(ClientConfig)
	lookup := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("READ Record#%s (%s.%s %d in %s %s)", lookup.ID, lookup.Host, lookup.Domain, lookup.TTL, lookup.Rtype, lookup.Record))

	config.rateLimiter.Take()
	zoneRead, err := cloudns.Zone{Domain: lookup.Domain}.List(&config.apiAccess)
	if err != nil {
		tflog.Error(ctx, err.Error())
		return diag.FromErr(err)
	}

	for _, zoneRecord := range zoneRead {
		wantedId := d.Id()
		actualId := zoneRecord.ID
		if wantedId == actualId {
			err = updateState(d, &zoneRecord)
			if err != nil {
				tflog.Error(ctx, err.Error())
				return diag.FromErr(err)
			}
			return nil
		}
	}

	return nil
}

func resourceDnsRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ClientConfig)
	record := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("UPDATE %s.%s %d in %s %s", record.Host, record.Domain, record.TTL, record.Rtype, record.Record))

	config.rateLimiter.Take()
	updated, err := record.Update(&config.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(updated.ID)

	return resourceDnsRecordRead(ctx, d, meta)
}

func resourceDnsRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ClientConfig)
	record := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("DELETE %s.%s %d in %s %s", record.Host, record.Domain, record.TTL, record.Rtype, record.Record))

	config.rateLimiter.Take()
	_, err := record.Destroy(&config.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDnsRecordRead(ctx, d, meta)
}

func resourceImportStateContext(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(ClientConfig)
	tflog.Debug(ctx, fmt.Sprintf("IMPORT ID: %#v", d.Id()))

	// Extract zone name and record id
	halves := strings.Split(d.Id(), "/")
	interestingHalf := halves[1]
	parts := strings.Split(interestingHalf, "_")
	if len(parts) != 3 {
		errMsg := fmt.Sprintf("Could not extract zone from record ID: %s", d.Id())
		tflog.Error(ctx, errMsg)
		return nil, errors.New(errMsg)
	}
	zone := fmt.Sprintf("%s.%s", parts[0], parts[1])
	wantedID := parts[2]

	tflog.Debug(ctx, fmt.Sprintf("Trying to import record %s from zone: %s", wantedID, zone))

	config.rateLimiter.Take()
	zoneRead, err := cloudns.Zone{Domain: zone}.List(&config.apiAccess)
	if err != nil {
		return nil, err
	}

	recordFound := false

	// Try to find record by ID from the zone record list.
	// @todo There is a more direct way to doing this with search parameters.
	for _, zoneRecord := range zoneRead {
		actualId := zoneRecord.ID
		//tflog.Debug(ctx, fmt.Sprintf("wantedID: %s, actualID: %s", wantedID, actualId))
		if wantedID == actualId {
			err = updateState(d, &zoneRecord)
			if err != nil {
				tflog.Error(ctx, fmt.Sprintf("IMPORT error updateState(): %s", err))
				return nil, err
			}
			tflog.Debug(ctx, fmt.Sprintf("IMPORT %s.%s %d in %s %s", zoneRecord.Host, zoneRecord.Domain, zoneRecord.TTL, zoneRecord.Rtype, zoneRecord.Record))
			recordFound = true
			break
		}
	}

	if !recordFound {
		errMsg := fmt.Sprintf("IMPORT error: could not find record with ID: %s", wantedID)
		tflog.Error(ctx, errMsg)
		return nil, errors.New(errMsg)
	}

	return []*schema.ResourceData{d}, nil
}

func updateState(d *schema.ResourceData, zoneRecord *cloudns.Record) error {
	err := d.Set("name", zoneRecord.Host)
	if err != nil {
		return err
	}

	err = d.Set("zone", zoneRecord.Domain)
	if err != nil {
		return err
	}

	err = d.Set("type", zoneRecord.Rtype)
	if err != nil {
		return err
	}

	err = d.Set("value", zoneRecord.Record)
	if err != nil {
		return err
	}

	err = d.Set("ttl", zoneRecord.TTL)
	if err != nil {
		return err
	}

	return nil
}

func toApiRecord(d *schema.ResourceData) cloudns.Record {
	id := d.Id()
	name := d.Get("name").(string)
	zone := d.Get("zone").(string)
	rtype := d.Get("type").(string)
	value := d.Get("value").(string)
	ttl := d.Get("ttl").(int)

	return cloudns.Record{
		ID:     id,
		Host:   name,
		Domain: zone,
		Rtype:  rtype,
		Record: value,
		TTL:    ttl,
	}
}
