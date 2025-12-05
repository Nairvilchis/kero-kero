package armadillo // import "go.mau.fi/whatsmeow/proto"


TYPES

type MessageApplicationSub interface {
	IsMessageApplicationSub()
}

type RealMessageApplicationSub interface {
	MessageApplicationSub
	proto.Message
}

type Unsupported_BusinessApplication waCommon.SubProtocol

func (*Unsupported_BusinessApplication) IsMessageApplicationSub()

type Unsupported_PaymentApplication waCommon.SubProtocol

func (*Unsupported_PaymentApplication) IsMessageApplicationSub()

type Unsupported_Voip waCommon.SubProtocol

func (*Unsupported_Voip) IsMessageApplicationSub()

