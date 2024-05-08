package art

import (
	"strings"

	"git.tcp.direct/kayos/common/squish"
)

const banner = "H4sIAAAAAAACA7VWz2qDMBy++wq95NBrtVuLZWzsRSpI6aSUrmuxMtjwEGRHD85m4mHvsfsexSdZqo35XzvFECHB7/t+f6MBbqdhWA7oMhzDcUMABvO72f10vC1QXGQfeGdXO4iX02qZDOaTEnMBXaLG29MjgCopioNEjbMByexfhtA1YqJeeyXRrVIJgNDAD5vKqAZmR0I3KTlSpV/F+f3hSJLXkQKYHU/REM/FLOZ9qlBuGcKZFUvFpPlNNKajZgOsf0x9rhFmK2ZfyEFaICT2AvEm52CxkBNbcjmt5PjuonLMiQQ9C6nj1OlBIdQspZWN/m9gxjZCLkYlVYcYKrJPPJmeks50hRB8Or//Ji9jxrIyS20JkUSAisPDtshE7mpWIeG3uS4xeDkul18JnrpPzBVoOQc6MOLjY0qa9CdzTaDlaMdl3OKOUs0T/U40BUGKhqGKSP0voU2SK9xHQkvpPnmwX3renOiG4BU/UL6vR4+rdWAGy735tPa9ZWBtFm+7g7XxvL0PRuB1eGMOb/HisNjunz2w81eLl/W75z+MagnQ+vYVusbpDuVaHW5wlvEHTKOp9AMKAAA="
const version = "0.2"

func String() string {
	c, err := squish.UnpackStr(banner)
	if err != nil {
		panic(err)
	}
	c = strings.ReplaceAll(c, "$1", strings.Split(version, ".")[0]) + "."
	return strings.ReplaceAll(c, "$2", strings.Split(version, ".")[1])
}
