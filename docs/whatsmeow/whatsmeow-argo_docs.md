package argo // import "go.mau.fi/whatsmeow/argo"


VARIABLES

var (
	Store                map[string]wire.Type
	QueryIDToMessageName map[string]string
)

FUNCTIONS

func GetQueryIDToMessageName() (map[string]string, error)
func GetStore() (map[string]wire.Type, error)
func Init() error
