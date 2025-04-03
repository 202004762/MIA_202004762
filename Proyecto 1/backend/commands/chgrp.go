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


type CHGRP struct{
	user string
	grp  string

}

func ParseChgrp(tokens []string) (string, error){
	if !stores.Auth.IsAuthenticated(){
		return "", errors.New("no hay ninguna sesion iniciada")

	}

	if strings.ToLower(stores.Auth.Username) != "root"{
		return "", errors.New("solo el usuario root puede cambiar de grupo a un usuario")

	}

	cmd := &CHGRP{}
	args := strings.Join(tokens, " ")
	reUser := regexp.MustCompile(`-user="?([^\s"]+)"?`)
	reGrp := regexp.MustCompile(`-grp="?([^\s"]+)"?`)
	userMatch := reUser.FindStringSubmatch(args)
	grpMatch := reGrp.FindStringSubmatch(args)
	if len(userMatch) < 2 || len(grpMatch) < 2{
		return "", errors.New("parametros invalidos. Se requieren -user y -grp")

	}

	cmd.user = userMatch[1]
	cmd.grp = grpMatch[1]
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

	if !groupExistsInFile(inode, sb, path, cmd.grp){
		return "", fmt.Errorf("el grupo '%s' no existe o esta eliminado", cmd.grp)

	}

	userFound := false
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

		lines := strings.Split(content, "\n")
		updated := false
		for i, line := range lines{
			fields := strings.Split(line, ",")
			if len(fields) >= 5 && strings.TrimSpace(fields[1]) == "U"{
				if strings.TrimSpace(fields[3]) == cmd.user{
					fields[2] = cmd.grp
					lines[i] = strings.Join(fields, ",")
					userFound = true
					updated = true

				}

			}

		}

		if updated{
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

			//fmt.Printf("DEBUG - Grupo cambiado para usuario '%s' en bloque %d\n", cmd.user, blk)

		}

	}

	if userFound{
		inode.I_mtime = float32(time.Now().Unix())
		_ = inode.Serialize(path, int64(sb.S_inode_start+1*sb.S_inode_size))
		return fmt.Sprintf("------------------------"+
			"CHGRP: grupo cambiado exitosamente\n "+
			"Usuario '%s' actualizado al grupo '%s'", cmd.user, cmd.grp), nil

	}

	return "", fmt.Errorf("el usuario '%s' no existe", cmd.user)

}

func groupExistsInFile(inode *structures.Inode, sb *structures.SuperBlock, path, groupName string) bool{
	for _, blk := range inode.I_block{
		if blk == -1{
			continue

		}

		offset := int64(sb.S_block_start + blk*sb.S_block_size)
		block := &structures.FileBlock{}
		_ = block.Deserialize(path, offset)
		content := strings.Trim(string(block.B_content[:]), "\x00")
		lines := strings.Split(content, "\n")
		for _, line := range lines{
			fields := strings.Split(line, ",")
			if len(fields) >= 3 && strings.TrimSpace(fields[1]) == "G"{
				if strings.TrimSpace(fields[2]) == groupName && strings.TrimSpace(fields[0]) != "0"{
					return true

				}

			}

		}

	}

	return false

}
