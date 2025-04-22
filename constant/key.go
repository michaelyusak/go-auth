package constant

const (
	// Context key
	UserAgentCtxKey  = userAgentKey("user-agent")
	DeviceInfoCtxKey = deviceInfoKey("device-info")

	// Header key
	UserAgentHeaderKey  = "User-Agent"
	DeviceInfoHeaderKey = "Device-Info"
)

type userAgentKey string
type deviceInfoKey string
