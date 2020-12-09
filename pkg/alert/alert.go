package alert

var (
	DefaultAlert = NopAlert{}
)

type Alert interface {
	Send(msg string) error
	AsyncSend(msg string)
	Close() error
}

var _ Alert = (*NopAlert)(nil)

type NopAlert struct {
}

func (n NopAlert) Send(_ string) error {
	return nil
}

func (n NopAlert) AsyncSend(_ string) {
	return
}

func (n NopAlert) Close() error {
	return nil
}

func Send(msg string) error {
	return DefaultAlert.Send(msg)
}

func AsyncSend(msg string) {
	DefaultAlert.AsyncSend(msg)
}

func Close() error {
	return DefaultAlert.Close()
}
