package socket // import "go.mau.fi/whatsmeow/socket"

Package socket implements a subset of the Noise protocol framework on top of
websockets as used by WhatsApp.

There shouldn't be any need to manually interact with this package. The Client
struct in the top-level whatsmeow package handles everything.

CONSTANTS

const (
	// Origin is the Origin header for all WhatsApp websocket connections
	Origin = "https://web.whatsapp.com"
	// URL is the websocket URL for the new multidevice protocol
	URL = "wss://web.whatsapp.com/ws/chat"
)
const (
	NoiseStartPattern = "Noise_XX_25519_AESGCM_SHA256\x00\x00\x00\x00"

	WAMagicValue = 6
)
const (
	FrameMaxSize    = 1 << 24
	FrameLengthSize = 3
)

VARIABLES

var (
	ErrFrameTooLarge     = errors.New("frame too large")
	ErrSocketClosed      = errors.New("frame socket is closed")
	ErrSocketAlreadyOpen = errors.New("frame socket is already open")
)
var WAConnHeader = []byte{'W', 'A', WAMagicValue, token.DictVersion}

TYPES

type DisconnectHandler func(ctx context.Context, socket *NoiseSocket, remote bool)

type FrameHandler func(context.Context, []byte)

type FrameSocket struct {
	URL         string
	HTTPHeaders http.Header
	HTTPClient  *http.Client

	Frames       chan []byte
	OnDisconnect func(ctx context.Context, remote bool)

	Header []byte

	// Has unexported fields.
}

func NewFrameSocket(log waLog.Logger, client *http.Client) *FrameSocket

func (fs *FrameSocket) Close(code websocket.StatusCode)

func (fs *FrameSocket) Connect(ctx context.Context) error

func (fs *FrameSocket) Context() context.Context

func (fs *FrameSocket) IsConnected() bool

func (fs *FrameSocket) SendFrame(data []byte) error

type NoiseHandshake struct {
	// Has unexported fields.
}

func NewNoiseHandshake() *NoiseHandshake

func (nh *NoiseHandshake) Authenticate(data []byte)

func (nh *NoiseHandshake) Decrypt(ciphertext []byte) (plaintext []byte, err error)

func (nh *NoiseHandshake) Encrypt(plaintext []byte) []byte

func (nh *NoiseHandshake) Finish(
	ctx context.Context,
	fs *FrameSocket,
	frameHandler FrameHandler,
	disconnectHandler DisconnectHandler,
) (*NoiseSocket, error)

func (nh *NoiseHandshake) MixIntoKey(data []byte) error

func (nh *NoiseHandshake) MixSharedSecretIntoKey(priv, pub [32]byte) error

func (nh *NoiseHandshake) Start(pattern string, header []byte)

type NoiseSocket struct {
	// Has unexported fields.
}

func (ns *NoiseSocket) IsConnected() bool

func (ns *NoiseSocket) SendFrame(ctx context.Context, plaintext []byte) error

func (ns *NoiseSocket) Stop(disconnect bool)

