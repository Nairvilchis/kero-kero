package binary // import "go.mau.fi/whatsmeow/binary"

Package binary implements encoding and decoding documents in WhatsApp's binary
XML format.

VARIABLES

var (
	ErrInvalidType    = errors.New("unsupported payload type")
	ErrInvalidJIDType = errors.New("invalid JID type")
	ErrInvalidNode    = errors.New("invalid node")
	ErrInvalidToken   = errors.New("invalid token with tag")
	ErrNonStringKey   = errors.New("non-string key")
)
    Errors returned by the binary XML decoder.

var (
	IndentXML            = false
	MaxBytesToPrintAsHex = 128
)
    Options to control how Node.XMLString behaves.


FUNCTIONS

func Marshal(n Node) ([]byte, error)
    Marshal encodes an XML element (Node) into WhatsApp's binary XML
    representation.

func Unpack(data []byte) ([]byte, error)
    Unpack unpacks the given decrypted data from the WhatsApp web API.

    It checks the first byte to decide whether to uncompress the data with
    zlib or just return as-is (without the first byte). There's currently no
    corresponding Pack function because Marshal already returns the data with a
    leading zero (i.e. not compressed).


TYPES

type AttrUtility struct {
	Attrs  Attrs
	Errors []error
}
    AttrUtility is a helper struct for reading multiple XML attributes and
    checking for errors afterwards.

    The functions return values directly and append any decoding errors to the
    Errors slice. The slice can then be checked after all necessary attributes
    are read, instead of having to check each attribute for errors separately.

func (au *AttrUtility) Bool(key string) bool

func (au *AttrUtility) Error() error
    Error returns the list of errors as a single error interface, or nil if
    there are no errors.

func (au *AttrUtility) GetBool(key string, require bool) (bool, bool)

func (au *AttrUtility) GetInt64(key string, require bool) (int64, bool)

func (au *AttrUtility) GetJID(key string, require bool) (jidVal types.JID, ok bool)

func (au *AttrUtility) GetString(key string, require bool) (strVal string, ok bool)

func (au *AttrUtility) GetUint64(key string, require bool) (uint64, bool)

func (au *AttrUtility) GetUnixMilli(key string, require bool) (time.Time, bool)

func (au *AttrUtility) GetUnixTime(key string, require bool) (time.Time, bool)

func (au *AttrUtility) Int(key string) int

func (au *AttrUtility) Int64(key string) int64

func (au *AttrUtility) JID(key string) types.JID
    JID returns the JID under the given key. If there's no valid JID under the
    given key, an error will be stored and a blank JID struct will be returned.

func (au *AttrUtility) OK() bool
    OK returns true if there are no errors.

func (au *AttrUtility) OptionalBool(key string) bool

func (au *AttrUtility) OptionalInt(key string) int

func (au *AttrUtility) OptionalJID(key string) *types.JID
    OptionalJID returns the JID under the given key. If there's no valid JID
    under the given key, this will return nil. However, if the attribute is
    completely missing, this will not store an error.

func (au *AttrUtility) OptionalJIDOrEmpty(key string) types.JID
    OptionalJIDOrEmpty returns the JID under the given key. If there's no
    valid JID under the given key, this will return an empty JID. However,
    if the attribute is completely missing, this will not store an error.

func (au *AttrUtility) OptionalString(key string) string
    OptionalString returns the string under the given key.

func (au *AttrUtility) OptionalUnixMilli(key string) time.Time

func (au *AttrUtility) OptionalUnixTime(key string) time.Time

func (au *AttrUtility) String(key string) string
    String returns the string under the given key. If there's no valid string
    under the given key, an error will be stored and an empty string will be
    returned.

func (au *AttrUtility) Uint64(key string) uint64

func (au *AttrUtility) UnixMilli(key string) time.Time

func (au *AttrUtility) UnixTime(key string) time.Time

type Attrs = map[string]any
    Attrs is a type alias for the attributes of an XML element (Node).

type ErrorList []error
    ErrorList is a list of errors that implements the error interface itself.

func (el ErrorList) Error() string
    Error returns all the errors in the list as a string.

type Node struct {
	Tag     string      // The tag of the element.
	Attrs   Attrs       // The attributes of the element.
	Content interface{} // The content inside the element. Can be nil, a list of Nodes, or a byte array.
}
    Node represents an XML element.

func Unmarshal(data []byte) (*Node, error)
    Unmarshal decodes WhatsApp's binary XML representation into a Node.

func (n *Node) AttrGetter() *AttrUtility
    AttrGetter returns the AttrUtility for this Node.

func (n *Node) GetChildByTag(tags ...string) Node
    GetChildByTag does the same thing as GetOptionalChildByTag, but returns the
    Node directly without the ok boolean.

func (n *Node) GetChildren() []Node
    GetChildren returns the Content of the node as a list of nodes. If the
    content is not a list of nodes, this returns nil.

func (n *Node) GetChildrenByTag(tag string) (children []Node)
    GetChildrenByTag returns the same list as GetChildren, but filters it by tag
    first.

func (n *Node) GetOptionalChildByTag(tags ...string) (val Node, ok bool)
    GetOptionalChildByTag finds the first child with the given tag and returns
    it. Each provided tag will recurse in, so this is useful for getting a
    specific nested element.

func (n *Node) UnmarshalJSON(data []byte) error

func (n *Node) XMLString() string
    XMLString converts the Node to its XML representation

