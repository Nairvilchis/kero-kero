package types // import "go.mau.fi/whatsmeow/types"

Package types contains various structs and other types used by whatsmeow.

CONSTANTS

const (
	DefaultUserServer = "s.whatsapp.net"
	GroupServer       = "g.us"
	LegacyUserServer  = "c.us"
	BroadcastServer   = "broadcast"
	HiddenUserServer  = "lid"
	MessengerServer   = "msgr"
	InteropServer     = "interop"
	NewsletterServer  = "newsletter"
	HostedServer      = "hosted"
	HostedLIDServer   = "hosted.lid"
	BotServer         = "bot"
)
    Known JID servers on WhatsApp


VARIABLES

var (
	EmptyJID            = JID{}
	GroupServerJID      = NewJID("", GroupServer)
	ServerJID           = NewJID("", DefaultUserServer)
	BroadcastServerJID  = NewJID("", BroadcastServer)
	StatusBroadcastJID  = NewJID("status", BroadcastServer)
	LegacyPSAJID        = NewJID("0", LegacyUserServer)
	PSAJID              = NewJID("0", DefaultUserServer)
	OfficialBusinessJID = NewJID("16505361212", LegacyUserServer)
	MetaAIJID           = NewJID("13135550002", DefaultUserServer)
	NewMetaAIJID        = NewJID("867051314767696", BotServer)
)
    Some JIDs that are contacted often.

var (
	WhatsAppDomain  = uint8(0)   // This is the main domain type that whatsapp uses
	LIDDomain       = uint8(1)   // This is the domain for LID type JIDs
	HostedDomain    = uint8(128) // This is the domain for Hosted type JIDs
	HostedLIDDomain = uint8(129) // This is the domain for Hosted LID type JIDs
)
var BotJIDMap = map[JID]JID{
	NewJID("867051314767696", BotServer):   NewJID("13135550002", DefaultUserServer),
	NewJID("1061492271844689", BotServer):  NewJID("13135550005", DefaultUserServer),
	NewJID("245886058483988", BotServer):   NewJID("13135550009", DefaultUserServer),
	NewJID("3509905702656130", BotServer):  NewJID("13135550012", DefaultUserServer),
	NewJID("1059680132034576", BotServer):  NewJID("13135550013", DefaultUserServer),
	NewJID("715681030623646", BotServer):   NewJID("13135550014", DefaultUserServer),
	NewJID("1644971366323052", BotServer):  NewJID("13135550015", DefaultUserServer),
	NewJID("582497970646566", BotServer):   NewJID("13135550019", DefaultUserServer),
	NewJID("645459357769306", BotServer):   NewJID("13135550022", DefaultUserServer),
	NewJID("294997126699143", BotServer):   NewJID("13135550023", DefaultUserServer),
	NewJID("1522631578502677", BotServer):  NewJID("13135550027", DefaultUserServer),
	NewJID("719421926276396", BotServer):   NewJID("13135550030", DefaultUserServer),
	NewJID("1788488635002167", BotServer):  NewJID("13135550031", DefaultUserServer),
	NewJID("24232338603080193", BotServer): NewJID("13135550033", DefaultUserServer),
	NewJID("689289903143209", BotServer):   NewJID("13135550035", DefaultUserServer),
	NewJID("871626054177096", BotServer):   NewJID("13135550039", DefaultUserServer),
	NewJID("362351902849370", BotServer):   NewJID("13135550042", DefaultUserServer),
	NewJID("1744617646041527", BotServer):  NewJID("13135550043", DefaultUserServer),
	NewJID("893887762270570", BotServer):   NewJID("13135550046", DefaultUserServer),
	NewJID("1155032702135830", BotServer):  NewJID("13135550047", DefaultUserServer),
	NewJID("333931965993883", BotServer):   NewJID("13135550048", DefaultUserServer),
	NewJID("853748013058752", BotServer):   NewJID("13135550049", DefaultUserServer),
	NewJID("1559068611564819", BotServer):  NewJID("13135550053", DefaultUserServer),
	NewJID("890487432705716", BotServer):   NewJID("13135550054", DefaultUserServer),
	NewJID("240254602395494", BotServer):   NewJID("13135550055", DefaultUserServer),
	NewJID("1578420349663261", BotServer):  NewJID("13135550062", DefaultUserServer),
	NewJID("322908887140421", BotServer):   NewJID("13135550065", DefaultUserServer),
	NewJID("3713961535514771", BotServer):  NewJID("13135550067", DefaultUserServer),
	NewJID("997884654811738", BotServer):   NewJID("13135550070", DefaultUserServer),
	NewJID("403157239387035", BotServer):   NewJID("13135550081", DefaultUserServer),
	NewJID("535242369074963", BotServer):   NewJID("13135550082", DefaultUserServer),
	NewJID("946293427247659", BotServer):   NewJID("13135550083", DefaultUserServer),
	NewJID("3664707673802291", BotServer):  NewJID("13135550084", DefaultUserServer),
	NewJID("1821827464894892", BotServer):  NewJID("13135550085", DefaultUserServer),
	NewJID("1760312477828757", BotServer):  NewJID("13135550086", DefaultUserServer),
	NewJID("439480398712216", BotServer):   NewJID("13135550087", DefaultUserServer),
	NewJID("1876735582800984", BotServer):  NewJID("13135550088", DefaultUserServer),
	NewJID("984025089825661", BotServer):   NewJID("13135550089", DefaultUserServer),
	NewJID("1001336351558186", BotServer):  NewJID("13135550090", DefaultUserServer),
	NewJID("3739346336347061", BotServer):  NewJID("13135550091", DefaultUserServer),
	NewJID("3632749426974980", BotServer):  NewJID("13135550092", DefaultUserServer),
	NewJID("427864203481615", BotServer):   NewJID("13135550093", DefaultUserServer),
	NewJID("1434734570493055", BotServer):  NewJID("13135550094", DefaultUserServer),
	NewJID("992873449225921", BotServer):   NewJID("13135550095", DefaultUserServer),
	NewJID("813087747426445", BotServer):   NewJID("13135550096", DefaultUserServer),
	NewJID("806369104931434", BotServer):   NewJID("13135550098", DefaultUserServer),
	NewJID("1220982902403148", BotServer):  NewJID("13135550099", DefaultUserServer),
	NewJID("1365893374104393", BotServer):  NewJID("13135550100", DefaultUserServer),
	NewJID("686482033622048", BotServer):   NewJID("13135550200", DefaultUserServer),
	NewJID("1454999838411253", BotServer):  NewJID("13135550201", DefaultUserServer),
	NewJID("718584497008509", BotServer):   NewJID("13135550202", DefaultUserServer),
	NewJID("743520384213443", BotServer):   NewJID("13135550301", DefaultUserServer),
	NewJID("1147715789823789", BotServer):  NewJID("13135550302", DefaultUserServer),
	NewJID("1173034540372201", BotServer):  NewJID("13135550303", DefaultUserServer),
	NewJID("974785541030953", BotServer):   NewJID("13135550304", DefaultUserServer),
	NewJID("1122200255531507", BotServer):  NewJID("13135550305", DefaultUserServer),
	NewJID("899669714813162", BotServer):   NewJID("13135550306", DefaultUserServer),
	NewJID("631880108970650", BotServer):   NewJID("13135550307", DefaultUserServer),
	NewJID("435816149330026", BotServer):   NewJID("13135550308", DefaultUserServer),
	NewJID("1368717161184556", BotServer):  NewJID("13135550309", DefaultUserServer),
	NewJID("7849963461784891", BotServer):  NewJID("13135550310", DefaultUserServer),
	NewJID("3609617065968984", BotServer):  NewJID("13135550312", DefaultUserServer),
	NewJID("356273980574602", BotServer):   NewJID("13135550313", DefaultUserServer),
	NewJID("1043447920539760", BotServer):  NewJID("13135550314", DefaultUserServer),
	NewJID("1052764336525346", BotServer):  NewJID("13135550315", DefaultUserServer),
	NewJID("2631118843732685", BotServer):  NewJID("13135550316", DefaultUserServer),
	NewJID("510505411332176", BotServer):   NewJID("13135550317", DefaultUserServer),
	NewJID("1945664239227513", BotServer):  NewJID("13135550318", DefaultUserServer),
	NewJID("1518594378764656", BotServer):  NewJID("13135550319", DefaultUserServer),
	NewJID("1378821579456138", BotServer):  NewJID("13135550320", DefaultUserServer),
	NewJID("490214716896013", BotServer):   NewJID("13135550321", DefaultUserServer),
	NewJID("1028577858870699", BotServer):  NewJID("13135550322", DefaultUserServer),
	NewJID("308915665545959", BotServer):   NewJID("13135550323", DefaultUserServer),
	NewJID("845884253678900", BotServer):   NewJID("13135550324", DefaultUserServer),
	NewJID("995031308616442", BotServer):   NewJID("13135550325", DefaultUserServer),
	NewJID("2787365464763437", BotServer):  NewJID("13135550326", DefaultUserServer),
	NewJID("1532790990671645", BotServer):  NewJID("13135550327", DefaultUserServer),
	NewJID("302617036180485", BotServer):   NewJID("13135550328", DefaultUserServer),
	NewJID("723376723197227", BotServer):   NewJID("13135550329", DefaultUserServer),
	NewJID("8393570407377966", BotServer):  NewJID("13135550330", DefaultUserServer),
	NewJID("1931159970680725", BotServer):  NewJID("13135550331", DefaultUserServer),
	NewJID("401073885688605", BotServer):   NewJID("13135550332", DefaultUserServer),
	NewJID("2234478453565422", BotServer):  NewJID("13135550334", DefaultUserServer),
	NewJID("814748673882312", BotServer):   NewJID("13135550335", DefaultUserServer),
	NewJID("26133635056281592", BotServer): NewJID("13135550336", DefaultUserServer),
	NewJID("1439804456676119", BotServer):  NewJID("13135550337", DefaultUserServer),
	NewJID("889851503172161", BotServer):   NewJID("13135550338", DefaultUserServer),
	NewJID("1018283232836879", BotServer):  NewJID("13135550339", DefaultUserServer),
	NewJID("1012781386779537", BotServer):  NewJID("13135559000", DefaultUserServer),
	NewJID("823280953239532", BotServer):   NewJID("13135559001", DefaultUserServer),
	NewJID("1597090934573334", BotServer):  NewJID("13135559002", DefaultUserServer),
	NewJID("485965054020343", BotServer):   NewJID("13135559003", DefaultUserServer),
	NewJID("1033381648363446", BotServer):  NewJID("13135559004", DefaultUserServer),
	NewJID("491802010206446", BotServer):   NewJID("13135559005", DefaultUserServer),
	NewJID("1017139033184870", BotServer):  NewJID("13135559006", DefaultUserServer),
	NewJID("499638325922174", BotServer):   NewJID("13135559008", DefaultUserServer),
	NewJID("468946335863664", BotServer):   NewJID("13135559009", DefaultUserServer),
	NewJID("1570389776875816", BotServer):  NewJID("13135559010", DefaultUserServer),
	NewJID("1004342694328995", BotServer):  NewJID("13135559011", DefaultUserServer),
	NewJID("1012240323971229", BotServer):  NewJID("13135559012", DefaultUserServer),
	NewJID("392171787222419", BotServer):   NewJID("13135559013", DefaultUserServer),
	NewJID("952081212945019", BotServer):   NewJID("13135559016", DefaultUserServer),
	NewJID("444507875070178", BotServer):   NewJID("13135559017", DefaultUserServer),
	NewJID("1274819440594668", BotServer):  NewJID("13135559018", DefaultUserServer),
	NewJID("1397041101147050", BotServer):  NewJID("13135559019", DefaultUserServer),
	NewJID("425657699872640", BotServer):   NewJID("13135559020", DefaultUserServer),
	NewJID("532292852562549", BotServer):   NewJID("13135559021", DefaultUserServer),
	NewJID("705863241720292", BotServer):   NewJID("13135559022", DefaultUserServer),
	NewJID("476449815183959", BotServer):   NewJID("13135559023", DefaultUserServer),
	NewJID("488071553854222", BotServer):   NewJID("13135559024", DefaultUserServer),
	NewJID("468693832665397", BotServer):   NewJID("13135559025", DefaultUserServer),
	NewJID("517422564037340", BotServer):   NewJID("13135559026", DefaultUserServer),
	NewJID("819805466613825", BotServer):   NewJID("13135559027", DefaultUserServer),
	NewJID("1847708235641382", BotServer):  NewJID("13135559028", DefaultUserServer),
	NewJID("716282970644228", BotServer):   NewJID("13135559029", DefaultUserServer),
	NewJID("521655380527741", BotServer):   NewJID("13135559030", DefaultUserServer),
	NewJID("476193631941905", BotServer):   NewJID("13135559031", DefaultUserServer),
	NewJID("485600497445562", BotServer):   NewJID("13135559032", DefaultUserServer),
	NewJID("440217235683910", BotServer):   NewJID("13135559033", DefaultUserServer),
	NewJID("523342446758478", BotServer):   NewJID("13135559034", DefaultUserServer),
	NewJID("514784864360240", BotServer):   NewJID("13135559035", DefaultUserServer),
	NewJID("505790121814530", BotServer):   NewJID("13135559036", DefaultUserServer),
	NewJID("420008964419580", BotServer):   NewJID("13135559037", DefaultUserServer),
	NewJID("492141680204555", BotServer):   NewJID("13135559038", DefaultUserServer),
	NewJID("388462787271952", BotServer):   NewJID("13135559039", DefaultUserServer),
	NewJID("423473920752072", BotServer):   NewJID("13135559040", DefaultUserServer),
	NewJID("489574180468229", BotServer):   NewJID("13135559041", DefaultUserServer),
	NewJID("432360635854105", BotServer):   NewJID("13135559042", DefaultUserServer),
	NewJID("477878201669248", BotServer):   NewJID("13135559043", DefaultUserServer),
	NewJID("351656951234045", BotServer):   NewJID("13135559044", DefaultUserServer),
	NewJID("430178036732582", BotServer):   NewJID("13135559045", DefaultUserServer),
	NewJID("434537312944552", BotServer):   NewJID("13135559046", DefaultUserServer),
	NewJID("1240614300631808", BotServer):  NewJID("13135559047", DefaultUserServer),
	NewJID("473135945605128", BotServer):   NewJID("13135559048", DefaultUserServer),
	NewJID("423669800729310", BotServer):   NewJID("13135559049", DefaultUserServer),
	NewJID("3685666705015792", BotServer):  NewJID("13135559050", DefaultUserServer),
	NewJID("504196509016638", BotServer):   NewJID("13135559051", DefaultUserServer),
	NewJID("346844785189449", BotServer):   NewJID("13135559052", DefaultUserServer),
	NewJID("504823088911074", BotServer):   NewJID("13135559053", DefaultUserServer),
	NewJID("402669415797083", BotServer):   NewJID("13135559054", DefaultUserServer),
	NewJID("490939640234431", BotServer):   NewJID("13135559055", DefaultUserServer),
	NewJID("875124128063715", BotServer):   NewJID("13135559056", DefaultUserServer),
	NewJID("468788962654605", BotServer):   NewJID("13135559057", DefaultUserServer),
	NewJID("562386196354570", BotServer):   NewJID("13135559058", DefaultUserServer),
	NewJID("372159285928791", BotServer):   NewJID("13135559059", DefaultUserServer),
	NewJID("531017479591050", BotServer):   NewJID("13135559060", DefaultUserServer),
	NewJID("1328873881401826", BotServer):  NewJID("13135559061", DefaultUserServer),
	NewJID("1608363646390484", BotServer):  NewJID("13135559062", DefaultUserServer),
	NewJID("1229628561554232", BotServer):  NewJID("13135559063", DefaultUserServer),
	NewJID("348802211530364", BotServer):   NewJID("13135559064", DefaultUserServer),
	NewJID("3708535859420184", BotServer):  NewJID("13135559065", DefaultUserServer),
	NewJID("415517767742187", BotServer):   NewJID("13135559066", DefaultUserServer),
	NewJID("479330341612638", BotServer):   NewJID("13135559067", DefaultUserServer),
	NewJID("480785414723083", BotServer):   NewJID("13135559068", DefaultUserServer),
	NewJID("387299107507991", BotServer):   NewJID("13135559069", DefaultUserServer),
	NewJID("333389813188944", BotServer):   NewJID("13135559070", DefaultUserServer),
	NewJID("391794130316996", BotServer):   NewJID("13135559071", DefaultUserServer),
	NewJID("457893470576314", BotServer):   NewJID("13135559072", DefaultUserServer),
	NewJID("435550496166469", BotServer):   NewJID("13135559073", DefaultUserServer),
	NewJID("1620162702100689", BotServer):  NewJID("13135559074", DefaultUserServer),
	NewJID("867491058616043", BotServer):   NewJID("13135559075", DefaultUserServer),
	NewJID("816224117357759", BotServer):   NewJID("13135559076", DefaultUserServer),
	NewJID("334065176362830", BotServer):   NewJID("13135559077", DefaultUserServer),
	NewJID("489973170554709", BotServer):   NewJID("13135559078", DefaultUserServer),
	NewJID("473060669049665", BotServer):   NewJID("13135559079", DefaultUserServer),
	NewJID("1221505815643060", BotServer):  NewJID("13135559080", DefaultUserServer),
	NewJID("889000703096359", BotServer):   NewJID("13135559081", DefaultUserServer),
	NewJID("475235961979883", BotServer):   NewJID("13135559082", DefaultUserServer),
	NewJID("3434445653519934", BotServer):  NewJID("13135559084", DefaultUserServer),
	NewJID("524503026827421", BotServer):   NewJID("13135559085", DefaultUserServer),
	NewJID("1179639046403856", BotServer):  NewJID("13135559086", DefaultUserServer),
	NewJID("471563305859144", BotServer):   NewJID("13135559087", DefaultUserServer),
	NewJID("533896609192881", BotServer):   NewJID("13135559088", DefaultUserServer),
	NewJID("365443583168041", BotServer):   NewJID("13135559089", DefaultUserServer),
	NewJID("836082305329393", BotServer):   NewJID("13135559090", DefaultUserServer),
	NewJID("1056787705969916", BotServer):  NewJID("13135559091", DefaultUserServer),
	NewJID("503312598958357", BotServer):   NewJID("13135559092", DefaultUserServer),
	NewJID("3718606738453460", BotServer):  NewJID("13135559093", DefaultUserServer),
	NewJID("826066052850902", BotServer):   NewJID("13135559094", DefaultUserServer),
	NewJID("1033611345091888", BotServer):  NewJID("13135559095", DefaultUserServer),
	NewJID("3868390816783240", BotServer):  NewJID("13135559096", DefaultUserServer),
	NewJID("7462677740498860", BotServer):  NewJID("13135559097", DefaultUserServer),
	NewJID("436288576108573", BotServer):   NewJID("13135559098", DefaultUserServer),
	NewJID("1047559746718900", BotServer):  NewJID("13135559099", DefaultUserServer),
	NewJID("1099299455255491", BotServer):  NewJID("13135559100", DefaultUserServer),
	NewJID("1202037301040633", BotServer):  NewJID("13135559101", DefaultUserServer),
	NewJID("1720619402074074", BotServer):  NewJID("13135559102", DefaultUserServer),
	NewJID("1030422235101467", BotServer):  NewJID("13135559103", DefaultUserServer),
	NewJID("827238979523502", BotServer):   NewJID("13135559104", DefaultUserServer),
	NewJID("1516443722284921", BotServer):  NewJID("13135559105", DefaultUserServer),
	NewJID("1174442747196709", BotServer):  NewJID("13135559106", DefaultUserServer),
	NewJID("1653165225503842", BotServer):  NewJID("13135559107", DefaultUserServer),
	NewJID("1037648777635013", BotServer):  NewJID("13135559108", DefaultUserServer),
	NewJID("551617757299900", BotServer):   NewJID("13135559109", DefaultUserServer),
	NewJID("1158813558718726", BotServer):  NewJID("13135559110", DefaultUserServer),
	NewJID("2463236450542262", BotServer):  NewJID("13135559111", DefaultUserServer),
	NewJID("1550393252501466", BotServer):  NewJID("13135559112", DefaultUserServer),
	NewJID("2057065188042796", BotServer):  NewJID("13135559113", DefaultUserServer),
	NewJID("506163028760735", BotServer):   NewJID("13135559114", DefaultUserServer),
	NewJID("2065249100538481", BotServer):  NewJID("13135559115", DefaultUserServer),
	NewJID("1041382867195858", BotServer):  NewJID("13135559116", DefaultUserServer),
	NewJID("886500209499603", BotServer):   NewJID("13135559117", DefaultUserServer),
	NewJID("1491615624892655", BotServer):  NewJID("13135559118", DefaultUserServer),
	NewJID("486563697299617", BotServer):   NewJID("13135559119", DefaultUserServer),
	NewJID("1175736513679463", BotServer):  NewJID("13135559120", DefaultUserServer),
	NewJID("491811473512352", BotServer):   NewJID("13165550064", DefaultUserServer),
}

TYPES

type AddressingMode string

const (
	AddressingModePN  AddressingMode = "pn"
	AddressingModeLID AddressingMode = "lid"
)
type BasicCallMeta struct {
	From           JID
	Timestamp      time.Time
	CallCreator    JID
	CallCreatorAlt JID
	CallID         string
	GroupJID       JID
}

type Blocklist struct {
	DHash string // TODO is this just a timestamp?
	JIDs  []JID
}
    Blocklist contains the user's current list of blocked users.

type BotEditType string

const (
	EditTypeFirst BotEditType = "first"
	EditTypeInner BotEditType = "inner"
	EditTypeLast  BotEditType = "last"
)
type BotListInfo struct {
	BotJID    JID
	PersonaID string
}

type BotProfileCommand struct {
	Name        string
	Description string
}

type BotProfileInfo struct {
	JID                 JID
	Name                string
	Attributes          string
	Description         string
	Category            string
	IsDefault           bool
	Prompts             []string
	PersonaID           string
	Commands            []BotProfileCommand
	CommandsDescription string
}

type BroadcastRecipient struct {
	LID JID
	PN  JID
}

type BusinessHoursConfig struct {
	DayOfWeek string
	Mode      string
	OpenTime  string
	CloseTime string
}
    BusinessHoursConfig contains business operating hours of a WhatsApp
    business.

type BusinessMessageLinkTarget struct {
	JID JID // The JID of the business.

	PushName      string // The notify / push name of the business.
	VerifiedName  string // The verified business name.
	IsSigned      bool   // Some boolean, seems to be true?
	VerifiedLevel string // I guess the level of verification, starting from "unknown".

	Message string // The message that WhatsApp clients will pre-fill in the input box when clicking the link.
}
    BusinessMessageLinkTarget contains the info that is found using a business
    message link (see Client.ResolveBusinessMessageLink)

type BusinessProfile struct {
	JID                   JID
	Address               string
	Email                 string
	Categories            []Category
	ProfileOptions        map[string]string
	BusinessHoursTimeZone string
	BusinessHours         []BusinessHoursConfig
}
    BusinessProfile contains the profile information of a WhatsApp business.

type CallRemoteMeta struct {
	RemotePlatform string // The platform of the caller's WhatsApp client
	RemoteVersion  string // Version of the caller's WhatsApp client
}

type Category struct {
	ID   string
	Name string
}
    Category contains a WhatsApp business category.

type ChatPresence string

const (
	ChatPresenceComposing ChatPresence = "composing"
	ChatPresencePaused    ChatPresence = "paused"
)
type ChatPresenceMedia string

const (
	ChatPresenceMediaText  ChatPresenceMedia = ""
	ChatPresenceMediaAudio ChatPresenceMedia = "audio"
)
type ContactInfo struct {
	Found bool

	FirstName    string
	FullName     string
	PushName     string
	BusinessName string
	// Only for LID members encountered in groups, the phone number in the form "+1∙∙∙∙∙∙∙∙80"
	RedactedPhone string
}
    ContactInfo contains the cached names of a WhatsApp user.

type ContactQRLinkTarget struct {
	JID      JID    // The JID of the user.
	Type     string // Might always be "contact".
	PushName string // The notify / push name of the user.
}
    ContactQRLinkTarget contains the info that is found using a contact QR link
    (see Client.ResolveContactQRLink)

type DeviceSentMeta struct {
	DestinationJID string // The destination user. This should match the MessageInfo.Recipient field.
	Phash          string
}
    DeviceSentMeta contains metadata from messages sent by another one of the
    user's own devices.

type EditAttribute string

const (
	EditAttributeEmpty        EditAttribute = ""
	EditAttributeMessageEdit  EditAttribute = "1"
	EditAttributePinInChat    EditAttribute = "2"
	EditAttributeAdminEdit    EditAttribute = "3" // only used in newsletters
	EditAttributeSenderRevoke EditAttribute = "7"
	EditAttributeAdminRevoke  EditAttribute = "8"
)
type GraphQLError struct {
	Extensions GraphQLErrorExtensions `json:"extensions"`
	Message    string                 `json:"message"`
	Path       []string               `json:"path"`
}

func (gqle GraphQLError) Error() string

type GraphQLErrorExtensions struct {
	ErrorCode   int    `json:"error_code"`
	IsRetryable bool   `json:"is_retryable"`
	Severity    string `json:"severity"`
}

type GraphQLErrors []GraphQLError

func (gqles GraphQLErrors) Error() string

func (gqles GraphQLErrors) Unwrap() []error

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors GraphQLErrors   `json:"errors"`
}

type GroupAnnounce struct {
	IsAnnounce        bool
	AnnounceVersionID string
}
    GroupAnnounce specifies whether only admins can send messages in the group.

type GroupDelete struct {
	Deleted      bool
	DeleteReason string
}

type GroupEphemeral struct {
	IsEphemeral       bool
	DisappearingTimer uint32
}
    GroupEphemeral contains the group's disappearing messages settings.

type GroupIncognito struct {
	IsIncognito bool
}

type GroupInfo struct {
	JID      JID
	OwnerJID JID
	OwnerPN  JID

	GroupName
	GroupTopic
	GroupLocked
	GroupAnnounce
	GroupEphemeral
	GroupIncognito

	GroupParent
	GroupLinkedParent
	GroupIsDefaultSub
	GroupMembershipApprovalMode

	AddressingMode     AddressingMode
	GroupCreated       time.Time
	CreatorCountryCode string

	ParticipantVersionID string
	Participants         []GroupParticipant

	MemberAddMode GroupMemberAddMode
}
    GroupInfo contains basic information about a group chat on WhatsApp.

type GroupIsDefaultSub struct {
	IsDefaultSubGroup bool
}

type GroupLinkChange struct {
	Type         GroupLinkChangeType
	UnlinkReason GroupUnlinkReason
	Group        GroupLinkTarget
}

type GroupLinkChangeType string

const (
	GroupLinkChangeTypeParent  GroupLinkChangeType = "parent_group"
	GroupLinkChangeTypeSub     GroupLinkChangeType = "sub_group"
	GroupLinkChangeTypeSibling GroupLinkChangeType = "sibling_group"
)
type GroupLinkTarget struct {
	JID JID
	GroupName
	GroupIsDefaultSub
}

type GroupLinkedParent struct {
	LinkedParentJID JID
}

type GroupLocked struct {
	IsLocked bool
}
    GroupLocked specifies whether the group info can only be edited by admins.

type GroupMemberAddMode string

const (
	GroupMemberAddModeAdmin     GroupMemberAddMode = "admin_add"
	GroupMemberAddModeAllMember GroupMemberAddMode = "all_member_add"
)
type GroupMembershipApprovalMode struct {
	IsJoinApprovalRequired bool
}

type GroupName struct {
	Name        string
	NameSetAt   time.Time
	NameSetBy   JID
	NameSetByPN JID
}
    GroupName contains the name of a group along with metadata of who set it and
    when.

type GroupParent struct {
	IsParent                      bool
	DefaultMembershipApprovalMode string // request_required
}

type GroupParticipant struct {
	// The primary JID that should be used to send messages to this participant.
	// Always equals either the LID or phone number.
	JID         JID
	PhoneNumber JID
	LID         JID

	IsAdmin      bool
	IsSuperAdmin bool

	// This is only present for anonymous users in announcement groups, it's an obfuscated phone number
	DisplayName string

	// When creating groups, adding some participants may fail.
	// In such cases, the error code will be here.
	Error      int
	AddRequest *GroupParticipantAddRequest
}
    GroupParticipant contains info about a participant of a WhatsApp group chat.

type GroupParticipantAddRequest struct {
	Code       string
	Expiration time.Time
}

type GroupParticipantRequest struct {
	JID         JID
	RequestedAt time.Time
}

type GroupTopic struct {
	Topic        string
	TopicID      string
	TopicSetAt   time.Time
	TopicSetBy   JID
	TopicSetByPN JID
	TopicDeleted bool
}
    GroupTopic contains the topic (description) of a group along with metadata
    of who set it and when.

type GroupUnlinkReason string

const (
	GroupUnlinkReasonDefault GroupUnlinkReason = "unlink_group"
	GroupUnlinkReasonDelete  GroupUnlinkReason = "delete_parent"
)
type IsOnWhatsAppResponse struct {
	Query string // The query string used
	JID   JID    // The canonical user ID
	IsIn  bool   // Whether the phone is registered or not.

	VerifiedName *VerifiedName // If the phone is a business, the verified business details.
}
    IsOnWhatsAppResponse contains information received in response to checking
    if a phone number is on WhatsApp.

type JID struct {
	User       string
	RawAgent   uint8
	Device     uint16
	Integrator uint16
	Server     string
}
    JID represents a WhatsApp user ID.

    There are two types of JIDs: regular JID pairs (user and server) and AD-JIDs
    (user, agent and device). AD JIDs are only used to refer to specific devices
    of users, so the server is always s.whatsapp.net (DefaultUserServer).
    Regular JIDs can be used for entities on any servers (users, groups,
    broadcasts).

func NewADJID(user string, agent, device uint8) JID
    NewADJID creates a new AD JID.

func NewJID(user, server string) JID
    NewJID creates a new regular JID.

func ParseJID(jid string) (JID, error)
    ParseJID parses a JID out of the given string. It supports both regular and
    AD JIDs.

func (jid JID) ADString() string

func (jid JID) ActualAgent() uint8

func (jid JID) IsBot() bool

func (jid JID) IsBroadcastList() bool
    IsBroadcastList returns true if the JID is a broadcast list, but not the
    status broadcast.

func (jid JID) IsEmpty() bool
    IsEmpty returns true if the JID has no server (which is required for all
    JIDs).

func (jid JID) MarshalText() ([]byte, error)
    MarshalText implements encoding.TextMarshaler for JID

func (jid *JID) Scan(src interface{}) error
    Scan scans the given SQL value into this JID.

func (jid JID) SignalAddress() *signalProtocol.SignalAddress
    SignalAddress returns the Signal protocol address for the user.

func (jid JID) SignalAddressUser() string

func (jid JID) String() string
    String converts the JID to a string representation. The output string can be
    parsed with ParseJID.

func (jid JID) ToNonAD() JID
    ToNonAD returns a version of the JID struct that doesn't have the agent and
    device set.

func (jid *JID) UnmarshalText(val []byte) error
    UnmarshalText implements encoding.TextUnmarshaler for JID

func (jid JID) UserInt() uint64
    UserInt returns the user as an integer. This is only safe to run on normal
    users, not on groups or broadcast lists.

func (jid JID) Value() (driver.Value, error)
    Value returns the string representation of the JID as a value that the SQL
    package can use.

type LocalChatSettings struct {
	Found bool

	MutedUntil time.Time
	Pinned     bool
	Archived   bool
}
    LocalChatSettings contains the cached local settings for a chat.

type MessageID = string
    MessageID is the internal ID of a WhatsApp message.

type MessageInfo struct {
	MessageSource
	ID        MessageID
	ServerID  MessageServerID
	Type      string
	PushName  string
	Timestamp time.Time
	Category  string
	Multicast bool
	MediaType string
	Edit      EditAttribute

	MsgBotInfo  MsgBotInfo
	MsgMetaInfo MsgMetaInfo

	VerifiedName   *VerifiedName
	DeviceSentMeta *DeviceSentMeta // Metadata for direct messages sent from another one of the user's own devices.
}
    MessageInfo contains metadata about an incoming message.

type MessageServerID = int
    MessageServerID is the server ID of a WhatsApp newsletter message.

type MessageSource struct {
	Chat     JID  // The chat where the message was sent.
	Sender   JID  // The user who sent the message.
	IsFromMe bool // Whether the message was sent by the current user instead of someone else.
	IsGroup  bool // Whether the chat is a group chat or broadcast list.

	AddressingMode AddressingMode // The addressing mode of the message (phone number or LID)
	SenderAlt      JID            // The alternative address of the user who sent the message
	RecipientAlt   JID            // The alternative address of the recipient of the message for DMs.

	// When sending a read receipt to a broadcast list message, the Chat is the broadcast list
	// and Sender is you, so this field contains the recipient of the read receipt.
	BroadcastListOwner  JID
	BroadcastRecipients []BroadcastRecipient
}
    MessageSource contains basic sender and chat information about a message.

func (ms *MessageSource) IsIncomingBroadcast() bool
    IsIncomingBroadcast returns true if the message was sent to a broadcast list
    instead of directly to the user.

    If this is true, it means the message shows up in the direct chat with the
    Sender.

func (ms *MessageSource) SourceString() string
    SourceString returns a log-friendly representation of who sent the message
    and where.

type MsgBotInfo struct {
	EditType              BotEditType
	EditTargetID          MessageID
	EditSenderTimestampMS time.Time
}
    MsgBotInfo targets <bot>

type MsgMetaInfo struct {
	// Bot things
	TargetID     MessageID
	TargetSender JID
	TargetChat   JID

	DeprecatedLIDSession *bool

	ThreadMessageID        MessageID
	ThreadMessageSenderJID JID
}
    MsgMetaInfo targets <meta>

type NewsletterKeyType string

const (
	NewsletterKeyTypeJID    NewsletterKeyType = "JID"
	NewsletterKeyTypeInvite NewsletterKeyType = "INVITE"
)
type NewsletterMessage struct {
	MessageServerID MessageServerID
	MessageID       MessageID
	Type            string
	Timestamp       time.Time
	ViewsCount      int
	ReactionCounts  map[string]int

	// This is only present when fetching messages, not in live updates
	Message *waE2E.Message
}

type NewsletterMetadata struct {
	ID         JID                       `json:"id"`
	State      WrappedNewsletterState    `json:"state"`
	ThreadMeta NewsletterThreadMetadata  `json:"thread_metadata"`
	ViewerMeta *NewsletterViewerMetadata `json:"viewer_metadata"`
}

type NewsletterMuteState string

const (
	NewsletterMuteOn  NewsletterMuteState = "on"
	NewsletterMuteOff NewsletterMuteState = "off"
)
func (nms *NewsletterMuteState) UnmarshalText(text []byte) error

type NewsletterMuted struct {
	Muted bool
}

type NewsletterPrivacy string

const (
	NewsletterPrivacyPrivate NewsletterPrivacy = "private"
	NewsletterPrivacyPublic  NewsletterPrivacy = "public"
)
func (np *NewsletterPrivacy) UnmarshalText(text []byte) error

type NewsletterReactionSettings struct {
	Value NewsletterReactionsMode `json:"value"`
}

type NewsletterReactionsMode string

const (
	NewsletterReactionsModeAll       NewsletterReactionsMode = "all"
	NewsletterReactionsModeBasic     NewsletterReactionsMode = "basic"
	NewsletterReactionsModeNone      NewsletterReactionsMode = "none"
	NewsletterReactionsModeBlocklist NewsletterReactionsMode = "blocklist"
)
type NewsletterRole string

const (
	NewsletterRoleSubscriber NewsletterRole = "subscriber"
	NewsletterRoleGuest      NewsletterRole = "guest"
	NewsletterRoleAdmin      NewsletterRole = "admin"
	NewsletterRoleOwner      NewsletterRole = "owner"
)
func (nr *NewsletterRole) UnmarshalText(text []byte) error

type NewsletterSettings struct {
	ReactionCodes NewsletterReactionSettings `json:"reaction_codes"`
}

type NewsletterState string

const (
	NewsletterStateActive       NewsletterState = "active"
	NewsletterStateSuspended    NewsletterState = "suspended"
	NewsletterStateGeoSuspended NewsletterState = "geosuspended"
)
func (ns *NewsletterState) UnmarshalText(text []byte) error

type NewsletterText struct {
	Text       string                   `json:"text"`
	ID         string                   `json:"id"`
	UpdateTime jsontime.UnixMicroString `json:"update_time"`
}

type NewsletterThreadMetadata struct {
	CreationTime      jsontime.UnixString         `json:"creation_time"`
	InviteCode        string                      `json:"invite"`
	Name              NewsletterText              `json:"name"`
	Description       NewsletterText              `json:"description"`
	SubscriberCount   int                         `json:"subscribers_count,string"`
	VerificationState NewsletterVerificationState `json:"verification"`
	Picture           *ProfilePictureInfo         `json:"picture"`
	Preview           ProfilePictureInfo          `json:"preview"`
	Settings          NewsletterSettings          `json:"settings"`
}

type NewsletterVerificationState string

const (
	NewsletterVerificationStateVerified   NewsletterVerificationState = "verified"
	NewsletterVerificationStateUnverified NewsletterVerificationState = "unverified"
)
func (nvs *NewsletterVerificationState) UnmarshalText(text []byte) error

type NewsletterViewerMetadata struct {
	Mute NewsletterMuteState `json:"mute"`
	Role NewsletterRole      `json:"role"`
}

type Presence string

const (
	PresenceAvailable   Presence = "available"
	PresenceUnavailable Presence = "unavailable"
)
type PrivacySetting string
    PrivacySetting is an individual setting value in the user's privacy
    settings.

const (
	PrivacySettingUndefined        PrivacySetting = ""
	PrivacySettingAll              PrivacySetting = "all"
	PrivacySettingContacts         PrivacySetting = "contacts"
	PrivacySettingContactBlacklist PrivacySetting = "contact_blacklist"
	PrivacySettingMatchLastSeen    PrivacySetting = "match_last_seen"
	PrivacySettingKnown            PrivacySetting = "known"
	PrivacySettingNone             PrivacySetting = "none"
)
    Possible privacy setting values.

type PrivacySettingType string
    PrivacySettingType is the type of privacy setting.

const (
	PrivacySettingTypeGroupAdd     PrivacySettingType = "groupadd"     // Valid values: PrivacySettingAll, PrivacySettingContacts, PrivacySettingContactBlacklist, PrivacySettingNone
	PrivacySettingTypeLastSeen     PrivacySettingType = "last"         // Valid values: PrivacySettingAll, PrivacySettingContacts, PrivacySettingContactBlacklist, PrivacySettingNone
	PrivacySettingTypeStatus       PrivacySettingType = "status"       // Valid values: PrivacySettingAll, PrivacySettingContacts, PrivacySettingContactBlacklist, PrivacySettingNone
	PrivacySettingTypeProfile      PrivacySettingType = "profile"      // Valid values: PrivacySettingAll, PrivacySettingContacts, PrivacySettingContactBlacklist, PrivacySettingNone
	PrivacySettingTypeReadReceipts PrivacySettingType = "readreceipts" // Valid values: PrivacySettingAll, PrivacySettingNone
	PrivacySettingTypeOnline       PrivacySettingType = "online"       // Valid values: PrivacySettingAll, PrivacySettingMatchLastSeen
	PrivacySettingTypeCallAdd      PrivacySettingType = "calladd"      // Valid values: PrivacySettingAll, PrivacySettingKnown
)
type PrivacySettings struct {
	GroupAdd     PrivacySetting // Valid values: PrivacySettingAll, PrivacySettingContacts, PrivacySettingContactBlacklist, PrivacySettingNone
	LastSeen     PrivacySetting // Valid values: PrivacySettingAll, PrivacySettingContacts, PrivacySettingContactBlacklist, PrivacySettingNone
	Status       PrivacySetting // Valid values: PrivacySettingAll, PrivacySettingContacts, PrivacySettingContactBlacklist, PrivacySettingNone
	Profile      PrivacySetting // Valid values: PrivacySettingAll, PrivacySettingContacts, PrivacySettingContactBlacklist, PrivacySettingNone
	ReadReceipts PrivacySetting // Valid values: PrivacySettingAll, PrivacySettingNone
	CallAdd      PrivacySetting // Valid values: PrivacySettingAll, PrivacySettingKnown
	Online       PrivacySetting // Valid values: PrivacySettingAll, PrivacySettingMatchLastSeen
}
    PrivacySettings contains the user's privacy settings.

type ProfilePictureInfo struct {
	URL  string `json:"url"`  // The full URL for the image, can be downloaded with a simple HTTP request.
	ID   string `json:"id"`   // The ID of the image. This is the same as UserInfo.PictureID.
	Type string `json:"type"` // The type of image. Known types include "image" (full res) and "preview" (thumbnail).

	DirectPath string `json:"direct_path"` // The path to the image, probably not very useful

	Hash []byte `json:"hash"` // Some kind of hash (format is unknown)
}
    ProfilePictureInfo contains the ID and URL for a WhatsApp user's profile
    picture or group's photo.

type ReceiptType string
    ReceiptType represents the type of a Receipt event.

const (
	// ReceiptTypeDelivered means the message was delivered to the device (but the user might not have noticed).
	ReceiptTypeDelivered ReceiptType = ""
	// ReceiptTypeSender is sent by your other devices when a message you sent is delivered to them.
	ReceiptTypeSender ReceiptType = "sender"
	// ReceiptTypeRetry means the message was delivered to the device, but decrypting the message failed.
	ReceiptTypeRetry ReceiptType = "retry"
	// ReceiptTypeRead means the user opened the chat and saw the message.
	ReceiptTypeRead ReceiptType = "read"
	// ReceiptTypeReadSelf means the current user read a message from a different device, and has read receipts disabled in privacy settings.
	ReceiptTypeReadSelf ReceiptType = "read-self"
	// ReceiptTypePlayed means the user opened a view-once media message.
	//
	// This is dispatched for both incoming and outgoing messages when played. If the current user opened the media,
	// it means the media should be removed from all devices. If a recipient opened the media, it's just a notification
	// for the sender that the media was viewed.
	ReceiptTypePlayed ReceiptType = "played"
	// ReceiptTypePlayedSelf probably means the current user opened a view-once media message from a different device,
	// and has read receipts disabled in privacy settings.
	ReceiptTypePlayedSelf ReceiptType = "played-self"

	ReceiptTypeServerError ReceiptType = "server-error"
	ReceiptTypeInactive    ReceiptType = "inactive"
	ReceiptTypePeerMsg     ReceiptType = "peer_msg"
	ReceiptTypeHistorySync ReceiptType = "hist_sync"
)
func (rt ReceiptType) GoString() string
    GoString returns the name of the Go constant for the ReceiptType value.

type StatusPrivacy struct {
	Type StatusPrivacyType
	List []JID

	IsDefault bool
}
    StatusPrivacy contains the settings for who to send status messages to by
    default.

type StatusPrivacyType string
    StatusPrivacyType is the type of list in StatusPrivacy.

const (
	// StatusPrivacyTypeContacts means statuses are sent to all contacts.
	StatusPrivacyTypeContacts StatusPrivacyType = "contacts"
	// StatusPrivacyTypeBlacklist means statuses are sent to all contacts, except the ones on the list.
	StatusPrivacyTypeBlacklist StatusPrivacyType = "blacklist"
	// StatusPrivacyTypeWhitelist means statuses are only sent to users on the list.
	StatusPrivacyTypeWhitelist StatusPrivacyType = "whitelist"
)
type UserInfo struct {
	VerifiedName *VerifiedName
	Status       string
	PictureID    string
	Devices      []JID
	LID          JID
}
    UserInfo contains info about a WhatsApp user.

type VerifiedName struct {
	Certificate *waVnameCert.VerifiedNameCertificate
	Details     *waVnameCert.VerifiedNameCertificate_Details
}
    VerifiedName contains verified WhatsApp business details.

type WrappedNewsletterState struct {
	Type NewsletterState `json:"type"`
}

