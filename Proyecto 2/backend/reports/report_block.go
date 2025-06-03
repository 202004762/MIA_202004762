package reports

import (
	structures "backend/structures"
	utils "backend/utils"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strings"
)


func ReportBlock(sb *structures.SuperBlock, diskPath string, outputPath string) error{
	err := utils.CreateParentDirs(outputPath)
	if err != nil{
		return err

	}

	dotFileName, outputImage := utils.GetFileNames(outputPath)
	var dot strings.Builder
	dot.WriteString("digraph G {\n")
	dot.WriteString("  node [shape=record];\n")
	blockCount := int(sb.S_blocks_count)
	for i := 0; i < blockCount; i++{
		offset := int64(sb.S_block_start) + int64(i) * int64(sb.S_block_size)
		fileBlock := &structures.FileBlock{}
		err := fileBlock.Deserialize(diskPath, offset)
		if err == nil{
			content := strings.Trim(string(fileBlock.B_content[:]), "\x00")
			if content != ""{
				label := strings.ReplaceAll(content, "\"", "")
				dot.WriteString(fmt.Sprintf(`  block%d [label="Bloque Archivo %d | %s"]`, i, i, label))
				dot.WriteString(";\n")
				continue

			}

		}

		folderBlock := &structures.FolderBlock{}
		err = folderBlock.Deserialize(diskPath, offset)
		if err == nil{
			entries := ""
			for _, entry := range folderBlock.B_content{
				name := strings.Trim(string(entry.B_name[:]), "\x00")
				if name != "" && entry.B_inodo != -1{
					entries += fmt.Sprintf("%s : %d\\l", name, entry.B_inodo)

				}

			}

			if entries != ""{
				dot.WriteString(fmt.Sprintf(`  block%d [label="Bloque Carpeta %d | %s"]`, i, i, entries))
				dot.WriteString(";\n")
				continue

			}

		}

		var pointerBlock [15]int32
		file, err := os.Open(diskPath)
		if err != nil{
			return err

		}

		defer file.Close()
		_, err = file.Seek(offset, 0)
		if err == nil{
			err = binary.Read(file, binary.LittleEndian, &pointerBlock)
			if err == nil{
				count := 0
				content := ""
				for _, val := range pointerBlock{
					if val != -1{
						content += fmt.Sprintf("%d ", val)
						count++

					}

				}

				if count > 0{
					dot.WriteString(fmt.Sprintf(`  block%d [label="Bloque Apuntadores %d | %s"]`, i, i, content))
					dot.WriteString(";\n")

				}

			}

		}

	}

	dot.WriteString("}")

	err = os.WriteFile(dotFileName, []byte(dot.String()), 0644)
	if err != nil{
		return err

	}

	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil{
		return err

	}

	fmt.Println("Imagen del reporte de bloques generada:", outputImage)
	return nil
	
}
