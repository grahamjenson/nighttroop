

# API Endpoints

$nsURL/api/v2/authorization/request/token=$token
$nsURL/api/v1/entries.json?token=$token&find[date][\$gte]=${gteCreatedAt.millisecondsSinceEpoch}&count=1000
$nsURL/api/v1/treatments.json?token=$token&find[created_at][\$gte]=${toNSDateString(gteCreatedAt)}&count=1000
$nsURL/api/v1/treatments.json?token=$token&find[eventType]=Profile+Switch&count=$count
$nsURL/api/v1/treatments.json?token=$token&find[eventType]=Note&count=$count
$nsURL/api/v1/profile/current.json?token=$token
$nsURL/api/v2/properties.json?token=$token
$nsURL/api/v1/devicestatus.json?token=$token&count=1

// Add Carbs endpoint
$nsURL/api/v3/treatments

// JWT endpoint
$nsURL/api/v2/authorization/request/token=$token
