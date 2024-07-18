package entities

type Type int

const (
	Unknown Type = iota
	Message
)

type Meta struct {
	ChatID   int
	Username string
}

type Event struct {
	Type Type
	Text string
	Meta interface{}
}

type UpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	ID      int              `json:"update_id"`
	Message *IncomingMessage `json:"message"` // Optional. New incoming message of any kind - text, photo, sticker, etc.
	// at most one of the optional parameters can be present in any given update
}

type IncomingMessage struct {
	Text string `json:"text"`
	From User   `json:"from"`
	Chat Chat   `json:"chat"`
}

type User struct {
	Id       int    `json:"id"`
	IsBot    bool   `json:"is_bot"`
	Username string `json:"username"`
}

type Chat struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}
