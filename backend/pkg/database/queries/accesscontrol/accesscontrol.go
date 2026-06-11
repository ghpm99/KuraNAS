package accesscontrol

import _ "embed"

//go:embed get_allowed_ips.sql
var GetAllowedIPsQuery string

//go:embed get_allowed_ip_by_id.sql
var GetAllowedIPByIDQuery string

//go:embed create_allowed_ip.sql
var CreateAllowedIPQuery string

//go:embed update_allowed_ip.sql
var UpdateAllowedIPQuery string

//go:embed delete_allowed_ip.sql
var DeleteAllowedIPQuery string
