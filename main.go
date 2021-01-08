package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	var path string
	flag.StringVar(&path, "path", ".", "版本信息存放目录")
	flag.Parse()

	if !isExists(path) {
		return
	}

	if strings.HasSuffix(path, "/") {
		path = strings.TrimRight(path, "/")
	}

	versionFile := fmt.Sprintf("%s/%s", path, "version")
	version, err := getVersion(versionFile)
	if err != nil {
		log.Fatal(err)
		return
	}

	newVersion := versionAdd(version)
	log.Printf("version change: %s -> %s", version, newVersion)

	if err := saveToFile(versionFile, fmt.Sprintf("version=%s", newVersion)); err != nil {
		log.Fatal(err)
	}

	versionGoFile := fmt.Sprintf("%s/%s", path, "version.go")
	content := fmt.Sprintf(`package info

// 不要动，会被发布脚本自动更新

const (
	Version     = "%s"
	PublishedAt = "%s"
)
`, newVersion, time.Now().Format("2006-01-02 15:04:05"))
	if err := saveToFile(versionGoFile, content); err != nil {
		log.Fatal(err)
	}
}

func saveToFile(file, content string) error {
	return ioutil.WriteFile(file, []byte(content), 0666)
}

func versionAdd(version string) string {
	versionItems := strings.Split(version, ".")
	items := make([]int, 0)
	for _, item := range versionItems {
		n, err := strconv.Atoi(item)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, n)
	}

	tmp := 1
	for i := len(items) - 1; i >= 0; i-- {
		items[i] += tmp
		if i < 1 {
			break
		}

		tmp = items[i] / 1000

		if items[i] >= 1000 {
			items[i] %= 1000
		}

		if tmp == 0 {
			break
		}
	}

	ret := make([]string, 0)
	for _, item := range items {
		ret = append(ret, fmt.Sprintf("%d", item))
	}
	return strings.Join(ret, ".")
}

func getVersion(file string) (string, error) {
	ret := "0.999.999"

	if !isExists(file) {
		return ret, nil
	}

	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()

	reg, _ := regexp.Compile(`^\s*version\s*=\s*(\d+\.\d+\.\d+)\s*$`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		items := reg.FindStringSubmatch(line)
		if len(items) > 0 {
			ret = items[1]
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return ret, nil
}

func isExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		fmt.Printf("%s: %s\n", path, err)
	}
	return err == nil
}
