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


type RMGRP struct{
	name string

}

func ParseRmgrp(tokens []string) (string, error){
	if !stores.Auth.IsAuthenticated(){
		return "", errors.New("no hay ninguna sesion iniciada")

	}

	if strings.ToLower(stores.Auth.Username) != "root"{
		return "", errors.New("solo el usuario root puede eliminar grupos")
		
	}

	cmd := &RMGRP{}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-name="?([^\s"]+)"?`)
	match := re.FindStringSubmatch(args)
	if len(match) < 2{
		return "", errors.New("falta el parametro obligatorio -name")

	}

	cmd.name = match[1]
	if cmd.name == ""{
		return "", errors.New("el nombre del grupo no puede estar vacio")

	}

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
		updated := false
		for i, line := range lines{
			fields := strings.Split(line, ",")
			if len(fields) >= 3 && strings.TrimSpace(fields[1]) == "G" && strings.TrimSpace(fields[2]) == cmd.name{
				fields[0] = "0" // marcar como eliminado
				lines[i] = strings.Join(fields, ",")
				updated = true
				found = true

			}

		}

		if updated{
			newContent := strings.Join(lines, "\n")
			if !strings.HasSuffix(newContent, "\n"){
				newContent += "\n"

			}

			copy(block.B_content[:], strings.Repeat("\x00", 64))
			copy(block.B_content[:], newContent)
			err = block.Serialize(path, offset)
			if err != nil{
				return "", err

			}

			//fmt.Printf("DEBUG - Grupo '%s' eliminado en bloque %d\n", cmd.name, blk)

		}

	}

	if found{
		inode.I_mtime = float32(time.Now().Unix())
		_ = inode.Serialize(path, int64(sb.S_inode_start+1*sb.S_inode_size))
		return fmt.Sprintf("------------------------"+
			"RMGRP: grupo eliminado exitosamente\n "+
			"Grupo '%s' eliminado", cmd.name), nil

	}

	return "", fmt.Errorf("el grupo '%s' no existe", cmd.name)

}
