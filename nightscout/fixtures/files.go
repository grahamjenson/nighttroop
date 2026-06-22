package fixtures_tests

import _ "embed"

//go:embed current_profile_loop.json
var CurrentProfileLoopFixture string

//go:embed current_profile.json
var CurrentProfileFixture string

//go:embed devicestatus_loop.json
var DeviceStatusLoopFixture string

//go:embed devicestatus.json
var DeviceStatusFixture string

//go:embed entries.json
var EntriesFixture []byte

//go:embed treatments.json
var TreatmentsFixture []byte

//go:embed status.json
var StatusFixture []byte

//go:embed properties.json
var PropertiesFixture []byte
