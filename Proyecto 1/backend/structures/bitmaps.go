package structures

import (
	"encoding/binary"
	"os"
)


func (sb *SuperBlock) CreateBitMaps(path string) error{
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil{
		return err

	}

	defer file.Close()
	_, err = file.Seek(int64(sb.S_bm_inode_start), 0)
	if err != nil{
		return err

	}

	inodeBitmap := make([]byte, sb.S_free_inodes_count)
	for i := range inodeBitmap{
		inodeBitmap[i] = '0'

	}

	err = binary.Write(file, binary.LittleEndian, inodeBitmap)
	if err != nil{
		return err

	}

	_, err = file.Seek(int64(sb.S_bm_block_start), 0)
	if err != nil{
		return err

	}

	blockBitmap := make([]byte, sb.S_free_blocks_count)
	for i := range blockBitmap{
		blockBitmap[i] = 'O'

	}

	err = binary.Write(file, binary.LittleEndian, blockBitmap)
	if err != nil{
		return err

	}

	return nil

}

func (sb *SuperBlock) UpdateBitmapInodeAt(path string, inodeIndex int32) error{
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil{
		return err

	}

	defer file.Close()
	offset := int64(sb.S_bm_inode_start) + int64(inodeIndex)
	_, err = file.Seek(offset, 0)
	if err != nil{
		return err

	}

	_, err = file.Write([]byte{'1'})
	return err

}

func (sb *SuperBlock) UpdateBitmapBlockAt(path string, blockIndex int32) error{
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil{
		return err

	}

	defer file.Close()
	offset := int64(sb.S_bm_block_start) + int64(blockIndex)
	_, err = file.Seek(offset, 0)
	if err != nil{
		return err

	}

	_, err = file.Write([]byte{'X'})
	return err

}
