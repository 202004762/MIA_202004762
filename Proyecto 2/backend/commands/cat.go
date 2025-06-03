package commands

import (
	"backend/stores"
	"backend/structures"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)


type CAT struct{
	files map[int]string

}

func ParseCat(tokens []string) (string, error){
	if !stores.Auth.IsAuthenticated(){
		return "", errors.New("no hay ninguna sesion iniciada")

	}

	cat := &CAT{files: make(map[int]string)}
	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-file(\d+)=\"?([^\s\"]+)\"?`)
	matches := re.FindAllStringSubmatch(args, -1)

	if matches == nil || len(matches) == 0{
		return "", errors.New("no se especificaron archivos para leer")

	}

	for _, match := range matches{
		num, err := strconv.Atoi(match[1])
		if err != nil{
			continue

		}

		cat.files[num] = match[2]

	}

	return cat.Execute()

}

func (c *CAT) Execute() (string, error){
	sb, _, path, err := stores.GetMountedPartitionSuperblock(stores.Auth.PartitionID)
	if err != nil{
		return "", err

	}

	inode := &structures.Inode{}
	offset := int64(sb.S_inode_start)
	err = inode.Deserialize(path, offset)
	if err != nil{
		return "", err

	}

	keys := make([]int, 0, len(c.files))
	for k := range c.files{
		keys = append(keys, k)

	}

	sort.Ints(keys)
	output := "------------------------" +
	"CAT: CAT realizado exitosamente\n"
	for _, k := range keys{
		filePath := c.files[k]
		content, err := readFile(filePath, inode, sb, path)
		if err != nil{
			output += fmt.Sprintf("archivo: %s\nerror: %v\n----------\n", filePath, err)
			continue

		}

		output += fmt.Sprintf("archivo: %s\ncontenido:\n%s\n", filePath, content)

	}	

	return output, nil

}

func readFile(filePath string, root *structures.Inode, sb *structures.SuperBlock, diskPath string) (string, error){
	parts := strings.Split(strings.TrimPrefix(filePath, "/"), "/")
	inode := root
	for _, part := range parts{
		found := false
		for _, blk := range inode.I_block{
			if blk == -1{
				continue

			}

			folderBlock := &structures.FolderBlock{}
			offset := int64(sb.S_block_start) + int64(blk)*64
			err := folderBlock.Deserialize(diskPath, offset)
			if err != nil{
				return "", err

			}

			for _, content := range folderBlock.B_content{
				name := strings.Trim(string(content.B_name[:]), "\x00")
				if name == part{
					if content.B_inodo == -1{
						return "", fmt.Errorf("el archivo %s no existe", part)

					}

					newInode := &structures.Inode{}
					offset := int64(sb.S_inode_start) + int64(content.B_inodo)*int64(binarySize(newInode))
					err = newInode.Deserialize(diskPath, offset)
					if err != nil{
						return "", err

					}

					inode = newInode
					found = true
					break

				}

			}

			if found{
				break

			}

		}

		if !found{
			return "", fmt.Errorf("no se encontró el archivo o carpeta: %s", part)

		}

	}

	if inode.I_type[0] != '1'{
		return "", fmt.Errorf("%s no es un archivo o no se puede leer", filePath)

	}

	content := ""
	for _, blk := range inode.I_block{
		if blk == -1{
			continue

		}

		fileBlock := &structures.FileBlock{}
		offset := int64(sb.S_block_start) + int64(blk)*64
		err := fileBlock.Deserialize(diskPath, offset)
		if err != nil{
			return "", err

		}

		content += strings.Trim(string(fileBlock.B_content[:]), "\x00")
		
	}

	return content, nil

}

func binarySize(v interface{}) int{
	return binary.Size(v)

}