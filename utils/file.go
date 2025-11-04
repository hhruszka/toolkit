//go:build linux

package utils

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)

// FileAbs returns absolute path a file provided in filePath
func FileAbs(filePath string) string {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return filePath
	}
	return absPath
}

// DoesFileExists checks if a file exists at the given path
func DoesFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsRegularFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.Mode().IsRegular()
}

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.Mode().IsDir(), nil
}

func IsExecutable(filePath string) bool {
	var fileInfo fs.FileInfo

	fileInfo, _ = os.Stat(filePath)
	fileStat := fileInfo.Sys().(*syscall.Stat_t)

	// We need to switch to syscall.Getuid() because
	// containers might not have $USER variable set and
	// then user.Current() will fail.
	uid := syscall.Getuid()
	gids, _ := syscall.Getgroups()

	// Otherwise, check 'others' permissions
	if fileInfo.Mode().Perm()&0001 != 0 {
		return true
	}

	// Check if the user is in any of the same groups as the file
	for _, gid := range gids {
		if gid == int(fileStat.Gid) && fileInfo.Mode().Perm()&0010 != 0 {
			return true
		}
	}

	// Check if the user is the owner of the file
	if uid == int(fileStat.Uid) && fileInfo.Mode().Perm()&0100 != 0 {
		return true
	}

	return false
}

// IsReadable checks if a file can be read
func IsReadable(filePath string) bool {
	file, fileOpenErr := os.OpenFile(filePath, os.O_RDONLY, 0)
	defer func() {
		if fileOpenErr == nil {
			_ = file.Close()
		}
	}()

	userInfo, userErr := user.Current()
	if userErr == nil {
		stat, err := os.Stat(filePath)
		if err == nil {
			// the current user is root so let's check if a file is readable by others
			// since root always has access to all system resources.
			if userInfo.Uid == "0" {
				return stat.Mode().Perm()&0004 != 0
			}
			// if the file is owned by a current user then its permissions do not matter since
			// the owner can always change them.
			if userInfo.Uid == fmt.Sprint(stat.Sys().(*syscall.Stat_t).Uid) {
				return true
			}
		}
	}

	// The current user does not own the file and also is not root.
	return fileOpenErr == nil
}

func IsWritable(filePath string) bool {
	file, fileOpenErr := os.OpenFile(filePath, os.O_WRONLY, 0)
	defer func() {
		if fileOpenErr == nil {
			_ = file.Close()
		}
	}()

	userInfo, userErr := user.Current()
	if userErr == nil {
		stat, err := os.Stat(filePath)
		if err == nil {
			// the current user is root so let's check if a file is writable by others
			// since root always has access to all system resources.
			if userInfo.Uid == "0" {
				return stat.Mode().Perm()&0002 != 0
			}
			// if the file is owned by a current user then its permissions do not matter since
			// the owner can always change them.
			if userInfo.Uid == fmt.Sprint(stat.Sys().(*syscall.Stat_t).Uid) {
				return true
			}
		}
	}

	return fileOpenErr == nil
}

func FileOwnership(filePath string) (userName string, groupName string) {
	//Get file information
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		//fmt.Println("Error getting file information:", err)
		return "", ""
	}

	//Extract the syscall.Stat_t structure
	stat := fileInfo.Sys().(*syscall.Stat_t)

	// Get the UID and GID
	uid := stat.Uid
	gid := stat.Gid

	// Convert UID to username
	userInfo, err := user.LookupId(fmt.Sprint(uid))
	if err == nil {
		userName = userInfo.Name
	}

	// Convert GID to group name (optional)
	group, err := user.LookupGroupId(fmt.Sprint(gid))
	if err == nil {
		groupName = group.Name
	}

	return userName, groupName
}

func FilePermissions(filePath string) string {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "----------"
	}

	mode := fileInfo.Mode()
	return fileModeToString(mode)
}

// Convert file mode to a string similar to `ls -l`
func fileModeToString(mode os.FileMode) string {
	perm := []byte("----------")
	if mode.IsDir() {
		perm[0] = 'd'
	}
	if mode&os.ModeSymlink != 0 {
		perm[0] = 'l'
	}
	if mode&os.ModeSocket != 0 {
		perm[0] = 's'
	}
	if mode&os.ModeCharDevice != 0 {
		perm[0] = 'c'
	}
	if mode&0400 != 0 {
		perm[1] = 'r'
	}
	if mode&0200 != 0 {
		perm[2] = 'w'
	}
	if mode&0100 != 0 {
		perm[3] = 'x'
	}
	if mode&0100 != 0 && mode&mode&os.ModeSetuid != 0 {
		perm[3] = 's'
	}
	if mode&0100 == 0 && mode&mode&os.ModeSetuid != 0 {
		perm[3] = 'S'
	}
	if mode&0040 != 0 {
		perm[4] = 'r'
	}
	if mode&0020 != 0 {
		perm[5] = 'w'
	}
	if mode&0010 != 0 {
		perm[6] = 'x'
	}
	if mode&0010 != 0 && mode&mode&os.ModeSetgid != 0 {
		perm[6] = 's'
	}
	if mode&0010 == 0 && mode&mode&os.ModeSetgid != 0 {
		perm[6] = 'S'
	}
	if mode&0004 != 0 {
		perm[7] = 'r'
	}
	if mode&0002 != 0 {
		perm[8] = 'w'
	}
	if mode&0001 != 0 {
		perm[9] = 'x'
	}
	if mode&0001 != 0 && mode&os.ModeSticky != 0 {
		perm[9] = 't'
	}
	return string(perm)
}
