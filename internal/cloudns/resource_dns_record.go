package cloudns

import (
	"context"
	"fmt"
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
	recordCreated, err := recordToCreate.Create(&clientConfig.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(recordCreated.ID)

	// do not re-read after creation to avoid potential stale caches on the underlying zone listing
	return nil
}

func resourceDnsRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ClientConfig)
	record := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("READ %s.%s %d in %s %s", record.Host, record.Domain, record.TTL, record.Rtype, record.Record))
	recordRead, err := record.Read(&config.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	expectedId := d.Id()

	// If we get a record response with a mismatched ID on it, then the record was externally tampered with and
	// another matching record exists (aside from the ID), so the underlying lib returns that (wtf)
	if expectedId != recordRead.ID {
		d.SetId("")
		return nil
	}

	err = d.Set("name", recordRead.Host)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("zone", recordRead.Domain)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("type", recordRead.Rtype)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("value", recordRead.Record)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("ttl", recordRead.TTL)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDnsRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(ClientConfig)
	record := toApiRecord(d)

	tflog.Debug(ctx, fmt.Sprintf("UPDATE %s.%s %d in %s %s", record.Host, record.Domain, record.TTL, record.Rtype, record.Record))
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
	_, err := record.Destroy(&config.apiAccess)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceDnsRecordRead(ctx, d, meta)
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
