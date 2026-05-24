package inbox

type Kind int

const (
	KindUndefined Kind = iota
	KindNotification
	KindTelegram
)

func (o Kind) String() string {
	switch o {
	case KindUndefined:
		return "undefined"
	case KindNotification:
		return "notification"
	case KindTelegram:
		return "Telegram"
	default:
		return ""
	}
}

type Data struct {
	IdempotencyKey string
	Kind           Kind
	Data           []byte
}
