resource "cloudns_dns_record" "some-record" {
  # something.cloudns.net 600 in A 1.2.3.4
  host        = "something"
  domain_name = "cloudns.net"
  ttl         = "600"
  record_type = "A"
  record      = "1.2.3.4"
}
