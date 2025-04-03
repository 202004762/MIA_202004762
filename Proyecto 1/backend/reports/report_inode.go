package reports

import (
	structures "backend/structures"
	utils "backend/utils"
	"fmt"
	"os"
	"os/exec"
	"time"
)


func ReportInode(superblock *structures.SuperBlock, diskPath string, path string) error{
	err := utils.CreateParentDirs(path)
	if err != nil{
		return err

	}

	dotFileName, outputImage := utils.GetFileNames(path)
	dotContent := `digraph G{
        node [shape=plaintext]
		rankdir=LR;
    `

	for i := int32(0); i < superblock.S_inodes_count; i++{
		inode := &structures.Inode{}
		err := inode.Deserialize(diskPath, int64(superblock.S_inode_start + (i * superblock.S_inode_size)))
		if err != nil{
			return err

		}

		atime := time.Unix(int64(inode.I_atime), 0).Format(time.RFC3339)
		ctime := time.Unix(int64(inode.I_ctime), 0).Format(time.RFC3339)
		mtime := time.Unix(int64(inode.I_mtime), 0).Format(time.RFC3339)
		dotContent += fmt.Sprintf(`inode%d [label=<
            <table border="1" cellborder="1" cellspacing="0" cellpadding="4">
				<tr><td colspan="2" bgcolor="gray"><font color="white"><b>REPORTE INODO %d</b></font></td></tr>
				<tr><td><b>i_uid</b></td><td>%d</td></tr>
				<tr><td><b>i_gid</b></td><td>%d</td></tr>
				<tr><td><b>i_size</b></td><td>%d</td></tr>
				<tr><td><b>i_atime</b></td><td>%s</td></tr>
				<tr><td><b>i_ctime</b></td><td>%s</td></tr>
				<tr><td><b>i_mtime</b></td><td>%s</td></tr>
				<tr><td><b>i_type</b></td><td>%c</td></tr>
				<tr><td><b>i_perm</b></td><td>%s</td></tr>
				<tr><td colspan="2" bgcolor="lightblue"><b>BLOQUES DIRECTOS</b></td></tr>
            `, i, i, inode.I_uid, inode.I_gid, inode.I_size, atime, ctime, mtime, rune(inode.I_type[0]), string(inode.I_perm[:]))

		for j, block := range inode.I_block{
			if j > 11{
				break

			}

			dotContent += fmt.Sprintf("<tr><td>%d</td><td>%d</td></tr>", j+1, block)

		}

		dotContent += fmt.Sprintf(`
                <tr><td colspan="2" bgcolor="palegreen"><b>BLOQUE INDIRECTO</b></td></tr>
				<tr><td>%d</td><td>%d</td></tr>
				<tr><td colspan="2" bgcolor="lightgreen"><b>BLOQUE INDIRECTO DOBLE</b></td></tr>
				<tr><td>%d</td><td>%d</td></tr>
				<tr><td colspan="2" bgcolor="lightyellow"><b>BLOQUE INDIRECTO TRIPLE</b></td></tr>
				<tr><td>%d</td><td>%d</td></tr>
            </table>>];
        `, 13, inode.I_block[12], 14, inode.I_block[13], 15, inode.I_block[14])
		if i < superblock.S_inodes_count-1{
			dotContent += fmt.Sprintf("inode%d -> inode%d;\n", i, i+1)

		}

	}

	dotContent += "}"
	dotFile, err := os.Create(dotFileName)
	if err != nil{
		return err

	}

	defer dotFile.Close()
	_, err = dotFile.WriteString(dotContent)
	if err != nil{
		return err

	}

	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil{
		return err

	}

	fmt.Println("Imagen de los inodos generada:", outputImage)
	return nil

}
