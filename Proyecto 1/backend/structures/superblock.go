package structures

import (
	"bytes"
	"backend/utils"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)


type SuperBlock struct{
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_inodes_count int32
	S_free_blocks_count int32
	S_mtime             float32
	S_umtime            float32
	S_mnt_count         int32
	S_magic             int32
	S_inode_size        int32
	S_block_size        int32
	S_first_ino         int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32

}

// Serialize escribe la estructura SuperBlock en un archivo binario en la posición especificada
func (sb *SuperBlock) Serialize(path string, offset int64) error{
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil{
		return err

	}

	defer file.Close()
	_, err = file.Seek(offset, 0)
	if err != nil{
		return err

	}

	return binary.Write(file, binary.LittleEndian, sb)

}

// Deserialize lee la estructura SuperBlock desde un archivo binario en la posición especificada
func (sb *SuperBlock) Deserialize(path string, offset int64) error{
	file, err := os.Open(path)
	if err != nil{
		return err

	}

	defer file.Close()
	_, err = file.Seek(offset, 0)
	if err != nil{
		return err

	}

	sbSize := binary.Size(sb)
	if sbSize <= 0{
		return fmt.Errorf("invalid SuperBlock size: %d", sbSize)

	}

	buffer := make([]byte, sbSize)
	_, err = file.Read(buffer)
	if err != nil{
		return err

	}

	reader := bytes.NewReader(buffer)
	return binary.Read(reader, binary.LittleEndian, sb)

}

// Print imprime los valores del SuperBlock
func (sb *SuperBlock) Print(){
	mountTime := time.Unix(int64(sb.S_mtime), 0)
	unmountTime := time.Unix(int64(sb.S_umtime), 0)
	fmt.Printf("Filesystem Type: %d\n", sb.S_filesystem_type)
	fmt.Printf("Inodes Count: %d\n", sb.S_inodes_count)
	fmt.Printf("Blocks Count: %d\n", sb.S_blocks_count)
	fmt.Printf("Free Inodes Count: %d\n", sb.S_free_inodes_count)
	fmt.Printf("Free Blocks Count: %d\n", sb.S_free_blocks_count)
	fmt.Printf("Mount Time: %s\n", mountTime.Format(time.RFC3339))
	fmt.Printf("Unmount Time: %s\n", unmountTime.Format(time.RFC3339))
	fmt.Printf("Mount Count: %d\n", sb.S_mnt_count)
	fmt.Printf("Magic: %d\n", sb.S_magic)
	fmt.Printf("Inode Size: %d\n", sb.S_inode_size)
	fmt.Printf("Block Size: %d\n", sb.S_block_size)
	fmt.Printf("First Inode: %d\n", sb.S_first_ino)
	fmt.Printf("First Block: %d\n", sb.S_first_blo)
	fmt.Printf("Bitmap Inode Start: %d\n", sb.S_bm_inode_start)
	fmt.Printf("Bitmap Block Start: %d\n", sb.S_bm_block_start)
	fmt.Printf("Inode Start: %d\n", sb.S_inode_start)
	fmt.Printf("Block Start: %d\n", sb.S_block_start)

}

// PrintInodes imprime todos los inodos
func (sb *SuperBlock) PrintInodes(path string) error{
	fmt.Println("\nInodos\n----------------")
	for i := int32(0); i < sb.S_inodes_count; i++{
		inode := &Inode{}
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil{
			return err

		}

		fmt.Printf("\nInodo %d:\n", i)
		inode.Print()

	}

	return nil

}

// PrintBlocks imprime todos los bloques de archivos y carpetas
func (sb *SuperBlock) PrintBlocks(path string) error{
	fmt.Println("\nBloques\n----------------")
	visited := make(map[int32]bool)

	for i := int32(0); i < sb.S_inodes_count; i++{
		inode := &Inode{}
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil{
			return err

		}

		for _, blockIndex := range inode.I_block{
			if blockIndex == -1 || visited[blockIndex]{
				continue

			}

			visited[blockIndex] = true
			if inode.I_type[0] == '0'{
				block := &FolderBlock{}
				err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil{
					return err

				}

				fmt.Printf("\nBloque %d:\n", blockIndex)
				for j, content := range block.B_content {
					name := strings.Trim(string(content.B_name[:]), "\x00")
					if name != ""{
						fmt.Printf("Content %d:\n  B_name: %s\n  B_inodo: %d\n", j+1, name, content.B_inodo)

					}

				}

			}else if inode.I_type[0] == '1'{
				block := &FileBlock{}
				err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil{
					return err

				}

				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()

			}

		}

	}

	return nil

}

// GetUsersBlock retorna el contenido completo del archivo users.txt
func (sb *SuperBlock) GetUsersBlock(path string) (string, error){
	inode := &Inode{}
	err := inode.Deserialize(path, int64(sb.S_inode_start+(1*sb.S_inode_size)))
	if err != nil{
		return "", err

	}

	if inode.I_type[0] != '1'{
		return "", fmt.Errorf("el inodo no corresponde a un archivo")

	}

	var content string
	for _, blockIndex := range inode.I_block{
		if blockIndex == -1{
			break

		}

		block := &FileBlock{}
		err := block.Deserialize(path, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil{
			return "", err

		}

		blockContent := strings.Trim(string(block.B_content[:]), "\x00")
		if blockContent != ""{
			content += blockContent + "\n"

		}

	}

	if content == ""{
		return "", fmt.Errorf("users.txt sin contenido o no encontrado")

	}

	return content, nil

}

// CreateFolder crea una carpeta, validando si debe buscar entre todos los inodos
func (sb *SuperBlock) CreateFolder(path string, parentsDir []string, destDir string) error{
	if len(parentsDir) == 0{
		return sb.createFolderInInode(path, 0, parentsDir, destDir)

	}

	success := false
	for i := int32(0); i < sb.S_inodes_count; i++{
		err := sb.createFolderInInode(path, i, parentsDir, destDir)
		if err == nil{
			success = true
			break

		}

	}

	if !success{
		return fmt.Errorf("no se pudo crear el directorio '%s'", destDir)

	}

	return nil

}

// GetFirstFreeBlock busca el primer bloque libre usando el bitmap
func (sb *SuperBlock) GetFirstFreeBlock(path string) (int32, error){
	file, err := os.Open(path)
	if err != nil{
		return -1, err

	}

	defer file.Close()
	_, err = file.Seek(int64(sb.S_bm_block_start), 0)
	if err != nil{
		return -1, err

	}

	bitmap := make([]byte, sb.S_blocks_count+sb.S_free_blocks_count)
	_, err = file.Read(bitmap)
	if err != nil{
		return -1, err

	}

	for i, b := range bitmap{
		if b == 'O'{
			return int32(i), nil

		}

	}

	return -1, fmt.Errorf("no hay bloques libres disponibles")

}

func (sb *SuperBlock) CreateFile(path string, parentsDir []string, destFile string, content string) error{
	// Buscar carpeta padre
	var parentInodeIndex int32 = -1
	var parentInode *Inode
	for i := int32(0); i < sb.S_inodes_count; i++{
		inode := &Inode{}
		err := inode.Deserialize(path, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil{
			return err

		}

		if inode.I_type[0] == '0'{
			for _, blk := range inode.I_block{
				if blk == -1{
					break

				}

				block := &FolderBlock{}
				err := block.Deserialize(path, int64(sb.S_block_start+(blk*sb.S_block_size)))
				if err != nil{
					return err

				}

				for _, content := range block.B_content{
					name := strings.Trim(string(content.B_name[:]), "\x00")
					if strings.EqualFold(name, parentsDir[len(parentsDir)-1]){
						parentInodeIndex = content.B_inodo
						parentInode = inode
						break

					}

				}

			}

		}

	}

	if parentInodeIndex == -1{
		return fmt.Errorf("directorio padre no encontrado")

	}

	// Crear inodo para archivo
	fileInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len(content)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'6', '6', '4'},

	}
	fileInodeOffset := int64(sb.S_first_ino)
	err := fileInode.Serialize(path, fileInodeOffset)
	if err != nil{
		return err

	}

	err = sb.UpdateBitmapInodeAt(path, sb.S_inodes_count)
	if err != nil{
		return err

	}

	fileInodeIndex := sb.S_inodes_count
	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size

	// Dividir contenido en chunks de 64
	chunks := utils.SplitStringIntoChunks(content)

	for i, chunk := range chunks{
		blockIndex := sb.S_blocks_count
		fb := &FileBlock{}
		copy(fb.B_content[:], []byte(chunk))
		offset := int64(sb.S_block_start) + int64(blockIndex)*int64(sb.S_block_size)
		err := fb.Serialize(path, offset)
		if err != nil{
			return err

		}

		err = sb.UpdateBitmapBlockAt(path, blockIndex)
		if err != nil{
			return err

		}

		sb.S_blocks_count++
		sb.S_free_blocks_count--
		sb.S_first_blo += sb.S_block_size
		fileInode.I_block[i] = blockIndex

	}

	// Re-escribir el inodo actualizado con los bloques asignados
	err = fileInode.Serialize(path, fileInodeOffset)
	if err != nil{
		return err

	}

	// Agregar entrada a carpeta padre
	for _, blk := range parentInode.I_block{
		if blk == -1{
			break

		}

		folderBlock := &FolderBlock{}
		err := folderBlock.Deserialize(path, int64(sb.S_block_start+(blk*sb.S_block_size)))
		if err != nil{
			return err

		}

		for i, entry := range folderBlock.B_content{
			if entry.B_inodo == -1{
				copy(folderBlock.B_content[i].B_name[:], destFile)
				folderBlock.B_content[i].B_inodo = fileInodeIndex
				err = folderBlock.Serialize(path, int64(sb.S_block_start+(blk*sb.S_block_size)))
				if err != nil{
					return err

				}

				return nil

			}

		}

	}

	return fmt.Errorf("no hay espacio para agregar el archivo a la carpeta padre")

}
