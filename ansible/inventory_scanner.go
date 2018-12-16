package ansible

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	l "go-slack-ansible/logger"

	"go.uber.org/zap"
	"github.com/joho/godotenv"
)

var DeployPath string
var InventoryPATH string

func InventoryScan(env string) []string {
	var lines []string
	lines = scanrHostsFromfile(setPathForDeploy(env))
	return lines
}

func setPathForDeploy(env string) string {
	if err := godotenv.Load(); err != nil {
		l.Logger.Error("[ERROR] Invalid godotenv", zap.Error(err))
		return "1"
	}

	DeployPath = os.Getenv("ANSIBLE_ROOT_PATH")
	InventoryPATH = os.Getenv("ANSIBLE_INVENTORY_PATH")
	return DeployPath + InventoryPATH
}

func scanrHostsFromfile(filePath string) []string {
	// ファイルを開く
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "File %s could not read: %v\n", filePath, err)
		os.Exit(1)
	}

	// 関数return時に閉じる
	defer f.Close()

	// Scannerで読み込む
	// lines := []string{}
	lines := make([]string, 0, 500) // ある程度行数が事前に見積もれるようであれば、makeで初期capacityを指定して予めメモリを確保しておくことが望ましい
	m := make(map[string]struct{})
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		// appendで追加
		s := scanner.Text()
		if s != "" && strings.HasPrefix(s, "[") {
			s = strings.TrimPrefix(s, "[")
			s = strings.TrimSuffix(s, "]")
			s = strings.TrimSuffix(s, ":children")
			if _, ok := m[s]; !ok {
				m[s] = struct{}{}
				lines = append(lines, s)
			}
		}
	}
	if serr := scanner.Err(); serr != nil {
		fmt.Fprintf(os.Stderr, "File %s scan error: %v\n", filePath, err)
	}

	return lines
}
