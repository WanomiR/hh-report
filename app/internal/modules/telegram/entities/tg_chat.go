package entities

type SearchParams struct {
	Area             int
	ProfessionalRole int
	Text             string
	Experience       string
}

type TgChat struct {
	IsTicking   bool
	StopTicking chan bool
	Id          int
	ParamsList  []SearchParams
}
