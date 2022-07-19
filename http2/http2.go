package http2

// HTTP/2 Version Identification, RFC 7540 section 3.1
const (
	ProtocolTLS = "h2"
	ProtocolTCP = "h2c"
)

// section 3.5
const ClientPreface = "PRI * HTTP/2.0\\r\\n\\r\\nSM\\r\\n\\r\\n"

// section 7
type ErrCode uint32

const (
	ErrCodeNo                 ErrCode = 0x0
	ErrCodeProtocol           ErrCode = 0x1
	ErrCodeInternal           ErrCode = 0x2
	ErrCodeFlowControl        ErrCode = 0x3
	ErrCodeSettingsTimeout    ErrCode = 0x4
	ErrCodeStreamClosed       ErrCode = 0x5
	ErrCodeFrameSize          ErrCode = 0x6
	ErrCodeRefusedStream      ErrCode = 0x7
	ErrCodeCancel             ErrCode = 0x8
	ErrCodeCompression        ErrCode = 0x9
	ErrCodeConnect            ErrCode = 0xa
	ErrCodeEnhanceYourCalm    ErrCode = 0xb
	ErrCodeInadequateSecurity ErrCode = 0xc
	ErrCodeHTTP11Required     ErrCode = 0xd
)

// section 5.4.1
type ConnError struct {
	Err error
	ErrCode
}

// section 5.4.2
type StreamError struct {
	Err error
	ErrCode
	StreamID uint32
}

const (
	maxConcurrentStreams = 1<<31 - 1
	maxInitialWindowSize = 1<<31 - 1
)

type SettingID uint16

const (
	SettingHeaderTableSize      SettingID = 0x1
	SettingEnablePush           SettingID = 0x2
	SettingMaxConcurrentStreams SettingID = 0x3
	SettingInitialWindowSize    SettingID = 0x4
	SettingMaxFrameSize         SettingID = 0x5
	SettingMaxHeaderListSize    SettingID = 0x6
)

type setting struct {
	ID    SettingID
	Value uint32
}

type Settings []setting

// Header is a collection of headers
type Header map[string][]string

// FrameType represents Frame Type Registry, defined in RFC 7540 section 11.2.
type FrameType uint8

const (
	FrameData         FrameType = 0x0
	FrameHeaders      FrameType = 0x1
	FramePriority     FrameType = 0x2
	FrameRSTStream    FrameType = 0x3
	FrameSettings     FrameType = 0x4
	FramePushPromise  FrameType = 0x5
	FramePing         FrameType = 0x6
	FrameGoAway       FrameType = 0x7
	FrameWindowUpdate FrameType = 0x8
	FrameContinuation FrameType = 0x9
)

// Flags is An 8-bit field reserved for boolean flags specific to the frame type.
type Flags uint8

const (
	FlagEndStream  Flags = 0x1
	FlagEndHeaders Flags = 0x4
	FlagAck        Flags = 0x1
	FlagPadded     Flags = 0x8
	FlagPriority   Flags = 0x20
)

// sectuin 6.5
type SettingsFrame struct {
	Ack bool
	Settings
}

// section 6.6
type PushPromiseFrame struct {
	StreamID         uint32
	PromisedStreamID uint32
	Header
	PadLen uint8
}

// section 6.7
type PingFrame struct {
	Ack  bool
	Data [8]byte
}

// section 6.8
type GoAwayFrame struct {
	LastStreamID uint32
	ErrCode
	DebugData []byte
}

// section 6.9
type WindowUpdateFrame struct {
	StreamID            uint32
	WindowSizeIncrement uint32
}
