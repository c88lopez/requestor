package main

const jsonFileName = "config.json"

type loginFields struct {
	Username []string
	Password []string
}

type loginParameters struct {
	Url    string
	Fields loginFields
}

type tokens struct {
	DateFrom string
	DateTo   string
}

type configJson struct {
	Login loginParameters

	Report  string
	Results string

	Days    []int
	Timeout int

	SkipNoDays bool

	Domain     string
	WithSafiro bool

	Tokens tokens

	Workers int
	Limit   int
}

func (f loginFields) getUsernameField() string {
	return f.Username[0]
}

func (f loginFields) getUsernameValue() string {
	return f.Username[1]
}

func (f loginFields) getPasswordField() string {
	return f.Password[0]
}

func (f loginFields) getPasswordValue() string {
	return f.Password[1]
}
