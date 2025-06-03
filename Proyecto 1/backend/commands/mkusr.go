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


type MKUSR struct{
	user string
	pass string
	grp  string

}

func ParseMkusr(tokens []string) (string, error){
	if !stores.Auth.IsAuthenticated(){
		return "", errors.New("no hay ninguna sesion iniciada")

	}

	if strings.ToLower(stores.Auth.Username) != "root"{
		return "", errors.New("solo el usuario root puede crear usuarios")

	}

	cmd := &MKUSR{}
	args := strings.Join(tokens, " ")
	reUser := regexp.MustCompile(`-user="?([^\s"]+)"?`)
	rePass := regexp.MustCompile(`-pass="?([^\s"]+)"?`)
	reGrp := regexp.MustCompile(`-grp="?([^\s"]+)"?`)
	userMatch := reUser.FindStringSubmatch(args)
	passMatch := rePass.FindStringSubmatch(args)
	grpMatch := reGrp.FindStringSubmatch(args)
	if len(userMatch) < 2 || len(passMatch) < 2 || len(grpMatch) < 2{
		return "", errors.New("parametros invalidos. Se requieren -user, -pass y -grp")

	}

	cmd.user = userMatch[1]
	cmd.pass = passMatch[1]
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

	existingUserIDs := make(map[int]bool)
	userMaxID := 0
	groupExists := false
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

		blockContent := strings.Trim(string(block.B_content[:]), "\x00")
		if blockContent == ""{
			continue

		}

		//fmt.Printf("DEBUG - Bloque existente %d contiene:\n%s\n", blk, blockContent)

		lines := strings.Split(blockContent, "\n")
		for _, line := range lines{
			if line == ""{
				continue

			}

			fields := strings.Split(line, ",")
			if len(fields) >= 5 && strings.TrimSpace(fields[1]) == "U"{
				if strings.TrimSpace(fields[3]) == cmd.user{
					return "", fmt.Errorf("el usuario '%s' ya existe", cmd.user)

				}

				id := parseID2(fields[0])
				if id > 0{
					existingUserIDs[id] = true
					if id > userMaxID{
						userMaxID = id

					}

				}

			}else if len(fields) >= 3 && strings.TrimSpace(fields[1]) == "G"{
				if strings.TrimSpace(fields[2]) == cmd.grp && fields[0] != "0"{
					groupExists = true

				}

			}

		}

	}

	if !groupExists{
		return "", fmt.Errorf("el grupo '%s' no existe", cmd.grp)

	}

	userMaxID++

	//fmt.Printf("DEBUG - Nuevo ID de usuario asignado: %d\n", userMaxID)

	newLine := fmt.Sprintf("%d,U,%s,%s,%s\n", userMaxID, cmd.grp, cmd.user, cmd.pass)
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

		existing := strings.Trim(string(block.B_content[:]), "\x00")
		totalContent := existing + newLine
		if len(totalContent) <= 64{
			copy(block.B_content[:], totalContent)
			err = block.Serialize(path, offset)
			if err != nil{
				return "", err

			}

			inode.I_mtime = float32(time.Now().Unix())
			_ = inode.Serialize(path, int64(sb.S_inode_start+1*sb.S_inode_size))

			//fmt.Printf("DEBUG - Usuario escrito en bloque existente: %d\n", blk)
			//fmt.Printf("DEBUG - Inodo I_block: %v\n", inode.I_block)

			return fmt.Sprintf("------------------------"+
				"MKUSR: usuario creado exitosamente\n "+
				"Usuario '%s' creado", cmd.user), nil

		}

	}

	for i := range inode.I_block{
		if inode.I_block[i] == -1{
			nuevoBloque := sb.S_blocks_count
			for _, blk := range inode.I_block{
				if int32(nuevoBloque) == blk{
					nuevoBloque++

				}

			}

			newBlock := &structures.FileBlock{}
			copy(newBlock.B_content[:], newLine)
			inode.I_block[i] = int32(nuevoBloque)
			offset := int64(sb.S_block_start + nuevoBloque*sb.S_block_size)
			err := newBlock.Serialize(path, offset)
			if err != nil{
				return "", err

			}

			sb.S_blocks_count = nuevoBloque + 1
			sb.S_free_blocks_count--
			sb.S_first_blo += sb.S_block_size
			err = sb.UpdateBitmapBlockAt(path, int32(nuevoBloque))
			if err != nil{
				return "", err

			}

			inode.I_mtime = float32(time.Now().Unix())
			_ = inode.Serialize(path, int64(sb.S_inode_start+1*sb.S_inode_size))

			//fmt.Printf("DEBUG - Se creó nuevo bloque para usuario en I_block[%d] = %d\n", i, nuevoBloque)
			//fmt.Printf("DEBUG - Inodo I_block actualizado: %v\n", inode.I_block)

			return fmt.Sprintf("------------------------"+
				"MKUSR: usuario creado exitosamente\n "+
				"Usuario '%s' creado", cmd.user), nil

		}

	}

	return "", errors.New("no hay espacio suficiente para agregar el usuario")

}

func parseID2(id string) int{
	var i int
	fmt.Sscanf(id, "%d", &i)
	return i

}
