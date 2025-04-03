package structures

import (
	utils "backend/utils"
	"strings"
	"time"
)


func (sb *SuperBlock) CreateUsersFile(path string) error{
	rootBlockIndex, err := sb.GetFirstFreeBlock(path)
	if err != nil{
		return err

	}

	rootInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{rootBlockIndex, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'},
		I_perm:  [3]byte{'7', '7', '7'},

	}

	err = rootInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil{
		return err

	}

	err = sb.UpdateBitmapInodeAt(path, sb.S_inodes_count)
	if err != nil{
		return err

	}

	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size
	rootBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: 0},
			{B_name: [12]byte{'.', '.'}, B_inodo: 0},
			{B_name: [12]byte{'-'}, B_inodo: -1},
			{B_name: [12]byte{'-'}, B_inodo: -1},

		},

	}

	err = rootBlock.Serialize(path, int64(sb.S_block_start)+int64(rootBlockIndex)*int64(sb.S_block_size))
	if err != nil{
		return err

	}

	err = sb.UpdateBitmapBlockAt(path, rootBlockIndex)
	if err != nil{
		return err

	}

	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	// ----------- Creamos /users.txt -----------
	usersText := "1,G,root\n1,U,root,root,123\n"

	err = rootInode.Deserialize(path, int64(sb.S_inode_start+0))
	if err != nil{
		return err

	}

	rootInode.I_atime = float32(time.Now().Unix())
	err = rootInode.Serialize(path, int64(sb.S_inode_start+0))
	if err != nil{
		return err

	}

	err = rootBlock.Deserialize(path, int64(sb.S_block_start)+int64(rootBlockIndex)*int64(sb.S_block_size))
	if err != nil{
		return err

	}

	rootBlock.B_content[2] = FolderContent{
		B_name:  [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'},
		B_inodo: sb.S_inodes_count,

	}

	err = rootBlock.Serialize(path, int64(sb.S_block_start)+int64(rootBlockIndex)*int64(sb.S_block_size))
	if err != nil{
		return err

	}

	usersBlockIndex, err := sb.GetFirstFreeBlock(path)
	if err != nil{
		return err

	}

	usersInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len(usersText)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{usersBlockIndex, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'1'},
		I_perm:  [3]byte{'7', '7', '7'},

	}

	err = usersInode.Serialize(path, int64(sb.S_first_ino))
	if err != nil{
		return err

	}

	err = sb.UpdateBitmapInodeAt(path, sb.S_inodes_count)
	if err != nil{
		return err

	}

	sb.S_inodes_count++
	sb.S_free_inodes_count--
	sb.S_first_ino += sb.S_inode_size
	usersBlock := &FileBlock{}
	copy(usersBlock.B_content[:], usersText)
	err = usersBlock.Serialize(path, int64(sb.S_block_start)+int64(usersBlockIndex)*int64(sb.S_block_size))
	if err != nil{
		return err

	}

	err = sb.UpdateBitmapBlockAt(path, usersBlockIndex)
	if err != nil{
		return err

	}

	sb.S_blocks_count++
	sb.S_free_blocks_count--
	sb.S_first_blo += sb.S_block_size

	return nil

}

func (sb *SuperBlock) createFolderInInode(path string, inodeIndex int32, parentsDir []string, destDir string) error{
	inode := &Inode{}
	err := inode.Deserialize(path, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil{
		return err

	}

	if inode.I_type[0] == '1'{
		return nil

	}

	for _, blockIndex := range inode.I_block{
		if blockIndex == -1{
			break

		}

		block := &FolderBlock{}
		err := block.Deserialize(path, int64(sb.S_block_start)+int64(blockIndex)*int64(sb.S_block_size))
		if err != nil{
			return err

		}

		for indexContent := 2; indexContent < len(block.B_content); indexContent++{
			content := block.B_content[indexContent]

			if len(parentsDir) != 0{
				if content.B_inodo == -1{
					break

				}

				parentDir, err := utils.First(parentsDir)
				if err != nil{
					return err

				}

				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				parentDirName := strings.Trim(parentDir, "\x00 ")
				if strings.EqualFold(contentName, parentDirName){
					return sb.createFolderInInode(path, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)

				}

			}else{
				if content.B_inodo != -1{
					continue

				}

				newBlockIndex, err := sb.GetFirstFreeBlock(path)
				if err != nil{
					return err

				}

				copy(content.B_name[:], destDir)
				content.B_inodo = sb.S_inodes_count
				block.B_content[indexContent] = content
				err = block.Serialize(path, int64(sb.S_block_start)+int64(blockIndex)*int64(sb.S_block_size))
				if err != nil{
					return err

				}

				folderInode := &Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{newBlockIndex, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'},
					I_perm:  [3]byte{'6', '6', '4'},

				}

				err = folderInode.Serialize(path, int64(sb.S_first_ino))
				if err != nil{
					return err

				}

				err = sb.UpdateBitmapInodeAt(path, sb.S_inodes_count)
				if err != nil{
					return err

				}

				sb.S_inodes_count++
				sb.S_free_inodes_count--
				sb.S_first_ino += sb.S_inode_size
				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: content.B_inodo},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},

					},

				}

				err = folderBlock.Serialize(path, int64(sb.S_block_start)+int64(newBlockIndex)*int64(sb.S_block_size))
				if err != nil{
					return err

				}

				err = sb.UpdateBitmapBlockAt(path, newBlockIndex)
				if err != nil{
					return err

				}

				sb.S_blocks_count++
				sb.S_free_blocks_count--
				sb.S_first_blo += sb.S_block_size

				return nil

			}

		}

	}

	return nil

}
