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
	//fs string

}

func ParseMkfs(tokens []string) (string, error) {
	cmd := &MKFS{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-id=[^\s]+|-type=[^\s]+`)
	matches := re.FindAllString(args, -1)

	if len(matches) != len(tokens) {
		// Identificar el parámetro inválido
		for _, token := range tokens {
			if !re.MatchString(token) {
				return "", fmt.Errorf("parámetro inválido: %s", token)
			}
		}
	}

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
			case "-id":
				if value == "" {
					return "", errors.New("el id no puede estar vacío")
				}
				cmd.id = value
			case "-type":
				if value != "full" {
					return "", errors.New("el tipo debe ser full")
				}
				cmd.typ = value
			/*case "-fs":
				// Verifica que el sistema de archivos sea "2fs" o "3fs"
				if value != "2fs" && value != "3fs" {
					return "", errors.New("el sistema de archivos debe ser 2fs o 3fs")
				}
				cmd.fs = value*/
			default:
				return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.id == "" {
		return "", errors.New("faltan parámetros requeridos: -id")
	}

	if cmd.typ == "" {
		cmd.typ = "full"
	}

	// Si no se proporcionó el sistema de archivos, se establece por defecto a "2fs"
	/*if cmd.fs == "" {
		cmd.fs = "2fs"
	}*/

	err := commandMkfs(cmd)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return fmt.Sprintf("------------------------MKFS: Sistema de archivos creado exitosamente\n"+
		"-> ID: %s\n"+
		"-> Tipo: %s\n"+
		"-> Sistema de archivos: EXT2",
		cmd.id, cmd.typ), nil

}

func commandMkfs(mkfs *MKFS) error{
	mountedPartition, partitionPath, err := stores.GetMountedPartition(mkfs.id)
	if err != nil{
		return err

	}

	n := calculateN(mountedPartition)

	if n <= 0 {
		return fmt.Errorf("no se puede formatear la partición: espacio insuficiente o inválido (n=%d)", n)
	
	}

	//fmt.Println("\nValor de n:", n)

	superBlock := createSuperBlock(mountedPartition, n)
	err = superBlock.CreateBitMaps(partitionPath)
	if err != nil{
		return err

	}

	err = superBlock.CreateUsersFile(partitionPath)
	if err != nil{
		return err

	}

	// Validar que sistema de archivos es
	/*if superBlock.S_filesystem_type == 3 {
		// Crear archivo users.txt ext3
		err = superBlock.CreateUsersFileExt3(partitionPath, int64(mountedPartition.Part_start+int32(binary.Size(structures.SuperBlock{}))))
		if err != nil {
			return err
		}
	} else {
		// Crear archivo users.txt ext2
		err = superBlock.CreateUsersFileExt2(partitionPath)
		if err != nil {
			return err
		}
	}*/

	// Serializar el superbloque
	err = superBlock.Serialize(partitionPath, int64(mountedPartition.Part_start))
	if err != nil{
		return err

	}

	return nil

}

func calculateN(partition *structures.PARTITION) int32{
	/*
		numerador = (partition_montada.size - sizeof(Structs::Superblock)
		denominador base = (4 + sizeof(Structs::Inodes) + 3 * sizeof(Structs::Fileblock))
		n = floor(numerador / denominador)
	*/

	numerator := int(partition.Part_size) - binary.Size(structures.SuperBlock{})
	denominator := 4 + binary.Size(structures.Inode{}) + 3 * binary.Size(structures.FileBlock{})	
	n := math.Floor(float64(numerator) / float64(denominator))

	//fmt.Println("SuperBlock size:", binary.Size(structures.SuperBlock{}))
	//fmt.Println("Partition Size:", partition.Part_size)
	//fmt.Printf("Numerador: %d, Denominador: %d, n: %f\n", numerator, denominator, n)

	return int32(n)

/*
   falta algo aqui, revisar repo para fase 2
*/

}

func createSuperBlock(partition *structures.PARTITION, n int32) *structures.SuperBlock{
	// Calcular punteros de las estructuras
	// Bitmaps
	bm_inode_start := partition.Part_start + int32(binary.Size(structures.SuperBlock{}))
	bm_block_start := bm_inode_start + n // n indica la cantidad de inodos, solo la cantidad para ser representada en un bitmap
	// Inodos
	inode_start := bm_block_start + (3 * n) // 3*n indica la cantidad de bloques, se multiplica por 3 porque se tienen 3 tipos de bloques
	// Bloques
	block_start := inode_start + (int32(binary.Size(structures.Inode{})) * n) // n indica la cantidad de inodos, solo que aquí indica la cantidad de estructuras Inode

	// Crear un nuevo superbloque
	superBlock := &structures.SuperBlock{
		S_filesystem_type:   2,
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