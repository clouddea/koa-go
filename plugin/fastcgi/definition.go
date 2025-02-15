package fastcgi

/*
 * Listening socket file number
 */
const FCGI_LISTENSOCK_FILENO = 0

type FCGI_Header struct {
	Version         uint8
	Type            uint8
	RequestIdB1     uint8
	RequestIdB0     uint8
	ContentLengthB1 uint8
	ContentLengthB0 uint8
	PaddingLength   uint8
	Reserved        uint8
}

/*
 * Number of bytes in a FCGI_Header.  Future versions of the protocol
 * will not reduce this number.
 */
const FCGI_HEADER_LEN = 8

/*
 * Value for version component of FCGI_Header
 */
const FCGI_VERSION_1 = 1

/*
 * Values for type component of FCGI_Header
 */
const FCGI_BEGIN_REQUEST = 1
const FCGI_ABORT_REQUEST = 2
const FCGI_END_REQUEST = 3
const FCGI_PARAMS = 4
const FCGI_STDIN = 5
const FCGI_STDOUT = 6
const FCGI_STDERR = 7
const FCGI_DATA = 8
const FCGI_GET_VALUES = 9
const FCGI_GET_VALUES_RESULT = 10
const FCGI_UNKNOWN_TYPE = 11
const FCGI_MAXTYPE = (FCGI_UNKNOWN_TYPE)

/*
 * Value for requestId component of FCGI_Header
 */
const FCGI_NULL_REQUEST_ID = 0

type FCGI_BeginRequestBody struct {
	RoleB1   uint8
	RoleB0   uint8
	Flags    uint8
	Reserved [5]uint8
}

type FCGI_BeginRequestRecord struct {
	Header FCGI_Header
	Body   FCGI_BeginRequestBody
}

/*
 * Mask for flags component of FCGI_BeginRequestBody
 */
const FCGI_KEEP_CONN = 1

/*
 * Values for role component of FCGI_BeginRequestBody
 */
const FCGI_RESPONDER = 1
const FCGI_AUTHORIZER = 2
const FCGI_FILTER = 3

type FCGI_EndRequestBody struct {
	AppStatusB3    uint8
	AppStatusB2    uint8
	AppStatusB1    uint8
	AppStatusB0    uint8
	ProtocolStatus uint8
	Reserved       [3]uint8
}

type FCGI_EndRequestRecord struct {
	Header FCGI_Header
	Body   FCGI_EndRequestBody
}

/*
 * Values for protocolStatus component of FCGI_EndRequestBody
 */
const FCGI_REQUEST_COMPLETE = 0
const FCGI_CANT_MPX_CONN = 1
const FCGI_OVERLOADED = 2
const FCGI_UNKNOWN_ROLE = 3

/*
 * Variable names for FCGI_GET_VALUES / FCGI_GET_VALUES_RESULT records
 */
const FCGI_MAX_CONNS = "FCGI_MAX_CONNS"
const FCGI_MAX_REQS = "FCGI_MAX_REQS"
const FCGI_MPXS_CONNS = "FCGI_MPXS_CONNS"

type FCGI_UnknownTypeBody struct {
	Type     uint8
	Reserved [7]uint8
}

type FCGI_UnknownTypeRecord struct {
	Header FCGI_Header
	Body   FCGI_UnknownTypeBody
}

type FCGI_NameValuePair11 struct {
	NameLength  uint8 // 最高位为0
	ValueLength uint8 // 最高位为0
	NameData    []byte
	ValueData   []byte
}

type FCGI_NameValuePair14 struct {
	NameLength  uint8  // 最高位为0
	ValueLength uint32 // 最高位为1
	NameData    []byte
	ValueData   []byte
}

type FCGI_NameValuePair41 struct {
	NameLength  uint32 // 最高位为1
	ValueLength uint8  // 最高位为0
	NameData    []byte
	ValueData   []byte
}

type FCGI_NameValuePair44 struct {
	NameLength  uint32 // 最高位为1
	ValueLength uint32 // 最高位为1
	NameData    []byte
	ValueData   []byte
}
