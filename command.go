package state

type (
	ExitStatus uint8

	Command interface {
		GetName() string
		Run() ExitStatus
	}

	dumpCommand State
	listCommand State
)

func (c *dumpCommand) GetName() string { return "dump" }

func (c *dumpCommand) Run() ExitStatus { return 0 }

func (c *listCommand) GetName() string { return "list" }

func (c *listCommand) Run() ExitStatus { return 0 }
