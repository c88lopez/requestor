package main

const jsonFileName = "config.json"

type configJson struct {
	Url   urlParameters
	Login loginParameters

	Report  string
	Results string

	Days    []int
	Timeout int

	Skip skipParameters

	Tokens tokens

	Workers int
	Limit   int
}

type urlParameters struct {
	Schema string
	Domain string
}

type loginParameters struct {
	Enabled bool

	Path   string
	Fields loginFieldsParameters
}

type loginFieldsParameters struct {
	Username []string
	Password []string
}

type tokens struct {
	DateFrom string
	DateTo   string
}

type skipParameters struct {
	NoDays bool
	Hints  []string
}

func (c configJson) getLoginPath() string {
	return c.Login.Path
}

func (c configJson) getUsernameField() string {
	return c.Login.Fields.Username[0]
}

func (c configJson) getUsernameValue() string {
	return c.Login.Fields.Username[1]
}

func (c configJson) getPasswordField() string {
	return c.Login.Fields.Password[0]
}

func (c configJson) getPasswordValue() string {
	return c.Login.Fields.Password[1]
}

func (c configJson) skipNoDays() bool {
	return c.Skip.NoDays
}
