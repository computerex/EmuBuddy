package wiiu

// TitleEntry represents a Wii U title entry
type TitleEntry struct {
	TitleID  uint64
	Name     string
	Region   uint8
	Category uint8
}

// Region constants
const (
	MCP_REGION_JAPAN  uint8 = 0x01
	MCP_REGION_USA    uint8 = 0x02
	MCP_REGION_EUROPE uint8 = 0x04
	MCP_REGION_CHINA  uint8 = 0x10
	MCP_REGION_KOREA  uint8 = 0x20
	MCP_REGION_TAIWAN uint8 = 0x40
)

// Title category constants
const (
	TITLE_CATEGORY_GAME uint8 = iota
	TITLE_CATEGORY_UPDATE
	TITLE_CATEGORY_DLC
	TITLE_CATEGORY_DEMO
	TITLE_CATEGORY_ALL
	TITLE_CATEGORY_DISC
)

// Title ID high constants
const (
	TID_HIGH_GAME            uint32 = 0x00050000
	TID_HIGH_DEMO            uint32 = 0x00050002
	TID_HIGH_SYSTEM_APP      uint32 = 0x00050010
	TID_HIGH_SYSTEM_DATA     uint32 = 0x0005001B
	TID_HIGH_SYSTEM_APPLET   uint32 = 0x00050030
	TID_HIGH_VWII_IOS        uint32 = 0x00000007
	TID_HIGH_VWII_SYSTEM_APP uint32 = 0x00070002
	TID_HIGH_VWII_SYSTEM     uint32 = 0x00070008
	TID_HIGH_DLC             uint32 = 0x0005000C
	TID_HIGH_UPDATE          uint32 = 0x0005000E
)

// GetFormattedRegion returns a human-readable region string
func GetFormattedRegion(region uint8) string {
	if region&MCP_REGION_EUROPE != 0 {
		if region&MCP_REGION_USA != 0 {
			if region&MCP_REGION_JAPAN != 0 {
				return "All"
			}
			return "USA/Europe"
		}
		if region&MCP_REGION_JAPAN != 0 {
			return "Europe/Japan"
		}
		return "Europe"
	}
	if region&MCP_REGION_USA != 0 {
		if region&MCP_REGION_JAPAN != 0 {
			return "USA/Japan"
		}
		return "USA"
	}
	if region&MCP_REGION_JAPAN != 0 {
		return "Japan"
	}
	return "Unknown"
}

// GetFormattedKind returns a human-readable title type
func GetFormattedKind(titleID uint64) string {
	switch uint32(titleID >> 32) {
	case TID_HIGH_GAME:
		return "Game"
	case TID_HIGH_DEMO:
		return "Demo"
	case TID_HIGH_SYSTEM_APP:
		return "System App"
	case TID_HIGH_SYSTEM_DATA:
		return "System Data"
	case TID_HIGH_SYSTEM_APPLET:
		return "System Applet"
	case TID_HIGH_VWII_IOS:
		return "vWii IOS"
	case TID_HIGH_VWII_SYSTEM_APP:
		return "vWii System App"
	case TID_HIGH_VWII_SYSTEM:
		return "vWii System"
	case TID_HIGH_DLC:
		return "DLC"
	case TID_HIGH_UPDATE:
		return "Update"
	default:
		return "Unknown"
	}
}
