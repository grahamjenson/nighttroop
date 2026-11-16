package static

import "nighttroop.t1d.nz/utils/templui/utils"

// The components' scripts are linked with their file's version, as the pages' own are, not with the time the server
// started, so a browser keeps them across restarts and deploys that do not change them
func init() {
	utils.ScriptURL = URL
}
