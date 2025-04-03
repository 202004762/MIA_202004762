package reports

import (
	"backend/structures"
	"backend/utils"
	"fmt"
	"os"
	"os/exec"
	"time"
)


func ReportSuperblock(sb *structures.SuperBlock, diskPath string, reportPath string) error{
	err := utils.CreateParentDirs(reportPath)
	if err != nil{
		return err

	}

	dotFileName, outputImage := utils.GetFileNames(reportPath)
	dotContent := fmt.Sprintf(`digraph G {
    node [shape=plaintext]
    tabla [label=<
    <table border="1" cellborder="1" cellspacing="0" cellpadding="4">
        <tr><td colspan="2" bgcolor="gray"><b><font color="white">REPORTE DE SUPERBLOQUE</font></b></td></tr>

        <tr><td><b>sb_filesystem_type</b></td><td>%d</td></tr>
        <tr><td><b>sb_inodes_count</b></td><td>%d</td></tr>
        <tr><td><b>sb_blocks_count</b></td><td>%d</td></tr>
        <tr><td><b>sb_free_inodes_count</b></td><td>%d</td></tr>
        <tr><td><b>sb_free_blocks_count</b></td><td>%d</td></tr>
        <tr><td><b>sb_mtime</b></td><td>%s</td></tr>
        <tr><td><b>sb_umtime</b></td><td>%s</td></tr>
        <tr><td><b>sb_mnt_count</b></td><td>%d</td></tr>
        <tr><td><b>sb_magic</b></td><td>%d</td></tr>
        <tr><td><b>sb_inode_size</b></td><td>%d</td></tr>
        <tr><td><b>sb_block_size</b></td><td>%d</td></tr>
        <tr><td><b>sb_first_ino</b></td><td>%d</td></tr>
        <tr><td><b>sb_first_blo</b></td><td>%d</td></tr>
        <tr><td><b>sb_bm_inode_start</b></td><td>%d</td></tr>
        <tr><td><b>sb_bm_block_start</b></td><td>%d</td></tr>
        <tr><td><b>sb_inode_start</b></td><td>%d</td></tr>
        <tr><td><b>sb_block_start</b></td><td>%d</td></tr>
    </table>>];
}`, sb.S_filesystem_type,
		sb.S_inodes_count,
		sb.S_blocks_count,
		sb.S_free_inodes_count,
		sb.S_free_blocks_count,
		time.Unix(int64(sb.S_mtime), 0).Format("2006-01-02 15:04"),
		time.Unix(int64(sb.S_umtime), 0).Format("2006-01-02 15:04"),
		sb.S_mnt_count,
		sb.S_magic,
		sb.S_inode_size,
		sb.S_block_size,
		sb.S_first_ino,
		sb.S_first_blo,
		sb.S_bm_inode_start,
		sb.S_bm_block_start,
		sb.S_inode_start,
		sb.S_block_start,
	)

	dotContent += "\n}"
	file, err := os.Create(dotFileName)
	if err != nil{
		return fmt.Errorf("error al crear el archivo DOT: %v", err)

	}

	defer file.Close()
	_, err = file.WriteString(dotContent)
	if err != nil{
		return fmt.Errorf("error al escribir el contenido DOT: %v", err)

	}

	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err = cmd.Run()
	if err != nil{
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)

	}

	fmt.Println("Imagen del reporte de SuperBloque generada en:", outputImage)
	return nil

}
