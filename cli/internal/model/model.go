package model

type Requirement struct {
	Feature      string `json:"feature"`
	Slug         string `json:"slug"`
	Confidence   string `json:"confidence"`
	EARS         string `json:"ears"`
	Fit          string `json:"fit"`
	SupersededBy string `json:"supersededBy,omitempty"`
	Line         int    `json:"-"`
}

func (r Requirement) Qualified() string { return r.Feature + "/" + r.Slug }

type Spec struct {
	Path         string
	Feature      string `json:"slug"`
	Status       string `json:"status"`
	Depth        string `json:"depth"`
	Sections     map[string][]string
	Requirements []Requirement `json:"requirements"`
	Line         int
}

type Open struct {
	Slug     string `json:"slug"`
	Status   string `json:"status"`
	Pass     string `json:"pass"`
	Asked    string `json:"asked"`
	Owner    string `json:"owner"`
	Question string `json:"question"`
	Cost     string `json:"cost"`
	Blocks   string `json:"blocks"`
	Line     int    `json:"-"`
}

type ParseError struct {
	Path string
	Line int
	Msg  string
}

func (e ParseError) Error() string {
	return e.Path + ":" + itoa(e.Line) + ": " + e.Msg
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
