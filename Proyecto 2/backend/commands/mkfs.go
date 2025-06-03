package commands

import (
	stores "backend/stores"
	structures "backend/structures"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)


type MKFS struct{
	id  string
	typ string
	fs string

}

func ParseMkfs(tokens []string) (string, error){
	cmd := &MKFS{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[^\s]+|-type=[^\s]+|-fs=[23]fs`)
	matches := re.FindAllString(args, -1)

	if len(matches) != len(tokens){
		for _, token := range tokens{
			if !re.MatchString(token){
				return "", fmt.Errorf("parámetro inválido: %s", token)

			}

		}

	}

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2{
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)

		}

		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\""){
			value = strings.Trim(value, "\"")

		}

		switch key{
			case "-id":
				if value == ""{
					return "", errors.New("el id no puede estar vacío")

				}

				cmd.id = value

			case "-type":
				if value != "full"{
					return "", errors.New("el tipo debe ser full")

				}

				cmd.typ = value

			case "-fs":
				if value != "2fs" && value != "3fs"{
					return "", errors.New("el sistema de archivos debe ser 2fs o 3fs")

				}

				cmd.fs = value

			default:
				return "", fmt.Errorf("parámetro desconocido: %s", key)

		}

	}

	if cmd.id == ""{
		return "", errors.New("faltan parámetros requeridos: -id")

	}

	if cmd.typ == ""{
		cmd.typ = "full"

	}

	if cmd.fs == ""{
		cmd.fs = "2fs"

	}

	err := commandMkfs(cmd)
	if err != nil{
		fmt.Println("Error:", err)

	}

	return fmt.Sprintf("------------------------MKFS: Sistema de archivos creado exitosamente\n"+
		"-> ID: %s\n"+
		"-> Tipo: %s\n"+
		"-> Sistema de archivos: %s",
		cmd.id, cmd.typ, map[string]string{"2fs": "EXT2", "3fs": "EXT3"}[cmd.fs]), nil

}

func commandMkfs(mkfs *MKFS) error{
	mountedPartition, partitionPath, err := stores.GetMountedPartition(mkfs.id)
	if err != nil{
		return err

	}

	n := calculateN(mountedPartition, mkfs.fs)
	if n <= 0 {
		return fmt.Errorf("no se puede formatear la partición: espacio insuficiente o inválido (n=%d)", n)
	
	}

	//fmt.Println("\nValor de n:", n)

	superBlock := createSuperBlock(mountedPartition, n, mkfs.fs)
	err = superBlock.CreateBitMaps(partitionPath)
	if err != nil{
		return err

	}

	// Validar que sistema de archivos es
	if superBlock.S_filesystem_type == 3{
		// Crear archivo users.txt ext3
		err = superBlock.CreateUsersFileExt3(partitionPath, int64(mountedPartition.Part_start+int32(binary.Size(structures.SuperBlock{}))))
		if err != nil{
			return err

		}

	}else{
		// Crear archivo users.txt ext2
		err = superBlock.CreateUsersFile(partitionPath)
		if err != nil{
			return err

		}

	}

	// Serializar el superbloque
	err = superBlock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil{
		return err

	}

	return nil

}

func calculateN(partition *structures.PARTITION, fs string) int32{
	/*
		numerador = (partition_montada.size - sizeof(Structs::Superblock)
		denominador base = (4 + sizeof(Structs::Inodes) + 3 * sizeof(Structs::Fileblock))
		n = floor(numerador / denominador)
	*/

	numerator := int(partition.Part_size) - binary.Size(structures.SuperBlock{})
	denominator := 4 + binary.Size(structures.Inode{}) + 3 * binary.Size(structures.FileBlock{})	
	temp := 0
	if fs == "3fs"{
		temp = binary.Size(structures.Journal{})

	}

	denominatorbase := denominator + temp
	n := math.Floor(float64(numerator) / float64(denominatorbase))

	//fmt.Println("SuperBlock size:", binary.Size(structures.SuperBlock{}))
	//fmt.Println("Partition Size:", partition.Part_size)
	//fmt.Printf("Numerador: %d, Denominador: %d, n: %f\n", numerator, denominator, n)

	return int32(n)

}

func createSuperBlock(partition *structures.PARTITION, n int32, fs string) *structures.SuperBlock{
	journal_start, bm_inode_start, bm_block_start, inode_start, block_start := calculateStartPositions(partition, fs, n)

	fmt.Printf("Journal Start: %d\n", journal_start)
	fmt.Printf("Bitmap Inode Start: %d\n", bm_inode_start)
	fmt.Printf("Bitmap Block Start: %d\n", bm_block_start)
	fmt.Printf("Inode Start: %d\n", inode_start)
	fmt.Printf("Block Start: %d\n", block_start)

	// Tipo de sistema de archivos
	var fsType int32

	if fs == "2fs"{
		fsType = 2

	}else{
		fsType = 3

	}

	// Crear un nuevo superbloque
	superBlock := &structures.SuperBlock{
		S_filesystem_type:   fsType,
		S_inodes_count:      0,
		S_blocks_count:      0,
		S_free_inodes_count: int32(n),
		S_free_blocks_count: int32(n * 3),
		S_mtime:             float32(time.Now().Unix()),
		S_umtime:            float32(time.Now().Unix()),
		S_mnt_count:         1,
		S_magic:             0xEF53,
		S_inode_size:        int32(binary.Size(structures.Inode{})),
		S_block_size:        int32(binary.Size(structures.FileBlock{})),
		S_first_ino:         inode_start,
		S_first_blo:         block_start,
		S_bm_inode_start:    bm_inode_start,
		S_bm_block_start:    bm_block_start,
		S_inode_start:       inode_start,
		S_block_start:       block_start,

	}

	return superBlock

}

func calculateStartPositions(partition *structures.PARTITION, fs string, n int32) (int32, int32, int32, int32, int32){
	superblockSize := int32(binary.Size(structures.SuperBlock{}))
	journalSize := int32(binary.Size(structures.Journal{}))
	inodeSize := int32(binary.Size(structures.Inode{}))

	// Inicializar posiciones
	// EXT2
	journalStart := int32(0)
	bmInodeStart := partition.Part_start + superblockSize
	bmBlockStart := bmInodeStart + n
	inodeStart := bmBlockStart + (3 * n)
	blockStart := inodeStart + (inodeSize * n)

	// Ajustar para EXT3
	if fs == "3fs"{
		journalStart = partition.Part_start + superblockSize
		bmInodeStart = journalStart + (journalSize * n)
		bmBlockStart = bmInodeStart + n
		inodeStart = bmBlockStart + (3 * n)
		blockStart = inodeStart + (inodeSize * n)

	}

	return journalStart, bmInodeStart, bmBlockStart, inodeStart, blockStart

}