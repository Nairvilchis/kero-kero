package appstate // import "go.mau.fi/whatsmeow/appstate"

Package appstate implements encoding and decoding WhatsApp's app state patches.

CONSTANTS

const (
	IndexMute                    = "mute"
	IndexPin                     = "pin_v1"
	IndexArchive                 = "archive"
	IndexContact                 = "contact"
	IndexClearChat               = "clearChat"
	IndexDeleteChat              = "deleteChat"
	IndexStar                    = "star"
	IndexDeleteMessageForMe      = "deleteMessageForMe"
	IndexMarkChatAsRead          = "markChatAsRead"
	IndexSettingPushName         = "setting_pushName"
	IndexSettingUnarchiveChats   = "setting_unarchiveChats"
	IndexUserStatusMute          = "userStatusMute"
	IndexLabelEdit               = "label_edit"
	IndexLabelAssociationChat    = "label_jid"
	IndexLabelAssociationMessage = "label_message"
)
    Constants for the first part of app state indexes.


VARIABLES

var (
	ErrMissingPreviousSetValueOperation = errors.New("missing value MAC of previous SET operation")
	ErrMismatchingLTHash                = errors.New("mismatching LTHash")
	ErrMismatchingPatchMAC              = errors.New("mismatching patch MAC")
	ErrMismatchingContentMAC            = errors.New("mismatching content MAC")
	ErrMismatchingIndexMAC              = errors.New("mismatching index MAC")
	ErrKeyNotFound                      = errors.New("didn't find app state key")
)
    Errors that this package can return.

var AllPatchNames = [...]WAPatchName{WAPatchCriticalBlock, WAPatchCriticalUnblockLow, WAPatchRegularHigh, WAPatchRegular, WAPatchRegularLow}
    AllPatchNames contains all currently known patch state names.


TYPES

type DownloadExternalFunc func(context.Context, *waServerSync.ExternalBlobReference) ([]byte, error)
    DownloadExternalFunc is a function that can download a blob of external app
    state patches.

type ExpandedAppStateKeys struct {
	Index           []byte
	ValueEncryption []byte
	ValueMAC        []byte
	SnapshotMAC     []byte
	PatchMAC        []byte
}

type HashState struct {
	Version uint64
	Hash    [128]byte
}

type Mutation struct {
	Operation waServerSync.SyncdMutation_SyncdOperation
	Action    *waSyncAction.SyncActionValue
	Version   int32
	Index     []string
	IndexMAC  []byte
	ValueMAC  []byte
}

type MutationInfo struct {
	// Index contains the thing being mutated (like `mute` or `pin_v1`), followed by parameters like the target JID.
	Index []string
	// Version is a static number that depends on the thing being mutated.
	Version int32
	// Value contains the data for the mutation.
	Value *waSyncAction.SyncActionValue
}
    MutationInfo contains information about a single mutation to the app state.

type PatchInfo struct {
	// Timestamp is the time when the patch was created. This will be filled automatically in EncodePatch if it's zero.
	Timestamp time.Time
	// Type is the app state type being mutated.
	Type WAPatchName
	// Mutations contains the individual mutations to apply to the app state in this patch.
	Mutations []MutationInfo
}
    PatchInfo contains information about a patch to the app state. A patch can
    contain multiple mutations, as long as all mutations are in the same app
    state type.

func BuildArchive(target types.JID, archive bool, lastMessageTimestamp time.Time, lastMessageKey *waCommon.MessageKey) PatchInfo
    BuildArchive builds an app state patch for archiving or unarchiving a chat.

    The last message timestamp and last message key are optional and can be set
    to zero values (`time.Time{}` and `nil`).

    Archiving a chat will also unpin it automatically.

func BuildDeleteChat(target types.JID, lastMessageTimestamp time.Time, lastMessageKey *waCommon.MessageKey) PatchInfo
    BuildDeleteChat builds an app state patch for deleting a chat.

func BuildLabelChat(target types.JID, labelID string, labeled bool) PatchInfo
    BuildLabelChat builds an app state patch for labeling or un(labeling) a
    chat.

func BuildLabelEdit(labelID string, labelName string, labelColor int32, deleted bool) PatchInfo
    BuildLabelEdit builds an app state patch for editing a label.

func BuildLabelMessage(target types.JID, labelID, messageID string, labeled bool) PatchInfo
    BuildLabelMessage builds an app state patch for labeling or un(labeling) a
    message.

func BuildMarkChatAsRead(target types.JID, read bool, lastMessageTimestamp time.Time, lastMessageKey *waCommon.MessageKey) PatchInfo
    BuildMarkChatAsRead builds an app state patch for marking a chat as read or
    unread.

func BuildMute(target types.JID, mute bool, muteDuration time.Duration) PatchInfo
    BuildMute builds an app state patch for muting or unmuting a chat.

    If mute is true and the mute duration is zero, the chat is muted forever.

func BuildMuteAbs(target types.JID, mute bool, muteEndTimestamp *int64) PatchInfo
    BuildMuteAbs builds an app state patch for muting or unmuting a chat with an
    absolute timestamp.

func BuildPin(target types.JID, pin bool) PatchInfo
    BuildPin builds an app state patch for pinning or unpinning a chat.

func BuildSettingPushName(pushName string) PatchInfo
    BuildSettingPushName builds an app state patch for setting the push name.

func BuildStar(target, sender types.JID, messageID types.MessageID, fromMe, starred bool) PatchInfo
    BuildStar builds an app state patch for starring or unstarring a message.

type PatchList struct {
	Name           WAPatchName
	HasMorePatches bool
	Patches        []*waServerSync.SyncdPatch
	Snapshot       *waServerSync.SyncdSnapshot
}
    PatchList represents a decoded response to getting app state patches from
    the WhatsApp servers.

func ParsePatchList(ctx context.Context, node *waBinary.Node, downloadExternal DownloadExternalFunc) (*PatchList, error)
    ParsePatchList will decode an XML node containing app state patches,
    including downloading any external blobs.

type Processor struct {
	Store *store.Device
	Log   waLog.Logger
	// Has unexported fields.
}

func NewProcessor(store *store.Device, log waLog.Logger) *Processor

func (proc *Processor) DecodePatches(ctx context.Context, list *PatchList, initialState HashState, validateMACs bool) (newMutations []Mutation, currentState HashState, err error)
    DecodePatches will decode all the patches in a PatchList into a list of app
    state mutations.

func (proc *Processor) EncodePatch(ctx context.Context, keyID []byte, state HashState, patchInfo PatchInfo) ([]byte, error)

func (proc *Processor) GetMissingKeyIDs(ctx context.Context, pl *PatchList) [][]byte

type WAPatchName string
    WAPatchName represents a type of app state patch.

const (
	// WAPatchCriticalBlock contains the user's settings like push name and locale.
	WAPatchCriticalBlock WAPatchName = "critical_block"
	// WAPatchCriticalUnblockLow contains the user's contact list.
	WAPatchCriticalUnblockLow WAPatchName = "critical_unblock_low"
	// WAPatchRegularLow contains some local chat settings like pin, archive status, and the setting of whether to unarchive chats when messages come in.
	WAPatchRegularLow WAPatchName = "regular_low"
	// WAPatchRegularHigh contains more local chat settings like mute status and starred messages.
	WAPatchRegularHigh WAPatchName = "regular_high"
	// WAPatchRegular contains protocol info about app state patches like key expiration.
	WAPatchRegular WAPatchName = "regular"
)
