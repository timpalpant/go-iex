package deep

const (
	ChannelID         uint32 = 1
	MessageProtocolID uint16 = 0x8004
)

type Protocol struct{}

func (p Protocol) ID() uint16 {
	return MessageProtocolID
}
