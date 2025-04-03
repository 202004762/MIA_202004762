package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)


type FolderBlock struct{
	B_content [4]FolderContent // 4 * 16 = 64 bytes

}

type FolderContent struct{
	B_name  [12]byte
	B_inodo int32

}

func (fb *FolderBlock) Serialize(path string, offset int64) error{
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil{
		return fmt.Errorf("error abriendo archivo para escritura: %w", err)

	}

	defer file.Close()
	if _, err = file.Seek(offset, 0); err != nil{
		return fmt.Errorf("error buscando offset %d: %w", offset, err)

	}

	if err = binary.Write(file, binary.LittleEndian, fb); err != nil{
		return fmt.Errorf("error escribiendo FolderBlock: %w", err)

	}

	return nil

}

func (fb *FolderBlock) Deserialize(path string, offset int64) error{
	file, err := os.Open(path)
	if err != nil{
		return fmt.Errorf("error abriendo archivo para lectura: %w", err)

	}

	defer file.Close()
	if _, err = file.Seek(offset, 0); err != nil{
		return fmt.Errorf("error buscando offset %d: %w", offset, err)

	}

	size := binary.Size(fb)
	if size <= 0{
		return fmt.Errorf("tamaño inválido de FolderBlock: %d", size)

	}

	buffer := make([]byte, size)
	if _, err = file.Read(buffer); err != nil{
		return fmt.Errorf("error leyendo FolderBlock: %w", err)

	}

	reader := bytes.NewReader(buffer)
	if err = binary.Read(reader, binary.LittleEndian, fb); err != nil{
		return fmt.Errorf("error deserializando FolderBlock: %w", err)

	}

	return nil

}

func (fb *FolderBlock) Print(){
	for i, content := range fb.B_content{
		name := strings.Trim(string(content.B_name[:]), "\x00 ")
		fmt.Printf("Content %d:\n", i+1)
		fmt.Printf("  B_name: %s\n", name)
		fmt.Printf("  B_inodo: %d\n", content.B_inodo)

	}

}
