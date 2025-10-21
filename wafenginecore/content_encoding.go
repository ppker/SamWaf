package wafenginecore

import (
	"SamWaf/common/zlog"
	"SamWaf/utils"
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/andybalholm/brotli"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"net/http"
	"strings"
)

// 返回内容前依据情况进行返回压缩数据
func (waf *WafEngine) compressContent(res *http.Response, isStaticAssist bool, inputBytes []byte, encodeType string) (respBytes []byte, err error) {

	// 如果是静态资源响应或资源类型请求，直接返回原始内容
	if isStaticAssist {
		return inputBytes, errors.New("静态资源或资源类型请求，跳过编码转换")
	}

	// 首先检查Content-Type头中是否明确指定了字符集
	contentType := res.Header.Get("Content-Type")
	var encodedBytes []byte = inputBytes
	charsetName := encodeType

	// 如果Content-Type中指定了字符集，需要将UTF-8编码的内容转换回原始编码
	if charsetName == "" && strings.Contains(contentType, "charset=") {

		parts := strings.Split(contentType, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "charset=") {
				charsetName = strings.TrimPrefix(part, "charset=")
				charsetName = strings.Trim(charsetName, `"'`)
				break
			}
		}
	}
	// 根据字符集进行编码转换
	if charsetName != "" && !strings.EqualFold(charsetName, "utf-8") && !strings.EqualFold(charsetName, "utf8") {
		switch strings.ToLower(charsetName) {
		case "gbk", "gb2312":
			// 将UTF-8转换为GBK
			encoder := simplifiedchinese.GBK.NewEncoder()
			encodedBytes, err = encoder.Bytes(inputBytes)
			if err != nil {
				zlog.Warn("编码转换失败(UTF-8 -> GBK): %v", err)
				// 转换失败时使用原始UTF-8内容
				encodedBytes = inputBytes
			}
		// 可以根据需要添加其他编码的支持
		default:
			zlog.Debug("不支持的字符集编码转换: %s，保持UTF-8编码", charsetName)
		}
	}

	// 根据Content-Encoding进行压缩
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		respBytes, err = utils.GZipEncode(encodedBytes)
	case "deflate":
		respBytes, err = utils.DeflateEncode(encodedBytes)
	case "br":
		respBytes, err = utils.BrotliEncode(encodedBytes)
	default:
		respBytes = encodedBytes
	}
	return
}

// 获取原始内容
func (waf *WafEngine) getOrgContent(resp *http.Response, isStaticAssist bool, defaultEncoding string) (cntBytes []byte, encodeType string, err error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return bodyBytes, "", fmt.Errorf("读取原始响应体失败: %v", err)
	}
	charsetNameFinal := ""
	// 重新设置响应体，以便后续处理
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 如果是静态资源响应或资源类型请求，直接返回原始内容
	if isStaticAssist {
		return bodyBytes, "", errors.New("静态资源或资源类型请求，跳过编码转换")
	}

	// 根据内容编码处理压缩
	var bodyReader io.Reader
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		gzipReader, gzipErr := gzip.NewReader(resp.Body)
		if gzipErr != nil {
			zlog.Warn("gzip解压失败: %v", gzipErr)
			// 失败时返回错误
			return bodyBytes, "", fmt.Errorf("gzip解压失败: %v", gzipErr)
		}
		bodyReader = gzipReader
		defer gzipReader.Close()
	case "deflate":
		deflateReader := flate.NewReader(resp.Body)
		bodyReader = deflateReader
		defer deflateReader.Close()
	case "br":
		brotliReader := brotli.NewReader(resp.Body)
		bodyReader = brotliReader
	default:
		bodyReader = resp.Body
	}
	// 创建缓冲读取器
	bufReader := bufio.NewReader(bodyReader)

	// 首先检查Content-Type头中是否明确指定了字符集
	contentType := resp.Header.Get("Content-Type")
	var currentEncoding encoding.Encoding

	if defaultEncoding != "" {
		switch strings.ToLower(defaultEncoding) {
		case "utf-8", "utf8":
			currentEncoding = unicode.UTF8
			charsetNameFinal = "utf-8"
			break
		case "gbk", "gb2312":
			currentEncoding = simplifiedchinese.GBK
			charsetNameFinal = "gbk"
			break
		default:
		}
	} else if strings.Contains(contentType, "charset=") {
		charsetName := ""
		parts := strings.Split(contentType, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "charset=") {
				charsetName = strings.TrimPrefix(part, "charset=")
				charsetName = strings.Trim(charsetName, `"'`)
				break
			}
		}

		// 如果找到了字符集
		if charsetName != "" {
			zlog.Debug("从Content-Type中检测到字符集: %s", charsetName)
			switch strings.ToLower(charsetName) {
			case "utf-8", "utf8":
				currentEncoding = unicode.UTF8
				charsetNameFinal = "utf-8"
				break
			case "gbk", "gb2312":
				currentEncoding = simplifiedchinese.GBK
				charsetNameFinal = "gbk"
				break
			default:
				currentEncoding = nil // 使用自动检测
			}
		}
	}

	// 如果没有从Content-Type中获取到编码或获取的编码不支持，则使用自动检测
	if currentEncoding == nil {
		// 增加检测字节数到1024，提高准确性
		peekBytes, peekErr := bufReader.Peek(1024)
		if peekErr != nil && peekErr != io.EOF {
			return bodyBytes, "", errors.New(fmt.Sprintf("编码检测错误，Peek失败: %v", peekErr))
		} else {
			// 先尝试用系统自带的自动检测
			detectedEncoding, name, certain := charset.DetermineEncoding(peekBytes, contentType)

			if !certain {
				// 自动检测不确定时，尝试从HTML头部meta标签或DOCTYPE中提取编码
				htmlHead := string(peekBytes)
				metaCharset := ""
				// 1. 检查meta标签
				metaIdx := strings.Index(strings.ToLower(htmlHead), "charset=")
				if metaIdx != -1 {
					// 提取charset=后面的内容
					substr := htmlHead[metaIdx+8:]
					// 只取等号后面第一个非字母数字下划线的内容
					for i, c := range substr {
						if !(c == '-' || c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
							metaCharset = substr[:i]
							break
						}
					}
					if metaCharset == "" {
						metaCharset = substr
					}
					metaCharset = strings.Trim(metaCharset, "\"' />")
				}
				// 2. 检查DOCTYPE声明
				if metaCharset == "" && strings.Contains(htmlHead, "xhtml1-transitional.dtd") {
					// XHTML 1.0 通常默认utf-8或iso-8859-1
					metaCharset = "utf-8"
				}
				if metaCharset != "" {
					zlog.Debug("通过HTML头部meta/doctype辅助检测到字符集: %s", metaCharset)
					switch strings.ToLower(metaCharset) {
					case "utf-8", "utf8":
						currentEncoding = unicode.UTF8
						charsetNameFinal = "utf-8"
						break
					case "gbk", "gb2312":
						currentEncoding = simplifiedchinese.GBK
						charsetNameFinal = "gbk"
						break
					default:
						zlog.Debug("meta标签检测到未知字符集: %s，保持UTF-8", metaCharset)
						return bodyBytes, "", errors.New(fmt.Sprintf("编码检测不确定，且未能从HTML头部meta/doctype提取到编码"))
					}
				} else {
					return bodyBytes, "", errors.New(fmt.Sprintf("编码检测不确定，且未能从HTML头部meta/doctype提取到编码"))
				}
			} else {
				zlog.Debug("编码检测确定为: %s", name)
				currentEncoding = detectedEncoding
			}
		}
	}

	// 使用检测到的编码创建转换读取器
	reader := transform.NewReader(bufReader, currentEncoding.NewDecoder())

	// 读取全部内容
	resBodyBytes, readErr := io.ReadAll(reader)
	if readErr != nil {
		zlog.Warn("读取响应体失败: %v", readErr)
		return bodyBytes, "", fmt.Errorf("读取响应体失败: %v", readErr)
	}

	return resBodyBytes, charsetNameFinal, nil
}

// 处理请求的Content-Encoding解压
func (waf *WafEngine) decompressRequestContent(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return nil, nil
	}

	// 读取原始请求体
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("读取请求体失败: %v", err)
	}

	// 重新设置请求体
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 根据Content-Encoding进行解压
	var bodyReader io.Reader
	switch req.Header.Get("Content-Encoding") {
	case "gzip":
		gzipReader, gzipErr := gzip.NewReader(bytes.NewReader(bodyBytes))
		if gzipErr != nil {
			zlog.Warn("请求gzip解压失败: %v", gzipErr)
			return bodyBytes, fmt.Errorf("请求gzip解压失败: %v", gzipErr)
		}
		bodyReader = gzipReader
		defer gzipReader.Close()
	case "deflate":
		deflateReader := flate.NewReader(bytes.NewReader(bodyBytes))
		bodyReader = deflateReader
		defer deflateReader.Close()
	case "br":
		brotliReader := brotli.NewReader(bytes.NewReader(bodyBytes))
		bodyReader = brotliReader
	default:
		// 没有压缩或不支持的压缩格式，直接返回原始内容
		return bodyBytes, nil
	}

	// 读取解压后的内容
	decompressedBytes, readErr := io.ReadAll(bodyReader)
	if readErr != nil {
		zlog.Warn("读取解压后的请求体失败: %v", readErr)
		return bodyBytes, fmt.Errorf("读取解压后的请求体失败: %v", readErr)
	}

	return decompressedBytes, nil
}
