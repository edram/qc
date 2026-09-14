package cli

type CLI struct {
	Search SearchCmd `kong:"cmd,help='Search enterprise and person records.'"`
}

func New() *CLI {
	return &CLI{}
}
