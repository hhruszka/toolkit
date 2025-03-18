//go:build linux

package utils

import (
	"path/filepath"
)

func IsFileListed(file string, list []string) bool {
	found := false
	for _, entry := range list {
		if filepath.Base(file) == filepath.Base(entry) {
			found = true
			break
		}
	}

	return found
}

//func IsWritable(path string) bool {
//	if fh, err := os.OpenFile(path, os.O_RDWR, 0); err == nil {
//		_ = fh.Close()
//		return true
//	}
//	return false
//}
//
//func IsReadable(path string) bool {
//	if fh, err := os.OpenFile(path, os.O_RDONLY, 0); err == nil {
//		_ = fh.Close()
//		return true
//	}
//	return false
//}
//
//func IsReadableByOthers(path string) bool {
//
//	info, err := os.Stat(path)
//
//	if err != nil {
//		return false
//	}
//
//	// Check if the file is readable by others - note that 0044 is octal (in this case 9 bits octal digit)
//	return info.Mode().Perm()&0044 != 0
//}
//
//// IsAccessible checks if the given path is accessible.
//func IsAccessible(path string) bool {
//	file, err := os.Open(path)
//	if err != nil {
//		return false
//	}
//	_ = file.Close()
//	return true
//}
//
//func IsExecutable(file any) bool {
//	var fileInfo fs.FileInfo
//
//	if path, ok := file.(string); ok {
//		fileInfo, _ = os.Stat(path)
//	} else {
//		if fileInfo, ok = file.(fs.FileInfo); !ok {
//			fmt.Printf("Internal error: function provided with unsupported type of a parameter %T. Aborting\n", file)
//			os.Exit(1)
//		}
//	}
//
//	fileStat := fileInfo.Sys().(*syscall.Stat_t)
//
//	// We need to switch to syscall.Getuid() because
//	// containers might not have $USER variable set and
//	// then user.Current() will fail.
//	uid := syscall.Getuid()
//	gids, _ := syscall.Getgroups()
//
//	// Check if the user is the owner of the file
//	if uid == int(fileStat.Uid) {
//		return fileInfo.Mode().Perm()&0100 != 0
//	}
//
//	// Check if the user is in any of the same groups as the file
//	for _, gid := range gids {
//		if gid == int(fileStat.Gid) {
//			return fileInfo.Mode().Perm()&0010 != 0
//		}
//	}
//
//	// Otherwise, check 'others' permissions
//	return fileInfo.Mode().Perm()&0001 != 0
//}
//
//type PathType int
//
//const (
//	Binary PathType = iota
//	Java
//	ShellScript
//	Python
//	Perl
//	Ruby
//	Directory
//	Other
//)
//
//var MimeTypeMapping map[string]PathType = map[string]PathType{"application/x-pie-executable": Binary, "application/x-executable": Binary, "application/x-perl": Perl, "text/x-python": Python, "text/x-python3": Python, "application/x-python-code": Python, "text/x-ruby": Ruby, "application/x-ruby": Ruby, "application/x-shellscript": ShellScript, "application/x-java-archive": Java, "inode/directory": Directory}
//
//func identifyPathType(path string) PathType {
//	// mimemagic does not recognize non-readable directories as directories, therefore this plug helps to fix it.
//	if ok, _ := IsDirectory(path); ok {
//		return Directory
//	}
//	mimeType, _ := mimemagic.MatchFilePath(path, -1)
//	if value, ok := MimeTypeMapping[mimeType.MediaType()]; ok {
//		return value
//	}
//	return Other
//}
//
//func IsPythonScript(file string) bool {
//	mimeType, _ := mimemagic.MatchFilePath(file, -1)
//	return IsFileListed(mimeType.MediaType(), []string{"text/x-python", "text/x-python3"})
//}
//
//func IsShellScript(file string) bool {
//	mimeType, _ := mimemagic.MatchFilePath(file, -1)
//	return IsFileListed(mimeType.MediaType(), []string{"application/x-shellscript"})
//}
//
//func IsPerlScript(file string) bool {
//	mimeType, _ := mimemagic.MatchFilePath(file, -1)
//	return IsFileListed(mimeType.MediaType(), []string{"application/x-perl"})
//}
//
//func IsScript(file string) (bool, error) {
//	mimeType, err := mimemagic.MatchFilePath(file, -1)
//	if err != nil {
//		return false, err
//	} else {
//		_, found := MimeTypeMapping[mimeType.MediaType()]
//		return found, nil
//	}
//}
