package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// UpsertEnvToken は .env ファイルの指定キーを上書き（なければ追加）する。
// 既存の内容（コメント・空行・順序）はそのまま保持する。
// ファイルが存在しない場合は新規作成する（パーミッション 0600）。
func UpsertEnvToken(path, key, value string) error {
	lines, err := readLines(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf(".envの読み込みに失敗: %w", err)
	}

	newLine := fmt.Sprintf(`%s="%s"`, key, value)
	found := false
	prefix := key + "="

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, prefix) {
			lines[i] = newLine
			found = true
			break
		}
	}

	if !found {
		// 末尾が空行でなければ空行を挟む
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
			lines = append(lines, "")
		}
		lines = append(lines, newLine)
	}

	return writeLines(path, lines)
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func writeLines(path string, lines []string) error {
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0600)
}
