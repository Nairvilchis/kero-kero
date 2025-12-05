package store // import "go.mau.fi/whatsmeow/store"

Package store contains interfaces for storing data needed for WhatsApp
multidevice.

VARIABLES

var BaseClientPayload = &waWa6.ClientPayload{
	UserAgent: &waWa6.ClientPayload_UserAgent{
		Platform:       waWa6.ClientPayload_UserAgent_WEB.Enum(),
		ReleaseChannel: waWa6.ClientPayload_UserAgent_RELEASE.Enum(),
		AppVersion:     waVersion.ProtoAppVersion(),
		Mcc:            proto.String("000"),
		Mnc:            proto.String("000"),
		OsVersion:      proto.String("0.1"),
		Manufacturer:   proto.String(""),
		Device:         proto.String("Desktop"),
		OsBuildNumber:  proto.String("0.1"),

		LocaleLanguageIso6391:       proto.String("en"),
		LocaleCountryIso31661Alpha2: proto.String("US"),
	},
	WebInfo: &waWa6.ClientPayload_WebInfo{
		WebSubPlatform: waWa6.ClientPayload_WebInfo_WEB_BROWSER.Enum(),
	},
	ConnectType:   waWa6.ClientPayload_WIFI_UNKNOWN.Enum(),
	ConnectReason: waWa6.ClientPayload_USER_ACTIVATED.Enum(),
}
var DeviceProps = &waCompanionReg.DeviceProps{
	Os: proto.String("whatsmeow"),
	Version: &waCompanionReg.DeviceProps_AppVersion{
		Primary:   proto.Uint32(0),
		Secondary: proto.Uint32(1),
		Tertiary:  proto.Uint32(0),
	},
	HistorySyncConfig: &waCompanionReg.DeviceProps_HistorySyncConfig{
		StorageQuotaMb:                           proto.Uint32(10240),
		InlineInitialPayloadInE2EeMsg:            proto.Bool(true),
		RecentSyncDaysLimit:                      nil,
		SupportCallLogHistory:                    proto.Bool(false),
		SupportBotUserAgentChatHistory:           proto.Bool(true),
		SupportCagReactionsAndPolls:              proto.Bool(true),
		SupportBizHostedMsg:                      proto.Bool(true),
		SupportRecentSyncChunkMessageCountTuning: proto.Bool(true),
		SupportHostedGroupMsg:                    proto.Bool(true),
		SupportFbidBotChatHistory:                proto.Bool(true),
		SupportAddOnHistorySyncMigration:         nil,
		SupportMessageAssociation:                proto.Bool(true),
		SupportGroupHistory:                      proto.Bool(false),
		OnDemandReady:                            nil,
		SupportGuestChat:                         nil,
		CompleteOnDemandReady:                    nil,
		ThumbnailSyncDaysLimit:                   nil,
	},
	PlatformType:    waCompanionReg.DeviceProps_UNKNOWN.Enum(),
	RequireFullSync: proto.Bool(false),
}
var MutedForever = time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
var NoopDevice = &Device{
	ID:          &types.EmptyJID,
	NoiseKey:    nilKey,
	IdentityKey: nilKey,

	Identities:    nilStore,
	Sessions:      nilStore,
	PreKeys:       nilStore,
	SenderKeys:    nilStore,
	AppStateKeys:  nilStore,
	AppState:      nilStore,
	Contacts:      nilStore,
	ChatSettings:  nilStore,
	MsgSecrets:    nilStore,
	PrivacyTokens: nilStore,
	EventBuffer:   nilStore,
	LIDs:          nilStore,
	Container:     nilStore,
}
var SignalProtobufSerializer = serialize.NewProtoBufSerializer()

FUNCTIONS

func SetOSInfo(name string, version [3]uint32)
func SetWAVersion(version WAVersionContainer)
    SetWAVersion sets the current WhatsApp web client version.

    In general, you should keep the library up-to-date instead of using this,
    as there may be code changes that are necessary too (like protobuf schema
    changes).


TYPES

type AllGlobalStores interface {
	LIDStore
}

type AllSessionSpecificStores interface {
	IdentityStore
	SessionStore
	PreKeyStore
	SenderKeyStore
	AppStateSyncKeyStore
	AppStateStore
	ContactStore
	ChatSettingsStore
	MsgSecretStore
	PrivacyTokenStore
	EventBuffer
}

type AllStores interface {
	AllSessionSpecificStores
	AllGlobalStores
}

type AppStateMutationMAC struct {
	IndexMAC []byte
	ValueMAC []byte
}

type AppStateStore interface {
	PutAppStateVersion(ctx context.Context, name string, version uint64, hash [128]byte) error
	GetAppStateVersion(ctx context.Context, name string) (uint64, [128]byte, error)
	DeleteAppStateVersion(ctx context.Context, name string) error

	PutAppStateMutationMACs(ctx context.Context, name string, version uint64, mutations []AppStateMutationMAC) error
	DeleteAppStateMutationMACs(ctx context.Context, name string, indexMACs [][]byte) error
	GetAppStateMutationMAC(ctx context.Context, name string, indexMAC []byte) (valueMAC []byte, err error)
}

type AppStateSyncKey struct {
	Data        []byte
	Fingerprint []byte
	Timestamp   int64
}

type AppStateSyncKeyStore interface {
	PutAppStateSyncKey(ctx context.Context, id []byte, key AppStateSyncKey) error
	GetAppStateSyncKey(ctx context.Context, id []byte) (*AppStateSyncKey, error)
	GetLatestAppStateSyncKeyID(ctx context.Context) ([]byte, error)
}

type BufferedEvent struct {
	Plaintext  []byte
	InsertTime time.Time
	ServerTime time.Time
}

type ChatSettingsStore interface {
	PutMutedUntil(ctx context.Context, chat types.JID, mutedUntil time.Time) error
	PutPinned(ctx context.Context, chat types.JID, pinned bool) error
	PutArchived(ctx context.Context, chat types.JID, archived bool) error
	GetChatSettings(ctx context.Context, chat types.JID) (types.LocalChatSettings, error)
}

type ContactEntry struct {
	JID       types.JID
	FirstName string
	FullName  string
}

func (ce ContactEntry) GetMassInsertValues() [3]any

type ContactStore interface {
	PutPushName(ctx context.Context, user types.JID, pushName string) (bool, string, error)
	PutBusinessName(ctx context.Context, user types.JID, businessName string) (bool, string, error)
	PutContactName(ctx context.Context, user types.JID, fullName, firstName string) error
	PutAllContactNames(ctx context.Context, contacts []ContactEntry) error
	PutManyRedactedPhones(ctx context.Context, entries []RedactedPhoneEntry) error
	GetContact(ctx context.Context, user types.JID) (types.ContactInfo, error)
	GetAllContacts(ctx context.Context) (map[types.JID]types.ContactInfo, error)
}

type Device struct {
	Log waLog.Logger

	NoiseKey       *keys.KeyPair
	IdentityKey    *keys.KeyPair
	SignedPreKey   *keys.PreKey
	RegistrationID uint32
	AdvSecretKey   []byte

	ID  *types.JID
	LID types.JID

	Account      *waAdv.ADVSignedDeviceIdentity
	Platform     string
	BusinessName string
	PushName     string

	LIDMigrationTimestamp int64

	FacebookUUID uuid.UUID

	Initialized   bool
	Identities    IdentityStore
	Sessions      SessionStore
	PreKeys       PreKeyStore
	SenderKeys    SenderKeyStore
	AppStateKeys  AppStateSyncKeyStore
	AppState      AppStateStore
	Contacts      ContactStore
	ChatSettings  ChatSettingsStore
	MsgSecrets    MsgSecretStore
	PrivacyTokens PrivacyTokenStore
	EventBuffer   EventBuffer
	LIDs          LIDStore
	Container     DeviceContainer
}

func (device *Device) ContainsPreKey(ctx context.Context, preKeyID uint32) (bool, error)

func (device *Device) ContainsSession(ctx context.Context, remoteAddress *protocol.SignalAddress) (bool, error)

func (device *Device) ContainsSignedPreKey(ctx context.Context, signedPreKeyID uint32) (bool, error)

func (device *Device) Delete(ctx context.Context) error

func (device *Device) DeleteAllSessions(ctx context.Context) error

func (device *Device) DeleteSession(ctx context.Context, remoteAddress *protocol.SignalAddress) error

func (device *Device) GetAltJID(ctx context.Context, jid types.JID) (types.JID, error)

func (device *Device) GetClientPayload() *waWa6.ClientPayload

func (device *Device) GetIdentityKeyPair() *identity.KeyPair

func (device *Device) GetJID() types.JID

func (device *Device) GetLID() types.JID

func (device *Device) GetLocalRegistrationID() uint32

func (device *Device) GetSubDeviceSessions(ctx context.Context, name string) ([]uint32, error)

func (device *Device) IsTrustedIdentity(ctx context.Context, address *protocol.SignalAddress, identityKey *identity.Key) (bool, error)

func (device *Device) LoadPreKey(ctx context.Context, id uint32) (*record.PreKey, error)

func (device *Device) LoadSenderKey(ctx context.Context, senderKeyName *protocol.SenderKeyName) (*groupRecord.SenderKey, error)

func (device *Device) LoadSession(ctx context.Context, address *protocol.SignalAddress) (*record.Session, error)

func (device *Device) LoadSignedPreKey(ctx context.Context, signedPreKeyID uint32) (*record.SignedPreKey, error)

func (device *Device) LoadSignedPreKeys(ctx context.Context) ([]*record.SignedPreKey, error)

func (device *Device) PutCachedSessions(ctx context.Context) error

func (device *Device) RemovePreKey(ctx context.Context, id uint32) error

func (device *Device) RemoveSignedPreKey(ctx context.Context, signedPreKeyID uint32) error

func (device *Device) Save(ctx context.Context) error

func (device *Device) SaveIdentity(ctx context.Context, address *protocol.SignalAddress, identityKey *identity.Key) error

func (device *Device) StorePreKey(ctx context.Context, preKeyID uint32, preKeyRecord *record.PreKey) error

func (device *Device) StoreSenderKey(ctx context.Context, senderKeyName *protocol.SenderKeyName, keyRecord *groupRecord.SenderKey) error

func (device *Device) StoreSession(ctx context.Context, address *protocol.SignalAddress, record *record.Session) error

func (device *Device) StoreSignedPreKey(ctx context.Context, signedPreKeyID uint32, record *record.SignedPreKey) error

func (device *Device) WithCachedSessions(ctx context.Context, addresses []string) (map[string]bool, context.Context, error)

type DeviceContainer interface {
	PutDevice(ctx context.Context, store *Device) error
	DeleteDevice(ctx context.Context, store *Device) error
}

type EventBuffer interface {
	GetBufferedEvent(ctx context.Context, ciphertextHash [32]byte) (*BufferedEvent, error)
	PutBufferedEvent(ctx context.Context, ciphertextHash [32]byte, plaintext []byte, serverTimestamp time.Time) error
	DoDecryptionTxn(ctx context.Context, fn func(context.Context) error) error
	ClearBufferedEventPlaintext(ctx context.Context, ciphertextHash [32]byte) error
	DeleteOldBufferedHashes(ctx context.Context) error
}

type IdentityStore interface {
	PutIdentity(ctx context.Context, address string, key [32]byte) error
	DeleteAllIdentities(ctx context.Context, phone string) error
	DeleteIdentity(ctx context.Context, address string) error
	IsTrustedIdentity(ctx context.Context, address string, key [32]byte) (bool, error)
}

type LIDMapping struct {
	LID types.JID
	PN  types.JID
}

func (lm LIDMapping) GetMassInsertValues() [2]any

type LIDStore interface {
	PutManyLIDMappings(ctx context.Context, mappings []LIDMapping) error
	PutLIDMapping(ctx context.Context, lid, jid types.JID) error
	GetPNForLID(ctx context.Context, lid types.JID) (types.JID, error)
	GetLIDForPN(ctx context.Context, pn types.JID) (types.JID, error)
	GetManyLIDsForPNs(ctx context.Context, pns []types.JID) (map[types.JID]types.JID, error)
}

type MessageSecretInsert struct {
	Chat   types.JID
	Sender types.JID
	ID     types.MessageID
	Secret []byte
}

type MsgSecretStore interface {
	PutMessageSecrets(ctx context.Context, inserts []MessageSecretInsert) error
	PutMessageSecret(ctx context.Context, chat, sender types.JID, id types.MessageID, secret []byte) error
	GetMessageSecret(ctx context.Context, chat, sender types.JID, id types.MessageID) ([]byte, types.JID, error)
}

type NoopStore struct {
	Error error
}

func (n *NoopStore) ClearBufferedEventPlaintext(ctx context.Context, ciphertextHash [32]byte) error

func (n *NoopStore) DeleteAllIdentities(ctx context.Context, phone string) error

func (n *NoopStore) DeleteAllSessions(ctx context.Context, phone string) error

func (n *NoopStore) DeleteAppStateMutationMACs(ctx context.Context, name string, indexMACs [][]byte) error

func (n *NoopStore) DeleteAppStateVersion(ctx context.Context, name string) error

func (n *NoopStore) DeleteDevice(ctx context.Context, store *Device) error

func (n *NoopStore) DeleteIdentity(ctx context.Context, address string) error

func (n *NoopStore) DeleteOldBufferedHashes(ctx context.Context) error

func (n *NoopStore) DeleteSession(ctx context.Context, address string) error

func (n *NoopStore) DoDecryptionTxn(ctx context.Context, fn func(context.Context) error) error

func (n *NoopStore) GenOnePreKey(ctx context.Context) (*keys.PreKey, error)

func (n *NoopStore) GetAllContacts(ctx context.Context) (map[types.JID]types.ContactInfo, error)

func (n *NoopStore) GetAppStateMutationMAC(ctx context.Context, name string, indexMAC []byte) (valueMAC []byte, err error)

func (n *NoopStore) GetAppStateSyncKey(ctx context.Context, id []byte) (*AppStateSyncKey, error)

func (n *NoopStore) GetAppStateVersion(ctx context.Context, name string) (uint64, [128]byte, error)

func (n *NoopStore) GetBufferedEvent(ctx context.Context, ciphertextHash [32]byte) (*BufferedEvent, error)

func (n *NoopStore) GetChatSettings(ctx context.Context, chat types.JID) (types.LocalChatSettings, error)

func (n *NoopStore) GetContact(ctx context.Context, user types.JID) (types.ContactInfo, error)

func (n *NoopStore) GetLIDForPN(ctx context.Context, pn types.JID) (types.JID, error)

func (n *NoopStore) GetLatestAppStateSyncKeyID(ctx context.Context) ([]byte, error)

func (n *NoopStore) GetManyLIDsForPNs(ctx context.Context, pns []types.JID) (map[types.JID]types.JID, error)

func (n *NoopStore) GetManySessions(ctx context.Context, addresses []string) (map[string][]byte, error)

func (n *NoopStore) GetMessageSecret(ctx context.Context, chat, sender types.JID, id types.MessageID) ([]byte, types.JID, error)

func (n *NoopStore) GetOrGenPreKeys(ctx context.Context, count uint32) ([]*keys.PreKey, error)

func (n *NoopStore) GetPNForLID(ctx context.Context, lid types.JID) (types.JID, error)

func (n *NoopStore) GetPreKey(ctx context.Context, id uint32) (*keys.PreKey, error)

func (n *NoopStore) GetPrivacyToken(ctx context.Context, user types.JID) (*PrivacyToken, error)

func (n *NoopStore) GetSenderKey(ctx context.Context, group, user string) ([]byte, error)

func (n *NoopStore) GetSession(ctx context.Context, address string) ([]byte, error)

func (n *NoopStore) HasSession(ctx context.Context, address string) (bool, error)

func (n *NoopStore) IsTrustedIdentity(ctx context.Context, address string, key [32]byte) (bool, error)

func (n *NoopStore) MarkPreKeysAsUploaded(ctx context.Context, upToID uint32) error

func (n *NoopStore) MigratePNToLID(ctx context.Context, pn, lid types.JID) error

func (n *NoopStore) PutAllContactNames(ctx context.Context, contacts []ContactEntry) error

func (n *NoopStore) PutAppStateMutationMACs(ctx context.Context, name string, version uint64, mutations []AppStateMutationMAC) error

func (n *NoopStore) PutAppStateSyncKey(ctx context.Context, id []byte, key AppStateSyncKey) error

func (n *NoopStore) PutAppStateVersion(ctx context.Context, name string, version uint64, hash [128]byte) error

func (n *NoopStore) PutArchived(ctx context.Context, chat types.JID, archived bool) error

func (n *NoopStore) PutBufferedEvent(ctx context.Context, ciphertextHash [32]byte, plaintext []byte, serverTimestamp time.Time) error

func (n *NoopStore) PutBusinessName(ctx context.Context, user types.JID, businessName string) (bool, string, error)

func (n *NoopStore) PutContactName(ctx context.Context, user types.JID, fullName, firstName string) error

func (n *NoopStore) PutDevice(ctx context.Context, store *Device) error

func (n *NoopStore) PutIdentity(ctx context.Context, address string, key [32]byte) error

func (n *NoopStore) PutLIDMapping(ctx context.Context, lid types.JID, jid types.JID) error

func (n *NoopStore) PutManyLIDMappings(ctx context.Context, mappings []LIDMapping) error

func (n *NoopStore) PutManyRedactedPhones(ctx context.Context, entries []RedactedPhoneEntry) error

func (n *NoopStore) PutManySessions(ctx context.Context, sessions map[string][]byte) error

func (n *NoopStore) PutMessageSecret(ctx context.Context, chat, sender types.JID, id types.MessageID, secret []byte) error

func (n *NoopStore) PutMessageSecrets(ctx context.Context, inserts []MessageSecretInsert) error

func (n *NoopStore) PutMutedUntil(ctx context.Context, chat types.JID, mutedUntil time.Time) error

func (n *NoopStore) PutPinned(ctx context.Context, chat types.JID, pinned bool) error

func (n *NoopStore) PutPrivacyTokens(ctx context.Context, tokens ...PrivacyToken) error

func (n *NoopStore) PutPushName(ctx context.Context, user types.JID, pushName string) (bool, string, error)

func (n *NoopStore) PutSenderKey(ctx context.Context, group, user string, session []byte) error

func (n *NoopStore) PutSession(ctx context.Context, address string, session []byte) error

func (n *NoopStore) RemovePreKey(ctx context.Context, id uint32) error

func (n *NoopStore) UploadedPreKeyCount(ctx context.Context) (int, error)

type PreKeyStore interface {
	GetOrGenPreKeys(ctx context.Context, count uint32) ([]*keys.PreKey, error)
	GenOnePreKey(ctx context.Context) (*keys.PreKey, error)
	GetPreKey(ctx context.Context, id uint32) (*keys.PreKey, error)
	RemovePreKey(ctx context.Context, id uint32) error
	MarkPreKeysAsUploaded(ctx context.Context, upToID uint32) error
	UploadedPreKeyCount(ctx context.Context) (int, error)
}

type PrivacyToken struct {
	User      types.JID
	Token     []byte
	Timestamp time.Time
}

type PrivacyTokenStore interface {
	PutPrivacyTokens(ctx context.Context, tokens ...PrivacyToken) error
	GetPrivacyToken(ctx context.Context, user types.JID) (*PrivacyToken, error)
}

type RedactedPhoneEntry struct {
	JID           types.JID
	RedactedPhone string
}

func (rpe RedactedPhoneEntry) GetMassInsertValues() [2]any

type SenderKeyStore interface {
	PutSenderKey(ctx context.Context, group, user string, session []byte) error
	GetSenderKey(ctx context.Context, group, user string) ([]byte, error)
}

type SessionStore interface {
	GetSession(ctx context.Context, address string) ([]byte, error)
	HasSession(ctx context.Context, address string) (bool, error)
	GetManySessions(ctx context.Context, addresses []string) (map[string][]byte, error)
	PutSession(ctx context.Context, address string, session []byte) error
	PutManySessions(ctx context.Context, sessions map[string][]byte) error
	DeleteAllSessions(ctx context.Context, phone string) error
	DeleteSession(ctx context.Context, address string) error
	MigratePNToLID(ctx context.Context, pn, lid types.JID) error
}

type WAVersionContainer [3]uint32
    WAVersionContainer is a container for a WhatsApp web version number.

func GetWAVersion() WAVersionContainer
    GetWAVersion gets the current WhatsApp web client version.

func ParseVersion(version string) (parsed WAVersionContainer, err error)
    ParseVersion parses a version string (three dot-separated numbers) into a
    WAVersionContainer.

func (vc WAVersionContainer) Hash() [16]byte
    Hash returns the md5 hash of the String representation of this version.

func (vc WAVersionContainer) IsZero() bool
    IsZero returns true if the version is zero.

func (vc WAVersionContainer) LessThan(other WAVersionContainer) bool

func (vc WAVersionContainer) ProtoAppVersion() *waWa6.ClientPayload_UserAgent_AppVersion

func (vc WAVersionContainer) String() string
    String returns the version number as a dot-separated string.

