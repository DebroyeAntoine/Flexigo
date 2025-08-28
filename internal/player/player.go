package player

type AudioPlayer interface {
	PlayMP3(data []byte) error
	Stop() error
}
