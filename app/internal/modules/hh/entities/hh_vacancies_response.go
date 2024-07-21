package entities

// https://api.hh.ru/openapi/redoc#tag/Poisk-vakansij/operation/get-vacancies

type ResponseVacancies struct {
	Arguments []Argument `json:"arguments"`
	Clusters  any        `json:"clusters"`
	Fixes     any        `json:"fixes"`
	Found     int        `json:"found"`
	Items     []Vacancy  `json:"items"`
	Page      int        `json:"page"`
	Pages     int        `json:"pages"`
	PerPage   int        `json:"per_page"`
	Suggests  any        `json:"suggests"`
}

type ClusterGroup struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Argument struct {
	Argument         string       `json:"argument"`
	ClusterGroup     ClusterGroup `json:"cluster_group"`
	DisableURL       string       `json:"disable_url"`
	Value            string       `json:"value"`
	ValueDescription string       `json:"value_description"`
	HexColor         string       `json:"hex_color,omitempty"`
	MetroType        string       `json:"metro_type,omitempty"`
}

type Vacancy struct {
	AcceptIncompleteResumes bool               `json:"accept_incomplete_resumes"`
	Address                 Address            `json:"address"`
	AlternateURL            string             `json:"alternate_url"`
	ApplyAlternateURL       string             `json:"apply_alternate_url"`
	Area                    Area               `json:"area"`
	Contacts                Contacts           `json:"contacts"`
	Counters                Counters           `json:"counters"`
	Department              Department         `json:"department"`
	Employer                Employer           `json:"employer"`
	HasTest                 bool               `json:"has_test"`
	ID                      string             `json:"id"` // vacancy id ??
	Name                    string             `json:"name"`
	PersonalDataResale      bool               `json:"personal_data_resale"`
	ProfessionalRoles       []ProfessionalRole `json:"professional_roles"`
	PublishedAt             string             `json:"published_at"`
	Relations               []any              `json:"relations"`
	ResponseLetterRequired  bool               `json:"response_letter_required"`
	ResponseURL             any                `json:"response_url"`
	Salary                  Salary             `json:"salary"`
	Schedule                Schedule           `json:"schedule"`
	ShowLogoInSearch        bool               `json:"show_logo_in_search"`
	Snippet                 Snippet            `json:"snippet"`
	SortPointDistance       float64            `json:"sort_point_distance"`
	Type                    Type               `json:"type"`
	URL                     string             `json:"url"`
}

type MetroStation struct {
	Lat         float64 `json:"lat"`
	LineID      string  `json:"line_id"`
	LineName    string  `json:"line_name"`
	Lng         float64 `json:"lng"`
	StationID   string  `json:"station_id"`
	StationName string  `json:"station_name"`
}

type Address struct {
	Building      string         `json:"building"`
	City          string         `json:"city"`
	Description   string         `json:"description"`
	Lat           float64        `json:"lat"`
	Lng           float64        `json:"lng"`
	MetroStations []MetroStation `json:"metro_stations"`
	Street        string         `json:"street"`
}

type Area struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Phone struct {
	City    string `json:"city"`
	Comment any    `json:"comment"`
	Country string `json:"country"`
	Number  string `json:"number"`
}

type Contacts struct {
	Email  string  `json:"email"`
	Name   string  `json:"name"`
	Phones []Phone `json:"phones"`
}

type Counters struct {
	Responses int `json:"responses"`
}

type Department struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Employer struct {
	AccreditedItEmployer bool   `json:"accredited_it_employer"`
	AlternateURL         string `json:"alternate_url"`
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Trusted              bool   `json:"trusted"`
	URL                  string `json:"url"`
}

type ProfessionalRole struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Salary struct {
	Currency string `json:"currency"`
	From     int    `json:"from"`
	Gross    bool   `json:"gross"`
	To       any    `json:"to"`
}

type Schedule struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Snippet struct {
	Requirement    string `json:"requirement"`
	Responsibility string `json:"responsibility"`
}

type Type struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
