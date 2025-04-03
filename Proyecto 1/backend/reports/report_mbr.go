package reports

import (
	structures "backend/structures"
	utils "backend/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)


func ReportMBR(mbr *structures.MBR, diskPath string, reportPath string) error{
	err := utils.CreateParentDirs(reportPath)
	if err != nil{
		return err

	}

	dotFileName, outputImage := utils.GetFileNames(reportPath)
	dotContent := fmt.Sprintf(`digraph G {
        node [shape=plaintext]
        tabla [label=<
        <table border="1" cellborder="1" cellspacing="0" cellpadding="4">
            <tr><td colspan="2" bgcolor="gray"><b><font color="white">REPORTE MBR</font></b></td></tr>
            <tr><td><b>mbr_tamano</b></td><td>%d</td></tr>
            <tr><td><b>mbr_fecha_creacion</b></td><td>%s</td></tr>
            <tr><td><b>mbr_disk_signature</b></td><td>%d</td></tr>
            `, mbr.Mbr_size, time.Unix(int64(mbr.Mbr_creation_date), 0), mbr.Mbr_disk_signature)

	for i, part := range mbr.Mbr_partitions{
		if part.Part_size == -1{
			continue

		}

		partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
		partStatus := rune(part.Part_status[0])
		partType := rune(part.Part_type[0])
		partFit := rune(part.Part_fit[0])
		dotContent += fmt.Sprintf(`
				<tr><td colspan="2" bgcolor="lightblue"><b>PARTICIÓN %d</b></td></tr>
				<tr><td>part_status</td><td>%c</td></tr>
				<tr><td>part_type</td><td>%c</td></tr>
				<tr><td>part_fit</td><td>%c</td></tr>
				<tr><td>part_start</td><td>%d</td></tr>
				<tr><td>part_size</td><td>%d</td></tr>
				<tr><td>part_name</td><td>%s</td></tr>
			`, i+1, partStatus, partType, partFit, part.Part_start, part.Part_size, partName)

		if partType == 'E'{
			offset := part.Part_start
			limit := part.Part_start + part.Part_size
			var ebr structures.EBR
			for offset < limit{
				err := ebr.Deserialize(diskPath, int64(offset))
				if err != nil{
					break

				}

				if ebr.Part_s > 0{
					logicalName := strings.TrimRight(string(ebr.Part_name[:]), "\x00")
					logicalFit := rune(ebr.Part_fit[0])
					logicalMount := rune(ebr.Part_mount[0])
					dotContent += fmt.Sprintf(`
                        <tr><td colspan="2" bgcolor="palegreen"><b>PARTICIÓN LÓGICA</b></td></tr>
						<tr><td>part_status</td><td>%c</td></tr>
						<tr><td>part_next</td><td>%d</td></tr>
						<tr><td>part_fit</td><td>%c</td></tr>
						<tr><td>part_start</td><td>%d</td></tr>
						<tr><td>part_size</td><td>%d</td></tr>
						<tr><td>part_name</td><td>%s</td></tr>
                    `, logicalMount, ebr.Part_next, logicalFit, ebr.Part_start, ebr.Part_s, logicalName)

				}

				if ebr.Part_next <= 0{
					break

				}

				offset = ebr.Part_next

			}

		}

	}

	dotContent += "</table>>] }"
	file, err := os.Create(dotFileName)
	if err != nil{
		return fmt.Errorf("error al crear el archivo: %v", err)

	}

	defer file.Close()
	_, err = file.WriteString(dotContent)
	if err != nil{
		return fmt.Errorf("error al escribir en el archivo: %v", err)

	}

	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil{
		return fmt.Errorf("error al ejecutar el comando Graphviz: %v", err)

	}

	fmt.Println("Imagen de la tabla generada:", outputImage)
	return nil

}
