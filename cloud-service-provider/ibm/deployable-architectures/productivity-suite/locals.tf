locals {
  env = {
    for tuple in regexall("(.*)=(.*)", file(".env")) : tuple[0] => tuple[1]
  }

  hf_token = sensitive(local.env["HF_TOKEN"])
  openai_api_key = sensitive(local.env["OPENAI_API_KEY"])
}