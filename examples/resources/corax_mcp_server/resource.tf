# Copyright (c) Trifork

resource "corax_mcp_server" "corax_data" {
  name      = "Corax Data"
  url       = "http://corax-data-mcp-cheetah-application.corax-ai.svc.cluster.local:8000/mcp"
  type      = "streamablehttp"
  is_public = true

  config = {
    token = {
      type    = "header"
      label   = "Authorization"
      default = null
    }
    filters = {
      type     = "header"
      label    = "X-Filters"
      default  = null
      required = false
    }
    timeZone = {
      type    = "header"
      label   = "X-Time-Zone"
      default = null
    }
    collectionId = {
      type    = "header"
      label   = "X-Collection-Id"
      default = null
    }
    dataProducts = {
      type    = "header"
      label   = "X-Data-Products"
      default = null
    }
    collectionType = {
      type    = "header"
      label   = "X-Collection-Type"
      default = null
    }
    userPreferences = {
      type     = "header"
      label    = "X-User-Preferences"
      default  = null
      required = false
    }
  }
}
