package reports

import (
	"backend/structures"
	"backend/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
)


func ReportDisk(mbr *structures.MBR, diskPath string, reportPath string) error{
	err := utils.CreateParentDirs(reportPath)
	if err != nil{
		return err

	}

	dotFileName, outputImage := utils.GetFileNames(reportPath)
	dotContent := `digraph G {
	node [shape=plaintext];
	structDisco [label=<
	<table border="1" cellborder="1" cellspacing="0" cellpadding="4">
	<tr><td bgcolor="gray"><b>MBR</b></td>
`

	totalSize := float64(mbr.Mbr_size)
	usedSpace := int32(0)
	for _, part := range mbr.Mbr_partitions{
		if part.Part_size <= 0{
			continue

		}

		partType := string(part.Part_type[:])
		partName := strings.Trim(string(part.Part_name[:]), "\x00")
		partStart := part.Part_start
		partSize := part.Part_size
		partPercent := float64(partSize) / totalSize * 100
		if partStart > usedSpace{
			freeSize := partStart - usedSpace
			freePercent := float64(freeSize) / totalSize * 100
			dotContent += fmt.Sprintf(`<td bgcolor="white">Libre<br/>%.2f%%</td>`, freePercent)
			usedSpace += freeSize

		}

		if partType == "E"{
			dotContent += fmt.Sprintf(`<td bgcolor="orange">Extendida<br/>%.2f%%<br/><table border="1" cellborder="1">`, partPercent)
			offset := partStart
			limit := partStart + partSize
			var ebr structures.EBR
			for offset < limit{
				err := ebr.Deserialize(diskPath, int64(offset))
				if err != nil{
					break

				}

				if ebr.Part_s > 0{
					logicalName := strings.TrimRight(string(ebr.Part_name[:]), "\x00")
					logicalSize := ebr.Part_s
					logicalPercent := float64(logicalSize) / totalSize * 100
					dotContent += fmt.Sprintf(`<tr><td bgcolor="lightgreen">%s<br/>%.2f%%</td></tr>`, logicalName, logicalPercent)
					offset += logicalSize

				}else{
					break

				}

				if ebr.Part_next <= 0{
					break

				}

				offset = ebr.Part_next
				
			}

			dotContent += `</table></td>`

		}else{
			dotContent += fmt.Sprintf(`<td bgcolor="lightblue">%s<br/>%.2f%%</td>`, partName, partPercent)

		}

		usedSpace += partSize

	}

	if usedSpace < mbr.Mbr_size{
		freeSize := mbr.Mbr_size - usedSpace
		freePercent := float64(freeSize) / totalSize * 100
		dotContent += fmt.Sprintf(`<td bgcolor="white">Libre<br/>%.2f%%</td>`, freePercent)

	}

	dotContent += "</tr></table>>];\n}"
	file, err := os.Create(dotFileName)
	if err != nil{
		return fmt.Errorf("error al crear archivo DOT: %v", err)

	}

	defer file.Close()
	_, err = file.WriteString(dotContent)
	if err != nil{
		return fmt.Errorf("error al escribir contenido DOT: %v", err)

	}

	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil{
		return fmt.Errorf("error al generar imagen DOT: %v", err)

	}

	fmt.Println("Reporte de disco generado:", outputImage)
	return nil

}
