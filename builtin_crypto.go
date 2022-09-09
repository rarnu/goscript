package goscript

import (
	"github.com/isyscore/isc-gobase/coder"
)

func (r *Runtime) builtinCrypto_md5(call FunctionCall) Value {
	s := coder.MD5String(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_sha1(call FunctionCall) Value {
	s := coder.Sha1String(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_sha256(call FunctionCall) Value {
	s := coder.Sha256String(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_base64Encrypt(call FunctionCall) Value {
	s := coder.Base64Encrypt(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_base64Decrypt(call FunctionCall) Value {
	s := coder.Base64Decrypt(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_aesEncrypt(call FunctionCall) Value {
	s := coder.AesEncrypt(call.Argument(0).toString().String(), call.Argument(1).toString().String(), call.Argument(2).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_aesDecrypt(call FunctionCall) Value {
	s := coder.AesDecrypt(call.Argument(0).toString().String(), call.Argument(1).toString().String(), call.Argument(2).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_aesEncryptECB(call FunctionCall) Value {
	s := coder.AesEncryptECB(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_aesDecryptECB(call FunctionCall) Value {
	s := coder.AesDecryptECB(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_aesEncryptCBC(call FunctionCall) Value {
	s := coder.AesEncryptCBC(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_aesDecryptCBC(call FunctionCall) Value {
	s := coder.AesDecryptCBC(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_desEncryptCBC(call FunctionCall) Value {
	s := coder.DESEncryptCBC(call.Argument(0).toString().String(), call.Argument(1).toString().String(), call.Argument(2).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_desDecryptCBC(call FunctionCall) Value {
	s := coder.DESDecryptCBC(call.Argument(0).toString().String(), call.Argument(1).toString().String(), call.Argument(2).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_desEncryptECB(call FunctionCall) Value {
	s := coder.DESEncryptECB(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_desDecryptECB(call FunctionCall) Value {
	s := coder.DESDecryptECB(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_rc4Encrypt(call FunctionCall) Value {
	s := coder.RC4Encrypt(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_rc4Decrypt(call FunctionCall) Value {
	s := coder.RC4Decrypt(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) initCrypto() {
	Crypto := r.newBaseObject(r.global.ObjectPrototype, "Crypto")
	// hash
	Crypto._putProp("md5", r.newNativeFunc(r.builtinCrypto_md5, nil, "md5", nil, 1), true, false, true)
	Crypto._putProp("sha1", r.newNativeFunc(r.builtinCrypto_sha1, nil, "sha1", nil, 1), true, false, true)
	Crypto._putProp("sha256", r.newNativeFunc(r.builtinCrypto_sha256, nil, "sha256", nil, 1), true, false, true)
	// 对称
	Crypto._putProp("base64Encrypt", r.newNativeFunc(r.builtinCrypto_base64Encrypt, nil, "base64Encrypt", nil, 1), true, false, true)
	Crypto._putProp("base64Decrypt", r.newNativeFunc(r.builtinCrypto_base64Decrypt, nil, "base64Decrypt", nil, 1), true, false, true)
	Crypto._putProp("aesEncrypt", r.newNativeFunc(r.builtinCrypto_aesEncrypt, nil, "aesEncrypt", nil, 3), true, false, true)
	Crypto._putProp("aesDecrypt", r.newNativeFunc(r.builtinCrypto_aesDecrypt, nil, "aesDecrypt", nil, 3), true, false, true)
	Crypto._putProp("aesEncryptECB", r.newNativeFunc(r.builtinCrypto_aesEncryptECB, nil, "aesEncryptECB", nil, 2), true, false, true)
	Crypto._putProp("aesDecryptECB", r.newNativeFunc(r.builtinCrypto_aesDecryptECB, nil, "aesDecryptECB", nil, 2), true, false, true)
	Crypto._putProp("aesEncryptCBC", r.newNativeFunc(r.builtinCrypto_aesEncryptCBC, nil, "aesEncryptCBC", nil, 2), true, false, true)
	Crypto._putProp("aesDecryptCBC", r.newNativeFunc(r.builtinCrypto_aesDecryptCBC, nil, "aesDecryptCBC", nil, 2), true, false, true)
	Crypto._putProp("desEncryptCBC", r.newNativeFunc(r.builtinCrypto_desEncryptCBC, nil, "desEncryptCBC", nil, 3), true, false, true)
	Crypto._putProp("desDecryptCBC", r.newNativeFunc(r.builtinCrypto_desDecryptCBC, nil, "desDecryptCBC", nil, 3), true, false, true)
	Crypto._putProp("desEncryptECB", r.newNativeFunc(r.builtinCrypto_desEncryptECB, nil, "desEncryptECB", nil, 2), true, false, true)
	Crypto._putProp("desDecryptECB", r.newNativeFunc(r.builtinCrypto_desDecryptECB, nil, "desDecryptECB", nil, 2), true, false, true)
	Crypto._putProp("rc4Encrypt", r.newNativeFunc(r.builtinCrypto_rc4Encrypt, nil, "rc4Encrypt", nil, 2), true, false, true)
	Crypto._putProp("rc4Decrypt", r.newNativeFunc(r.builtinCrypto_rc4Decrypt, nil, "rc4Decrypt", nil, 2), true, false, true)

	r.addToGlobal("Crypto", Crypto.val)
}
