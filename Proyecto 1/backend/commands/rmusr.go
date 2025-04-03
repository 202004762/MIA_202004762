package commands

import (
	"backend/stores"
	"backend/structures"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)


type RMUSR struct{
	user string

}

func ParseRmusr(tokens []string) (string, error){
	if !stores.Auth.IsAuthenticated(){
		return "", errors.New("no hay ninguna sesion iniciada")

	}

	if strings.ToLower(stores.Auth.Username) != "root"{
		return "", errors.New("solo el usuario root puede eliminar usuarios")

	}

	cmd := &RMUSR{}
	args := strings.Join(tokens, " ")
	reUser := regexp.MustCompile(`-user="?([^\s"]+)"?`)
	userMatch := reUser.FindStringSubmatch(args)
	if len(userMatch) < 2{
		return "", errors.New("falta el parametro obligatorio -user")

	}

	cmd.user = userMatch[1]
	sb, _, path, err := stores.GetMountedPartitionSuperblock(stores.Auth.PartitionID)
	if err != nil{
		return "", err

	}

	inode := &structures.Inode{}
	offset := int64(sb.S_inode_start + 1*sb.S_inode_size)
	err = inode.Deserialize(path, offset)
	if err != nil{
		return "", err

	}

	found := false
	for _, blk := range inode.I_block{
		if blk == -1{
			continue

		}

		offset := int64(sb.S_block_start + blk*sb.S_block_size)
		block := &structures.FileBlock{}
		err := block.Deserialize(path, offset)
		if err != nil{
			return "", err

		}

		content := strings.Trim(string(block.B_content[:]), "\x00")
		if content == ""{
			continue

		}

		//fmt.Printf("DEBUG - Bloque %d antes de eliminar:\n%s\n", blk, content)

		lines := strings.Split(content, "\n")
		modified := false
		for i, line := range lines{
			if line == ""{
				continue

			}

			fields := strings.Split(line, ",")
			if len(fields) >= 5 && strings.TrimSpace(fields[1]) == "U" && strings.TrimSpace(fields[3]) == cmd.user{
				fields[0] = "0"
				lines[i] = strings.Join(fields, ",")
				found = true
				modified = true

			}

		}

		if modified{
			newContent := strings.Join(lines, "\n")
			if !strings.HasSuffix(newContent, "\n"){
				newContent += "\n"

			}

			copy(block.B_content[:], strings.Repeat("\x00", 64)) // limpiar
			copy(block.B_content[:], newContent)
			err := block.Serialize(path, offset)
			if err != nil{
				return "", err

			}

			//fmt.Printf("DEBUG - Usuario '%s' eliminado en bloque %d\n", cmd.user, blk)

		}

	}

	if found{
		inode.I_mtime = float32(time.Now().Unix())
		_ = inode.Serialize(path, int64(sb.S_inode_start+1*sb.S_inode_size))
		return fmt.Sprintf("------------------------"+
			"RMUSR: usuario eliminado exitosamente\n "+
			"Usuario '%s' eliminado", cmd.user), nil

	}

	return "", fmt.Errorf("el usuario '%s' no existe", cmd.user)

}
