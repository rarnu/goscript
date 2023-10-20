package goscript

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (r *Runtime) builtinFile_fileExists(call FunctionCall) Value {
	b := privateFileExists(call.Argument(0).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_directoryExists(call FunctionCall) Value {
	b := privateDirectoryExists(call.Argument(0).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_extractFilePath(call FunctionCall) Value {
	s := privateExtractFilePath(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinFile_extractFileName(call FunctionCall) Value {
	s := privateExtractFileName(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinFile_extractFileExt(call FunctionCall) Value {
	s := privateExtractFileExt(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinFile_changeFileExt(call FunctionCall) Value {
	s := privateChangeFileExt(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinFile_mkdirs(call FunctionCall) Value {
	b := privateMkDirs(call.Argument(0).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_deleteDirs(call FunctionCall) Value {
	b := privateDeleteDirs(call.Argument(0).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_deleteFile(call FunctionCall) Value {
	b := privateDeleteFile(call.Argument(0).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_readFile(call FunctionCall) Value {
	s := privateReadFile(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinFile_readFileLines(call FunctionCall) Value {
	s := privateReadFileLines(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinFile_writeFile(call FunctionCall) Value {
	b := privateWriteFile(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_appendFile(call FunctionCall) Value {
	b := privateAppendFile(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_copyFile(call FunctionCall) Value {
	b := privateCopyFile(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_renameFile(call FunctionCall) Value {
	b := privateRenameFile(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_dirChild(call FunctionCall) Value {
	e0, err := privateChild(call.Argument(0).toString().String())
	if err != nil {
		return _null
	} else {
		var ret0 []map[string]any
		for _, item := range e0 {
			_m := map[string]any{
				"name":  item.Name(),
				"isdir": item.IsDir(),
			}
			ret0 = append(ret0, _m)
		}
		return r.ToValue(ret0)
	}
}

func (r *Runtime) builtinFile_fileSize(call FunctionCall) Value {
	i := privateSize(call.Argument(0).toString().String())
	return intToValue(i)
}

func (r *Runtime) initFile() {
	File := r.newBaseObject(r.global.ObjectPrototype, "File")
	File._putProp("fileExists", r.newNativeFunc(r.builtinFile_fileExists, nil, "fileExists", nil, 1), true, false, true)
	File._putProp("directoryExists", r.newNativeFunc(r.builtinFile_directoryExists, nil, "directoryExists", nil, 1), true, false, true)
	File._putProp("extractFilePath", r.newNativeFunc(r.builtinFile_extractFilePath, nil, "extractFilePath", nil, 1), true, false, true)
	File._putProp("extractFileName", r.newNativeFunc(r.builtinFile_extractFileName, nil, "extractFileName", nil, 1), true, false, true)
	File._putProp("extractFileExt", r.newNativeFunc(r.builtinFile_extractFileExt, nil, "extractFileExt", nil, 1), true, false, true)
	File._putProp("changeFileExt", r.newNativeFunc(r.builtinFile_changeFileExt, nil, "changeFileExt", nil, 2), true, false, true)
	File._putProp("mkdirs", r.newNativeFunc(r.builtinFile_mkdirs, nil, "mkdirs", nil, 1), true, false, true)
	File._putProp("deleteDirs", r.newNativeFunc(r.builtinFile_deleteDirs, nil, "deleteDirs", nil, 1), true, false, true)
	File._putProp("deleteFile", r.newNativeFunc(r.builtinFile_deleteFile, nil, "deleteFile", nil, 1), true, false, true)
	File._putProp("readFile", r.newNativeFunc(r.builtinFile_readFile, nil, "readFile", nil, 1), true, false, true)
	File._putProp("readFileLines", r.newNativeFunc(r.builtinFile_readFileLines, nil, "readFileLines", nil, 1), true, false, true)
	File._putProp("writeFile", r.newNativeFunc(r.builtinFile_writeFile, nil, "writeFile", nil, 2), true, false, true)
	File._putProp("appendFile", r.newNativeFunc(r.builtinFile_appendFile, nil, "appendFile", nil, 2), true, false, true)
	File._putProp("copyFile", r.newNativeFunc(r.builtinFile_copyFile, nil, "copyFile", nil, 2), true, false, true)
	File._putProp("renameFile", r.newNativeFunc(r.builtinFile_renameFile, nil, "renameFile", nil, 2), true, false, true)
	File._putProp("dirChild", r.newNativeFunc(r.builtinFile_dirChild, nil, "dirChild", nil, 1), true, false, true)
	File._putProp("fileSize", r.newNativeFunc(r.builtinFile_fileSize, nil, "fileSize", nil, 1), true, false, true)
	r.addToGlobal("File", File.val)
}

// migrate from gobase

func privateFileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		return os.IsExist(err)
	}
	return true
}

func privateDirectoryExists(dirPath string) bool {
	if s, err := os.Stat(dirPath); err != nil {
		return false
	} else {
		return s.IsDir()
	}
}

func privateExtractFilePath(filePath string) string {
	idx := strings.LastIndex(filePath, string(os.PathSeparator))
	return filePath[:idx]
}

func privateExtractFileName(filePath string) string {
	idx := strings.LastIndex(filePath, string(os.PathSeparator))
	return filePath[idx+1:]
}

func privateExtractFileExt(filePath string) string {
	idx := strings.LastIndex(filePath, ".")
	if idx != -1 {
		return filePath[idx+1:]
	}
	return ""
}

func privateChangeFileExt(filePath string, ext string) string {
	mext := privateExtractFileExt(filePath)
	if mext == "" {
		return filePath + "." + ext
	} else {
		return filePath[:len(filePath)-len(mext)] + ext
	}
}

func privateMkDirs(path string) bool {
	if !privateDirectoryExists(path) {
		return os.MkdirAll(path, os.ModePerm) == nil
	} else {
		return false
	}
}

func privateDeleteDirs(path string) bool {
	return os.RemoveAll(path) == nil
}

func privateDeleteFile(filePath string) bool {
	return os.Remove(filePath) == nil
}

func privateReadFile(filePath string) string {
	var ret = ""
	if b, err := os.ReadFile(filePath); err == nil {
		ret = string(b)
	}
	return ret
}

func privateReadFileLines(filePath string) []string {
	var ret []string
	if b, err := os.ReadFile(filePath); err == nil {
		ret = strings.Split(string(b), "\n")
	}
	return ret
}

func privateWriteFile(filePath string, text string) bool {
	return privateWriteFileBytes(filePath, []byte(text))
}

func privateWriteFileBytes(filePath string, data []byte) bool {
	p0 := privateExtractFilePath(filePath)
	if !privateDirectoryExists(p0) {
		privateMkDirs(p0)
	}
	if privateFileExists(filePath) {
		privateDeleteFile(filePath)
	}
	if fl, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return false
	} else {
		_, err := fl.Write(data)
		_ = fl.Close()
		return err == nil
	}
}

func privateAppendFile(filePath string, text string) bool {
	return privateAppendFileBytes(filePath, []byte(text))
}

func privateAppendFileBytes(filePath string, data []byte) bool {
	p0 := privateExtractFilePath(filePath)
	if !privateDirectoryExists(p0) {
		privateMkDirs(p0)
	}
	if fl, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return false
	} else {
		_, err := fl.Write(data)
		_ = fl.Close()
		return err == nil
	}
}

func privateCopyFile(srcFilePath string, destFilePath string) bool {
	p0 := privateExtractFilePath(destFilePath)
	if !privateDirectoryExists(p0) {
		privateMkDirs(p0)
	}
	src, _ := os.Open(srcFilePath)
	defer func(src *os.File) { _ = src.Close() }(src)
	dst, _ := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	defer func(dst *os.File) { _ = dst.Close() }(dst)
	_, err := io.Copy(dst, src)
	return err == nil
}

func privateRenameFile(srcFilePath string, destFilePath string) bool {
	p0 := privateExtractFilePath(destFilePath)
	if !privateDirectoryExists(p0) {
		privateMkDirs(p0)
	}
	return os.Rename(srcFilePath, destFilePath) == nil
}

func privateChild(filePath string) ([]os.DirEntry, error) {
	return os.ReadDir(filePath)
}

// Size 返回文件/目录的大小
func privateSize(filePath string) int64 {
	if !privateDirectoryExists(filePath) {
		fi, err := os.Stat(filePath)
		if err == nil {
			return fi.Size()
		}
		return 0
	} else {
		var size int64
		err := filepath.Walk(filePath, func(_ string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				size += info.Size()
			}
			return err
		})
		if err != nil {
			return 0
		}
		return size
	}
}
