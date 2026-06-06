resource "technitium_dns_app_config" "blocker" {
  name   = "Advanced Blocking"
  config = jsonencode({ "enableBlocking" = true })
}
