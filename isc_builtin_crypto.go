package goscript

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/rc4"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	h0 "encoding/hex"
	"fmt"
	"strings"
)

func (r *Runtime) builtinCrypto_md5(call FunctionCall) Value {
	s := privateMD5String(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_sha1(call FunctionCall) Value {
	s := privateSha1String(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_sha256(call FunctionCall) Value {
	s := privateSha256String(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_base64Encrypt(call FunctionCall) Value {
	s := privateBase64Encrypt(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_base64Decrypt(call FunctionCall) Value {
	s := privateBase64Decrypt(call.Argument(0).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_aesEncrypt(call FunctionCall) Value {
	s := privateAesEncrypt(call.Argument(0).toString().String(), call.Argument(1).toString().String(), call.Argument(2).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_aesDecrypt(call FunctionCall) Value {
	s := privateAesDecrypt(call.Argument(0).toString().String(), call.Argument(1).toString().String(), call.Argument(2).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_aesEncryptECB(call FunctionCall) Value {
	s := privateAesEncryptECB(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_aesDecryptECB(call FunctionCall) Value {
	s := privateAesDecryptECB(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_aesEncryptCBC(call FunctionCall) Value {
	s := privateAesEncryptCBC(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_aesDecryptCBC(call FunctionCall) Value {
	s := privateAesDecryptCBC(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_desEncryptCBC(call FunctionCall) Value {
	s := privateDESEncryptCBC(call.Argument(0).toString().String(), call.Argument(1).toString().String(), call.Argument(2).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_desDecryptCBC(call FunctionCall) Value {
	s := privateDESDecryptCBC(call.Argument(0).toString().String(), call.Argument(1).toString().String(), call.Argument(2).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_desEncryptECB(call FunctionCall) Value {
	s := privateDESEncryptECB(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_desDecryptECB(call FunctionCall) Value {
	s := privateDESDecryptECB(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_rc4Encrypt(call FunctionCall) Value {
	s := privateRC4Encrypt(call.Argument(0).toString().String(), call.Argument(1).toString().String())
	return r.ToValue(s)
}

func (r *Runtime) builtinCrypto_rc4Decrypt(call FunctionCall) Value {
	s := privateRC4Decrypt(call.Argument(0).toString().String(), call.Argument(1).toString().String())
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

// migrate from gobase

func privateMD5String(s string) string {
	b := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", b)
}

func privateSha1String(s string) string {
	b := sha1.Sum([]byte(s))
	return fmt.Sprintf("%x", b)
}

func privateSha256String(s string) string {
	b := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", b)
}

func privateBase64Encrypt(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}

func privateBase64Decrypt(content string) string {
	b, _ := base64.StdEncoding.DecodeString(content)
	return string(b)
}

func privateAesDecrypt(content string, key string, iv string) string {
	b, _ := base64.StdEncoding.DecodeString(content)
	block, _ := aes.NewCipher([]byte(key))
	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	originData := make([]byte, len(b))
	mode.CryptBlocks(originData, b)
	origData := pkcs5UnPadding(originData)
	return string(origData)
}

func privateAesEncrypt(content string, key string, iv string) string {
	origData := pkcs5Padding([]byte(content), aes.BlockSize)
	block, _ := aes.NewCipher([]byte(key))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	crypted := make([]byte, len(origData))
	mode.CryptBlocks(crypted, origData)
	return base64.StdEncoding.EncodeToString(crypted)
}

// AesDecryptECB 兼容java的AES解密方式
func privateAesDecryptECB(content string, key string) string {
	b, _ := base64.StdEncoding.DecodeString(content)
	cp, _ := aes.NewCipher([]byte(key))
	d := make([]byte, len(b))
	size := 16
	for bs, be := 0, size; bs < len(b); bs, be = bs+size, be+size {
		cp.Decrypt(d[bs:be], b[bs:be])
	}
	return strings.TrimSpace(string(d))
}

// AesEncryptECB 兼容java的AES加密方式
func privateAesEncryptECB(content string, key string) string {
	b := padding([]byte(content))
	cp, _ := aes.NewCipher([]byte(key))
	d := make([]byte, len(b))
	size := 16
	for bs, be := 0, size; bs < len(b); bs, be = bs+size, be+size {
		cp.Encrypt(d[bs:be], b[bs:be])
	}
	return base64.StdEncoding.EncodeToString(d)
}

func privateAesEncryptCBC(content string, key string) string {
	origData := []byte(content)
	k := []byte(key)
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, _ := aes.NewCipher(k)
	blockSize := block.BlockSize()                            // 获取秘钥块的长度
	origData = pkcs5Padding(origData, blockSize)              // 补全码
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize]) // 加密模式
	encrypted := make([]byte, len(origData))                  // 创建数组
	blockMode.CryptBlocks(encrypted, origData)                // 加密
	return h0.EncodeToString(encrypted)
}

func privateAesDecryptCBC(content string, key string) string {
	encrypted, _ := h0.DecodeString(content)
	k := []byte(key)
	block, _ := aes.NewCipher(k)                              // 分组秘钥
	blockSize := block.BlockSize()                            // 获取秘钥块的长度
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize]) // 加密模式
	decrypted := make([]byte, len(encrypted))                 // 创建数组
	blockMode.CryptBlocks(decrypted, encrypted)               // 解密
	decrypted = pkcs5UnPadding(decrypted)                     // 去除补全码
	return string(decrypted)
}

func privateDESEncryptCBC(content string, key string, iv string) string {
	block, _ := des.NewCipher([]byte(key))
	data := pkcs5Padding([]byte(content), block.BlockSize())
	dest := make([]byte, len(data))
	blockMode := cipher.NewCBCEncrypter(block, []byte(iv))
	blockMode.CryptBlocks(dest, data)
	return fmt.Sprintf("%x", dest)
}

func privateDESDecryptCBC(content string, key string, iv string) string {
	b, _ := h0.DecodeString(content)
	block, _ := des.NewCipher([]byte(key))
	blockMode := cipher.NewCBCDecrypter(block, []byte(iv))
	originData := make([]byte, len(b))
	blockMode.CryptBlocks(originData, b)
	origData := pkcs5UnPadding(originData)
	return string(origData)
}

func privateDESEncryptECB(content string, key string) string {
	block, _ := des.NewCipher([]byte(key))
	size := block.BlockSize()
	data := pkcs5Padding([]byte(content), size)
	if len(data)%size != 0 {
		return ""
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Encrypt(dst, data[:size])
		data = data[size:]
		dst = dst[size:]
	}
	return fmt.Sprintf("%x", out)
}

func privateDESDecryptECB(content string, key string) string {
	b, _ := h0.DecodeString(content)
	block, _ := des.NewCipher([]byte(key))
	size := block.BlockSize()
	out := make([]byte, len(b))
	dst := out
	for len(b) > 0 {
		block.Decrypt(dst, b[:size])
		b = b[size:]
		dst = dst[size:]
	}
	return string(pkcs5UnPadding(out))
}

func privateRC4Encrypt(content string, key string) string {
	dest := make([]byte, len(content))
	cp, _ := rc4.NewCipher([]byte(key))
	cp.XORKeyStream(dest, []byte(content))
	return fmt.Sprintf("%x", dest)
}

func privateRC4Decrypt(content string, key string) string {
	b, _ := h0.DecodeString(content)
	dest := make([]byte, len(b))
	cp, _ := rc4.NewCipher([]byte(key))
	cp.XORKeyStream(dest, b)
	return string(dest)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func padding(src []byte) []byte {
	paddingCount := aes.BlockSize - len(src)%aes.BlockSize
	if paddingCount == 0 {
		return src
	} else {
		return append(src, bytes.Repeat([]byte{byte(0)}, paddingCount)...)
	}
}
