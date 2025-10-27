package namedargs

import (
	"fmt"
	"os"
	"strings"
)

func Parse(args []string) (map[string]string, error) {
	namedArgs := make(map[string]string)
	for _, arg := range args {
		index := strings.Index(arg, "=")
		if index == -1 {
			return nil, fmt.Errorf("invalid arg: %s", arg)
		}
		namedArgs[arg[:index]] = arg[index+1:]
	}
	return namedArgs, nil
}

func ParseArgs() (map[string]string, error) {
	return Parse(os.Args[1:])
}
