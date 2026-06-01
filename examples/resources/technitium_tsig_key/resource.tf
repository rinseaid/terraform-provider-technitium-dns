resource "technitium_tsig_key" "transfer" {
  key_name  = "xfer-key.example.com"
  algorithm = "hmac-sha256"
}
