provider "cloudns" {
  # Optionally provided by CLOUDNS_SUB_AUTH_ID
  sub_auth_id = 1234

  # Optionally provided by CLOUDNS_PASSWORD
  password = "verysecret"

  # Optional, ClouDNS currently maxxes out at 20 requests per second per ip. Defaults to 5.
  rate_limit = 5
}
