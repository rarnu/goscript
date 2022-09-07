package goscript

import (
	f0 "github.com/isyscore/isc-gobase/file"
)

func (r *Runtime) builtinFile_fileExists(call FunctionCall) Value {
	b := f0.FileExists(call.Argument(0).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_directoryExists(call FunctionCall) Value {
	b := f0.DirectoryExists(call.Argument(0).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_extractFilePath(call FunctionCall) Value {
	s := f0.ExtractFilePath(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinFile_extractFileName(call FunctionCall) Value {
	s := f0.ExtractFileName(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinFile_extractFileExt(call FunctionCall) Value {
	s := f0.ExtractFileExt(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinFile_changeFileExt(call FunctionCall) Value {
	s := f0.ChangeFileExt(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinFile_mkdirs(call FunctionCall) Value {
	b := f0.MkDirs(call.Argument(0).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_deleteDirs(call FunctionCall) Value {
	b := f0.DeleteDirs(call.Argument(0).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_deleteFile(call FunctionCall) Value {
	b := f0.DeleteFile(call.Argument(0).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_readFile(call FunctionCall) Value {
	s := f0.ReadFile(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinFile_readFileLines(call FunctionCall) Value {
	s := f0.ReadFileLines(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinFile_writeFile(call FunctionCall) Value {
	b := f0.WriteFile(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_appendFile(call FunctionCall) Value {
	b := f0.AppendFile(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_copyFile(call FunctionCall) Value {
	b := f0.CopyFile(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_renameFile(call FunctionCall) Value {
	b := f0.RenameFile(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.toBoolean(b)
}

func (r *Runtime) builtinFile_dirChild(call FunctionCall) Value {
	e0, err := f0.Child(call.Argument(0).toString().String())
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
	i := f0.Size(call.Argument(0).toString().String())
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
