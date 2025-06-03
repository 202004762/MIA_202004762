package reports

import (
	structures "backend/structures"
	utils "backend/utils"
	"fmt"
	"os"
	"os/exec"
	"strings"
)


func ReportTree(sb *structures.SuperBlock, diskPath string, outputPath string) error{
	err := utils.CreateParentDirs(outputPath)
	if err != nil{
		return err

	}

	dotFileName, outputImage := utils.GetFileNames(outputPath)
	var dot strings.Builder
	dot.WriteString("digraph G {\n")
	dot.WriteString("node [shape=plaintext fontname=\"Helvetica\"];\n")
	dot.WriteString("rankdir=LR;\n\n")
	visited := make(map[int32]bool)
	err = drawInodeTree(sb, diskPath, &dot, 0, visited)
	if err != nil{
		return err

	}

	dot.WriteString("}\n")
	dotFile, err := os.Create(dotFileName)
	if err != nil{
		return err

	}

	defer dotFile.Close()
	_, err = dotFile.WriteString(dot.String())
	if err != nil{
		return err

	}

	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil{
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)

	}

	fmt.Println("Reporte del árbol generado:", outputImage)
	return nil

}

func drawInodeTree(sb *structures.SuperBlock, path string, dot *strings.Builder, inodeIndex int32, visited map[int32]bool) error{
	if visited[inodeIndex]{
		return nil

	}

	visited[inodeIndex] = true
	inode := &structures.Inode{}
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil{
		return err

	}

	dot.WriteString(fmt.Sprintf(`inode%d [label=<
		<table border="1" cellborder="1" cellspacing="0">
		<tr><td colspan="2" bgcolor="gray"><b>Inodo %d</b></td></tr>
		<tr><td><b>UID</b></td><td>%d</td></tr>
		<tr><td><b>GID</b></td><td>%d</td></tr>
		<tr><td><b>Size</b></td><td>%d</td></tr>
		<tr><td><b>Type</b></td><td>%c</td></tr>
		<tr><td><b>Perm</b></td><td>%s</td></tr>
		`, inodeIndex, inodeIndex, inode.I_uid, inode.I_gid, inode.I_size, inode.I_type[0], string(inode.I_perm[:])))

	for i, blk := range inode.I_block{
		if blk != -1{
			dot.WriteString(fmt.Sprintf("<tr><td><b>Bloque[%d]</b></td><td>%d</td></tr>\n", i, blk))

		}

	}

	dot.WriteString("</table>>];\n")
	for i, blk := range inode.I_block{
		if blk == -1{
			continue

		}

		switch{
			case i <= 11:
				if inode.I_type[0] == '0'{
					folder := &structures.FolderBlock{}
					err := folder.Deserialize(path, int64(sb.S_block_start) + int64(blk)*int64(sb.S_block_size))
					if err != nil{
						continue

					}

					dot.WriteString(fmt.Sprintf(`folder%d [label=<
						<table border="1" cellborder="1" cellspacing="0">
						<tr><td colspan="2" bgcolor="lightblue"><b>FolderBlock %d</b></td></tr>`, blk, blk))
					for _, c := range folder.B_content{
						name := strings.Trim(string(c.B_name[:]), "\x00")
						if name != "" && c.B_inodo != -1{
							dot.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>\n", name, c.B_inodo))

						}

					}

					dot.WriteString("</table>>];\n")
					dot.WriteString(fmt.Sprintf("inode%d -> folder%d;\n", inodeIndex, blk))
					for _, c := range folder.B_content{
						if c.B_inodo != -1 && string(c.B_name[:]) != "." && string(c.B_name[:]) != ".."{
							drawInodeTree(sb, path, dot, c.B_inodo, visited)

						}

					}

				}else{
					file := &structures.FileBlock{}
					err := file.Deserialize(path, int64(sb.S_block_start) + int64(blk)*int64(sb.S_block_size))
					if err != nil{
						continue

					}

					content := strings.Trim(string(file.B_content[:]), "\x00")
					dot.WriteString(fmt.Sprintf(`file%d [label=<
						<table border="1" cellborder="1" cellspacing="0">
						<tr><td bgcolor="yellow"><b>FileBlock %d</b></td></tr>
						<tr><td>%s</td></tr>
						</table>>];`, blk, blk, content))
					dot.WriteString(fmt.Sprintf("\ninode%d -> file%d;\n", inodeIndex, blk))

				}

			case i == 12:
				pointers := make([]int32, 16)
				offset := int64(sb.S_block_start) + int64(blk)*int64(sb.S_block_size)
				file, _ := os.Open(path)
				defer file.Close()
				for j := 0; j < 16; j++{
					buf := make([]byte, 4)
					file.ReadAt(buf, offset+int64(j*4))
					pointers[j] = int32(buf[0]) | int32(buf[1])<<8 | int32(buf[2])<<16 | int32(buf[3])<<24

				}

				dot.WriteString(fmt.Sprintf(`pointer%d [label=<
					<table border="1" cellborder="1" cellspacing="0">
					<tr><td colspan="2" bgcolor="orange"><b>Indirecto Simple %d</b></td></tr>`, blk, blk))
				for i, p := range pointers{
					dot.WriteString(fmt.Sprintf("<tr><td>%d</td><td>%d</td></tr>", i, p))
					if p != -1{
						dot.WriteString(fmt.Sprintf("inode%d -> pointer%d;\npointer%d -> file%d;\n", inodeIndex, blk, blk, p))

					}

				}

				dot.WriteString("</table>>];\n")

		}

	}

	return nil

}
