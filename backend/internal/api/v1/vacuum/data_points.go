package vacuum

const (
	dpWholeHouseCleaning   = "101"
	dpPause                = "102"
	dpReturnToDock         = "103"
	dpState                = "105"
	dpBatteryPercent       = "106"
	dpCleaningSeconds      = "107"
	dpCleaningArea         = "108"
	dpSuction              = "109"
	dpWaterLevel           = "110"
	dpRemoteControl        = "111"
	dpLocate               = "112"
	dpDoNotDisturb         = "113"
	dpVolume               = "114"
	dpConsumableReset      = "115"
	dpTotalCleaningSeconds = "116"
	dpTotalCleaningArea    = "117"
	dpTotalCleaningCount   = "118"
	dpSideBrushSeconds     = "119"
	dpMainBrushSeconds     = "120"
	dpFilterSeconds        = "121"
	dpPath                 = "123"
	dpMapCommand           = "127"
	dpFaultBitmap          = "122"
	dpAutoEmpty            = "135"
	dpAutoEmptyFrequency   = "136"
	dpCarpetBoost          = "137"
	dpResumeAfterCharge    = "151"
)

const (
	maxFaultCode = 30
	maxVolume    = 10
)

const unknownState = "unknown"

type dataPointWrite struct {
	dataPointID string
	value       any
}

func (write dataPointWrite) asPayload() map[string]any {
	return map[string]any{write.dataPointID: write.value}
}

var commandWrites = map[string]dataPointWrite{
	"start":  {dpWholeHouseCleaning, true},
	"stop":   {dpWholeHouseCleaning, false},
	"pause":  {dpPause, true},
	"resume": {dpPause, false},
	"dock":   {dpReturnToDock, true},
	"locate": {dpLocate, true},
}

var remoteControlDirections = map[string]bool{
	"forward":  true,
	"backward": true,
	"left":     true,
	"right":    true,
	"stop":     true,
}

var stateByDeviceValue = map[string]string{
	"idle":              "idle",
	"sweep":             "cleaning",
	"mop":               "cleaning",
	"totaling":          "cleaning",
	"pointing":          "cleaning",
	"areaing":           "cleaning",
	"curpointing":       "cleaning",
	"selectroom":        "cleaning",
	"pause":             "paused",
	"tocharge":          "returning",
	"chargring":         "charging",
	"fullcharge":        "charged",
	"dustcenterworking": "emptying",
	"dormant":           "sleeping",
	"remotectl":         "remote_control",
	"fault":             "fault",
}

type consumablePart struct {
	name        string
	dataPointID string
	resetValue  string
	maxSeconds  int
}

var consumableParts = []consumablePart{
	{name: "filter", dataPointID: dpFilterSeconds, resetValue: "resetfilter", maxSeconds: 540000},
	{name: "main_brush", dataPointID: dpMainBrushSeconds, resetValue: "resetmainbrush", maxSeconds: 1080000},
	{name: "side_brush", dataPointID: dpSideBrushSeconds, resetValue: "resetsidebrush", maxSeconds: 720000},
}

var allowedSuctionLevels = map[string]bool{"quiet": true, "auto": true, "strong": true, "max": true}

var allowedWaterLevels = map[string]bool{"low": true, "mid": true, "high": true}

var allowedAutoEmptyFrequencies = map[string]bool{"0": true, "1": true, "2": true, "3": true}
