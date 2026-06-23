package nightscout

//go:generate goverter gen .

// goverter:converter
//
// goverter:output:file ./z_converters.go
type MapConverter interface {
	//ActivityToDBModel(source Activity) user.Activity
}
