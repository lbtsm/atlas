package new_marker

type Writer interface {
	ResolveMessage(message Message) bool
}
