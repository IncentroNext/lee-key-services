
locals {

  // map base identifier to environment variable name
  services = {
    numberservice = "NUMBER_SERVICE",
    orderservice = "ORDER_SERVICE",
    paymentservice = "PAYMENT_SERVICE",
    printservice = "PRINT_SERVICE",
    website = "WEBSITE_SERVICE",
  }

  // map service identifiers to service schemes
  schemes = {for k, v in local.services :
  k => "https://${k}-v1-${var.proj_hash}-ew.a.run.app"
  }

  // map environment variables to service schemes (for Cloud Run config)
  envs = {for k, v in local.services : v => local.schemes[k]}
}
