
# Adding an A record on the apex of the "something.cloudns.net" zone
resource "cloudns_dns_record" "some-record" {
  # something.cloudns.net 600 in A 1.2.3.4
  name  = ""
  zone  = "something.cloudns.net"
  type  = "A"
  value = "1.2.3.4"
  ttl   = "600"
}


# Adding an A record on the "something.cloudns.net" zone
resource "cloudns_dns_record" "some-record" {
  # something-else.something.cloudns.net 600 in A 1.2.3.4
  name  = "something-else"
  zone  = "something.cloudns.net"
  type  = "A"
  value = "1.2.3.5"
  ttl   = "600"
}


# Adding an MX record on the apex of the "something.cloudns.net" zone
resource "cloudns_dns_record" "some-record" {
  # something.cloudns.net 600 in MX mail.example.com
  name     = ""
  zone     = "something.cloudns.net"
  type     = "MX"
  value    = "mail.example.com"
  ttl      = "3600"
  priority = "20"
}

