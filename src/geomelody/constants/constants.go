package constants

import (
	"os"
)

const (
	API_PATH = "api/v1/geomelody"
)

var (
	LAST_API_URL = ""
	LAST_API_KEY = ""

	MUSIC_MIX_URL     = ""
	MUSIC_MIX_API_KEY = ""
)

func InitConstantsVars() {
	LAST_API_URL = os.Getenv("LAST_API_URL")
	LAST_API_KEY = os.Getenv("LAST_API_KEY")

	MUSIC_MIX_URL = os.Getenv("MUSIC_MIX_URL")
	MUSIC_MIX_API_KEY = os.Getenv("MUSIC_MIX_API_KEY")
}
