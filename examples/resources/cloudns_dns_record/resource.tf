resource "cloudns_dns_record" "some-record" {
  # something.cloudns.net 600 in A 1.2.3.4
  name  = ""
  zone  = "something.cloudns.net"
  type  = "A"
  value = "1.2.3.4"
  ttl   = "600"
}
