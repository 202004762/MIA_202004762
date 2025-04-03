package reports

import (
	"backend/structures"
	"backend/utils"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)


func ReportLs(superblock *structures.SuperBlock, diskPath string, outputPath string, logicalPath string) error{
	err := utils.CreateParentDirs(outputPath)
	if err != nil{
		return err

	}

	dotFile, imagePath := utils.GetFileNames(outputPath)
	dot := `digraph G {
		node [shape=plaintext]
		structura [label=<
		<table border="1" cellborder="1" cellspacing="0" cellpadding="4">
		<tr>
			<td><b>Permisos</b></td>
			<td><b>Owner</b></td>
			<td><b>Grupo</b></td>
			<td><b>Size (Bytes)</b></td>
			<td><b>Fecha</b></td>
			<td><b>Hora</b></td>
			<td><b>Tipo</b></td>
			<td><b>Name</b></td>
		</tr>
	`

	cleanPath := filepath.Clean(logicalPath)
	parentDirs, _ := utils.GetParentDirectories(cleanPath)
	var currentInode int32 = 0
	for _, dir := range parentDirs{
		inodo := &structures.Inode{}
		err := inodo.Deserialize(diskPath, int64(superblock.S_inode_start+currentInode*superblock.S_inode_size))
		if err != nil{
			return err

		}

		found := false
		for _, blockIndex := range inodo.I_block{
			if blockIndex == -1{
				continue

			}

			block := &structures.FolderBlock{}
			err = block.Deserialize(diskPath, int64(superblock.S_block_start+blockIndex*superblock.S_block_size))
			if err != nil{
				return err

			}

			for _, entry := range block.B_content{
				name := strings.Trim(string(entry.B_name[:]), "\x00")
				if name == dir{
					currentInode = entry.B_inodo
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

	inodo := &structures.Inode{}
	err = inodo.Deserialize(diskPath, int64(superblock.S_inode_start+currentInode*superblock.S_inode_size))
	if err != nil{
		return err

	}

	for _, blockIndex := range inodo.I_block{
		if blockIndex == -1{
			break

		}

		block := &structures.FolderBlock{}
		err = block.Deserialize(diskPath, int64(superblock.S_block_start+blockIndex*superblock.S_block_size))
		if err != nil{
			return err

		}

		for _, content := range block.B_content{
			name := strings.Trim(string(content.B_name[:]), "\x00")
			if name == "" || name == "." || name == ".."{
				continue

			}

			childInode := &structures.Inode{}
			err = childInode.Deserialize(diskPath, int64(superblock.S_inode_start+content.B_inodo*superblock.S_inode_size))
			if err != nil{
				return err

			}

			perm := fmt.Sprintf("-%c%c%c", childInode.I_perm[0], childInode.I_perm[1], childInode.I_perm[2])
			owner := fmt.Sprintf("User%d", childInode.I_uid)
			group := fmt.Sprintf("Grupo%d", childInode.I_gid)
			size := childInode.I_size
			date := time.Unix(int64(childInode.I_ctime), 0).Format("02/01/2006")
			hour := time.Unix(int64(childInode.I_ctime), 0).Format("15:04")
			tipo := "Archivo"
			if childInode.I_type[0] == '0'{
				tipo = "Carpeta"

			}

			dot += fmt.Sprintf(`
				<tr>
					<td>%s</td>
					<td>%s</td>
					<td>%s</td>
					<td>%d</td>
					<td>%s</td>
					<td>%s</td>
					<td>%s</td>
					<td>%s</td>
				</tr>
			`, perm, owner, group, size, date, hour, tipo, name)

		}

	}

	dot += "</table>>]; }"
	f, err := os.Create(dotFile)
	if err != nil{
		return err

	}

	defer f.Close()
	_, err = f.WriteString(dot)
	if err != nil{
		return err

	}

	cmd := exec.Command("dot", "-Tpng", dotFile, "-o", imagePath)
	err = cmd.Run()
	if err != nil{
		return fmt.Errorf("error al generar imagen con Graphviz: %v", err)

	}

	fmt.Println("Reporte LS generado en:", imagePath)
	return nil

}
