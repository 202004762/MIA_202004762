package reports

import (
	"backend/structures"
	"backend/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)


func ReportFile(superblock *structures.SuperBlock, diskPath string, outputPath string, filePath string) error{
	err := utils.CreateParentDirs(outputPath)
	if err != nil{
		return err

	}

	cleanPath := filepath.Clean(filePath)
	parentDirs, fileName := utils.GetParentDirectories(cleanPath)
	var targetInodeIndex int32 = 0
	for _, dir := range parentDirs{
		found := false
		inode := &structures.Inode{}
		err := inode.Deserialize(diskPath, int64(superblock.S_inode_start+targetInodeIndex*superblock.S_inode_size))
		if err != nil{
			return err

		}

		for _, block := range inode.I_block{
			if block == -1{
				continue

			}

			folderBlock := &structures.FolderBlock{}
			err := folderBlock.Deserialize(diskPath, int64(superblock.S_block_start+block*superblock.S_block_size))
			if err != nil{
				return err

			}

			for _, content := range folderBlock.B_content{
				name := strings.Trim(string(content.B_name[:]), "\x00")
				if name == dir{
					targetInodeIndex = content.B_inodo
					found = true
					break

				}

			}

			if found{
				break

			}

		}

		if !found{
			return fmt.Errorf("directorio '%s' no encontrado", dir)

		}

	}

	inode := &structures.Inode{}
	err = inode.Deserialize(diskPath, int64(superblock.S_inode_start+targetInodeIndex*superblock.S_inode_size))
	if err != nil{
		return err

	}

	found := false
	for _, block := range inode.I_block{
		if block == -1{
			continue

		}

		folderBlock := &structures.FolderBlock{}
		err := folderBlock.Deserialize(diskPath, int64(superblock.S_block_start+block*superblock.S_block_size))
		if err != nil{
			return err

		}

		for _, content := range folderBlock.B_content{
			name := strings.Trim(string(content.B_name[:]), "\x00")
			if name == fileName{
				targetInodeIndex = content.B_inodo
				found = true
				break

			}

		}

		if found{
			break

		}

	}

	if !found{
		return fmt.Errorf("archivo '%s' no encontrado", fileName)

	}

	inode = &structures.Inode{}
	err = inode.Deserialize(diskPath, int64(superblock.S_inode_start+targetInodeIndex*superblock.S_inode_size))
	if err != nil{
		return err

	}

	var content strings.Builder
	for _, block := range inode.I_block{
		if block == -1{
			break

		}

		fileBlock := &structures.FileBlock{}
		err := fileBlock.Deserialize(diskPath, int64(superblock.S_block_start+block*superblock.S_block_size))
		if err != nil{
			return err

		}

		text := strings.Trim(string(fileBlock.B_content[:]), "\x00")
		content.WriteString(text)

	}

	reportFile, err := os.Create(outputPath)
	if err != nil{
		return fmt.Errorf("error al crear el archivo de salida: %v", err)

	}

	defer reportFile.Close()
	_, err = reportFile.WriteString(fmt.Sprintf("Archivo: %s\n\nContenido:\n%s\n", fileName, content.String()))
	if err != nil{
		return fmt.Errorf("error al escribir el contenido: %v", err)

	}

	fmt.Println("Reporte de archivo generado en:", outputPath)
	return nil

}
