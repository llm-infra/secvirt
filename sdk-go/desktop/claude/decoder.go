package claude

type Decoder struct {
	parser *Parser
}

func NewDecoder() *Decoder {
	return &Decoder{
		parser: NewParser(),
	}
}

func (d *Decoder) Decode(data []byte) ([]Message, error) {
	return d.parser.ProcessLine(string(data))
}
