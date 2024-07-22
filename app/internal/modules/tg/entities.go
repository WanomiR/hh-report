package tg

type UpdatesResponse struct {
	Ok          bool     `json:"ok"`
	Description string   `json:"description,omitempty"`
	Result      []Update `json:"result"`
}

type Update struct {
	ID      int      `json:"update_id"`
	Message *Message `json:"message"` // Optional. New incoming message of any kind - text, photo, sticker, etc.
	// at most one of the optional parameters can be present in any given update
}

type Message struct {
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

type Query struct {
	Area       string
	Role       string
	Text       string
	Experience string
}
