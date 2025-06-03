package analyzer

import (
	commands "backend/commands"
	"errors"
	"fmt"
	"strings"
)


func Analyzer(input string) (interface{}, error){
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "#"){
		return nil, nil

	}

	tokens := strings.Fields(input)

	if len(tokens) == 0{
		return nil, errors.New("no se proporcionó ningún comando")

	}

	switch tokens[0]{
	case "mkdisk":
		return commands.ParseMkdisk(tokens[1:])

	case "rmdisk":
		return commands.ParseRmdisk(tokens[1:])

	case "fdisk":
		return commands.ParseFdisk(tokens[1:])

	case "mount":
		return commands.ParseMount(tokens[1:])

	case "mounted":
		return nil, commands.ParseMounted()

	case "mkfs":
		return commands.ParseMkfs(tokens[1:])

	case "cat":
		return commands.ParseCat(tokens[1:])

	case "login":
		return commands.ParseLogin(tokens[1:])

	case "logout":
		return commands.ParseLogout(tokens[1:])

	case "mkgrp":
		return commands.ParseMkgrp(tokens[1:])

	case "rmgrp":
		return commands.ParseRmgrp(tokens[1:])

	case "mkusr":
		return commands.ParseMkusr(tokens[1:])

	case "rmusr":
		return commands.ParseRmusr(tokens[1:])

	case "chgrp":
		return commands.ParseChgrp(tokens[1:])

	case "mkdir":
		return commands.ParseMkdir(tokens[1:])

	case "mkfile":
		return commands.ParseMkfile(tokens[1:])

	case "rep":
		return commands.ParseRep(tokens[1:])

	default:
		return nil, fmt.Errorf("comando desconocido: %s", tokens[0])

	}

}
