package log

type Message struct {
	Level Level
	Tags  []string
	Text  string
}
