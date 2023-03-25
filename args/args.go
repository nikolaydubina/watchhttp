package args

const terminator = "--"

// GetCommandFromArgs returns command from command args.
// command is determined as a whole string after terminator (`--`) if terminator present.
// if no terminator, considering whole string as command.
// this goes well with flag.Parse which will stop parsing before terminator.
func GetCommandFromArgs(args []string) (command []string, hasFlags bool) {
	idx := -1
	for i, q := range args {
		if q == terminator {
			idx = i
			break
		}
	}

	hasFlags = idx > 0

	// command
	idx += 1
	if idx >= len(args) {
		return nil, hasFlags
	}

	// check if starts with `-` which means it is not a command but argument
	if args[idx][0] == '-' {
		return nil, true
	}

	// it is a command
	command = make([]string, len(args[idx:]))
	copy(command, args[idx:])

	return command, hasFlags
}
