package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
)

func loadConfig(path string) (Config, error) {

	// data, err := os.ReadFile(path)
	/*这里不使用ReadFile()方法，因为如果配置文件过大，是不能将文件一次性读入内存的
	推荐使用OpenFile()方法打开文件，然后使用json.NewDecoder()函数解码文件内容。
	另外OpenFile()方法，能够自定义打开文件的方式，例如只读、读写等，还可以设置文件权限等。
	*/
	file, err := os.OpenFile(path, os.O_RDONLY, 0644) //OpenFile()打开的是文件的句柄，而不是文件内容，所以不会一次性将整个文件读入内存。
	if err != nil {
		return Config{}, fmt.Errorf("读取配置文件失败: %w", err) //这里使用错误的包装功能，将原始错误包装在新的错误中
	}
	defer file.Close() //关闭文件句柄，释放文件资源
	//值得注意的是*os.File类型，它实现了io.Reader接口，也可以直接使用json.NewDecoder()函数解码文件内容。
	//但是*os.File类型没有实现io.ReaderAt接口，所以不能使用json.Unmarshal()函数解码文件内容
	//所以这里可以使用json.NewDecoder()函数解码文件内容。
	var cfg Config
	/*
		这里的bytes.NewReader(data)是为了创建一个新的io.Reader对象,这个对象实现了io.Reader接口
		可以使用Read（）方法逐字节的读取数据，然后使用json.NewDecoder()函数创建一个JSON解码器，最后使用Decode()方法解码数据。
		相比于json.Unmarshal()函数，json.Unmarshal()函数会一次性将整个JSON数据加载到内存中，如果配置文件非常大，
		可能会导致内存占用过高。而使用json.NewDecoder()函数可以逐步读取和解码JSON数据，避免了内存占用过高的问题。
		同时json.Unmarshal没有提供DisallowUnknownFields()方法来禁止未知字段，而json.NewDecoder()提供了这个方法，可以更严格地验证配置文件的结构，确保没有多余的字段存在。
	*/
	decoder := json.NewDecoder(file)
	/*
			type Decoder struct {
		    r       io.Reader  // 传入的 *os.File
		    buf     []byte     // 内部缓冲区（初始为空）
		    scanp   int        // 已扫描到的位置
		    scanned int64      // 已消耗的字节总数
		    scan    scanner    // JSON 状态机
		}
	*/
	decoder.DisallowUnknownFields() //禁止出现Config中没有定义的字段
	/*
	   这里的Decode()方法，每调用一次，就会从文件中读取一个JSON对象，将其解码到cfg结构体中。
	   如果配置文件中存在多个JSON对象，Decode()方法会在第一次调用时成功解码第一个对象
	   但在第二次调用时会尝试解码第二个对象，此时如果第二个对象存在，就会成功解码
	   如果第二个对象不存在，就会返回io.EOF错误。
	*/
	//实际上，这里还是有问题，因为假设配置文件中存在一个超级大的单json对象，那么Decode（）方法就会
	//一次性的读入整个文件到内存，仍然有OOM的风险，正常的做法是使用Token方法逐步解析JSON数据。
	if err := decoder.Decode(&cfg); err != nil { //这里解码第一个JSON对象，如果失败，就会返回错误。
		return Config{}, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if err := ensureSingleJSONDocument(decoder); err != nil { //这里额外检查是否存在多个JSON对象，如果存在，就会返回错误。
		return Config{}, err
	}

	if err := validateAndNormalizeConfig(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func ensureSingleJSONDocument(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("配置文件中存在多余的 JSON 内容")
		}
		return fmt.Errorf("检查配置文件尾部内容失败: %w", err)
	}
	return nil
}

func validateAndNormalizeConfig(cfg *Config) error {
	if len(cfg.Targets) == 0 {
		return errors.New("配置文件中至少需要一个探测目标")
	}
	seenNames := make(map[string]struct{}, len(cfg.Targets)) //这里我使用了一个map来标记出现过的目标名称，防止重复探测
	//这里不能写成for i,target:= range cfg.Targets，因为target是一个值拷贝的副本，
	//对它的修改不会影响到cfg.Targets中的元素，所以需要使用索引来访问和修改原始的目标对象。
	//除非后续再写回去，如cfg.Targets[i] = target，否则对target的修改是没有意义的。
	for i := range cfg.Targets {
		//这里使用防御性编程的思想，格式化和验证配置文件中的每个目标，确保它们符合预期的格式和要求。
		target := &cfg.Targets[i]
		target.Index = i
		target.Name = strings.TrimSpace(target.Name)
		target.Protocol = strings.ToLower(strings.TrimSpace(target.Protocol))
		target.Address = strings.TrimSpace(target.Address)
		target.Expect.Contains = strings.TrimSpace(target.Expect.Contains)
		if target.Name == "" {
			return fmt.Errorf("第 %d 个目标缺少 name", i+1)
		}
		if _, exists := seenNames[target.Name]; exists {
			return fmt.Errorf("目标名称重复: %s", target.Name)
		}
		seenNames[target.Name] = struct{}{} //不占用额外内存的空结构体，作为标记使用

		if target.Protocol == "" {
			return fmt.Errorf("目标 %s 缺少 protocol", target.Name)
		}
		if target.Address == "" {
			return fmt.Errorf("目标 %s 缺少 address", target.Name)
		}
		if target.RetryCount < 0 || target.RetryCount > 3 {
			return fmt.Errorf("目标 %s 的 retry_count 必须在 0 到 3 之间", target.Name)
		}
		switch target.Protocol {
		case "http":
			if err := validateHTTPConfig(*target); err != nil {
				return err
			}
		case "tcp":
			if err := validateTCPConfig(*target); err != nil {
				return err
			}
		default:
			return fmt.Errorf("目标 %s 使用了不支持的 protocol: %s", target.Name, target.Protocol)
		}
	}
	return nil
}

func validateHTTPConfig(target Target) error {
	parsedURL, err := url.ParseRequestURI(target.Address)
	if err != nil {
		return fmt.Errorf("HTTP 目标 %s 的地址不合法: %w", target.Name, err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("HTTP 目标 %s 只能使用 http 或 https 协议", target.Name)
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("HTTP 目标 %s 缺少主机名", target.Name)
	}
	if target.Expect.Connected != nil {
		return fmt.Errorf("HTTP 目标 %s 不能配置 expect.connected", target.Name)
	}
	if target.Expect.StatusCode != nil {
		if *target.Expect.StatusCode < 100 || *target.Expect.StatusCode > 599 {
			return fmt.Errorf("HTTP 目标 %s 的 expect.status_code 必须在 100 到 599 之间", target.Name)
		}
	}
	return nil
}

func validateTCPConfig(target Target) error {
	if _, _, err := net.SplitHostPort(target.Address); err != nil {
		return fmt.Errorf("TCP 目标 %s 的地址必须是 host:port 格式: %w", target.Name, err)
	}
	if target.Expect.StatusCode != nil {
		return fmt.Errorf("TCP 目标 %s 不能配置 expect.status_code", target.Name)
	}
	if target.Expect.Contains != "" {
		return fmt.Errorf("TCP 目标 %s 不能配置 expect.contains", target.Name)
	}
	if target.Expect.Connected != nil && !*target.Expect.Connected {
		return fmt.Errorf("TCP 目标 %s 的 expect.connected 只能为 true", target.Name)
	}
	return nil
}
