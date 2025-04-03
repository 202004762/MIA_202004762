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


type MKGRP struct{
	name string

}

func ParseMkgrp(tokens []string) (string, error){
	if !stores.Auth.IsAuthenticated(){
		return "", errors.New("no hay ninguna sesion iniciada")

	}

	if strings.ToLower(stores.Auth.Username) != "root"{
		return "", errors.New("solo el usuario root puede crear grupos")

	}

	cmd := &MKGRP{}
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

	content, err := sb.GetUsersBlock(path)
	if err != nil{
		return "", err

	}

	lines := strings.Split(strings.Trim(content, "\x00"), "\n")
	existingIDs := make(map[int]bool)
	maxID := 0
	for _, line := range lines{
		fields := strings.Split(line, ",")
		if len(fields) >= 3 && strings.TrimSpace(fields[1]) == "G"{
			id := parseID(fields[0])
			if id > 0{
				existingIDs[id] = true
				if id > maxID{
					maxID = id

				}

				if strings.TrimSpace(fields[2]) == cmd.name && fields[0] != "0"{
					return "", fmt.Errorf("el grupo '%s' ya existe", cmd.name)

				}

			}

		}

	}

	maxID++

	//fmt.Printf("DEBUG - Nuevo ID asignado: %d\n", maxID)

	newLine := fmt.Sprintf("%d,G,%s\n", maxID, cmd.name)
	inode := &structures.Inode{}
	offset := int64(sb.S_inode_start + 1*sb.S_inode_size)
	err = inode.Deserialize(path, offset)
	if err != nil{
		return "", err

	}

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

		// DEBUG: contenido actual del bloque
		//fmt.Printf("DEBUG - Bloque existente %d contiene:\n%s\n", blk, existing)

		totalContent := existing + newLine
		if len(totalContent) <= 64{
			copy(block.B_content[:], totalContent)
			err = block.Serialize(path, offset)
			if err != nil{
				return "", err

			}

			inode.I_mtime = float32(time.Now().Unix())
			_ = inode.Serialize(path, int64(sb.S_inode_start+1*sb.S_inode_size))

			// DEBUG: bloques utilizados por el inodo
			//fmt.Printf("DEBUG - Inodo I_block actualizado (escritura en bloque existente): %v\n", inode.I_block)

			return fmt.Sprintf("------------------------"+
				"MKGRP: grupo creado exitosamente\n "+
				"Grupo '%s' creado", cmd.name), nil

		}

	}

	for i := range inode.I_block{
		if inode.I_block[i] == -1{
			newBlock := &structures.FileBlock{}
			copy(newBlock.B_content[:], newLine)
			inode.I_block[i] = int32(sb.S_blocks_count)

			// DEBUG: creación de nuevo bloque
			//fmt.Printf("DEBUG - Creando nuevo bloque en posicion %d (I_block[%d])\n", sb.S_blocks_count, i)

			offset := int64(sb.S_block_start + sb.S_blocks_count*sb.S_block_size)
			err := newBlock.Serialize(path, offset)
			if err != nil{
				return "", err

			}

			sb.S_blocks_count++
			sb.S_free_blocks_count--
			sb.S_first_blo += sb.S_block_size
			err = sb.UpdateBitmapBlockAt(path, inode.I_block[i])
			if err != nil{
				return "", err

			}

			inode.I_mtime = float32(time.Now().Unix())
			_ = inode.Serialize(path, int64(sb.S_inode_start+1*sb.S_inode_size))

			//DEBUG: bloques utilizados por el inodo
			//fmt.Printf("DEBUG - Inodo I_block actualizado (nuevo bloque): %v\n", inode.I_block)

			return fmt.Sprintf("------------------------"+
				"MKGRP: grupo creado exitosamente\n "+
				"Grupo '%s' creado", cmd.name), nil

		}

	}

	return "", errors.New("no hay espacio suficiente para agregar el grupo")

}

func parseID(id string) int{
	var i int
	fmt.Sscanf(id, "%d", &i)
	return i

}
