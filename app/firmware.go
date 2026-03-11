package app

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type FirmwarePath struct {
	ExecutableLinkPath string
	BinaryDir          string
	DefaultExecutableFileName string
}

var firmwarePaths = map[string]FirmwarePath{
	"A9E": {
		ExecutableLinkPath: "/vol_data/firmware/etos",	// 実行ファイル(シンボリックリンク)のファイルパス
		BinaryDir:          "/vol_data/firmware/bin/",	// 実行ファイル(実体)の保存先ディレクトリ
		DefaultExecutableFileName: "default",	// 出荷時のデフォルトファームウェアのファイル名
	},
	"G3L": {
		ExecutableLinkPath: "/root/etos",	// 実行ファイル(シンボリックリンク)のファイルパス
		BinaryDir:          "/srv/bin/",	// 実行ファイル(実体)の保存先ディレクトリ
		DefaultExecutableFileName: "default",	// 出荷時のデフォルトファームウェアのファイル名
	},
}

const (
	// 12Mbyte を 128kbps で転送を考慮した時間
	timeoutMinute = 30 * time.Minute
)

var FWUpdateFlag bool = false
var CurrentFWName string = ""

func GetFirmwarePath(model string) FirmwarePath {
	return firmwarePaths[model]
}

/*
loadFirmware ファームウェアの更新処理

1. クラウドからファームウェアをダウンロード、ファイルとして保存
TODO チェックサムの確認
2. バージョンアップの必要があるか確認する
正常なファームウェアが存在　かつ　シンボリックリンクの情報が正しい -> なにもしない
バージョンアップ後のファームウェアが存在しない、または、チェックサムに不整合がある場合、バージョンアップの必要性があると判断する
必要がない場合 -> なにもしない
必要がある場合 -> 以下の更新処理を行う
3. 実行ファイル(シンボリックリンクのリンク先)を最新のファームウェアに置き換え
4. 古いファームウェアを削除
デフォルトのfirmware, 1世代前のファームウェア, 最新のファームウェア 以外を削除
*/
func loadFirmware() error {

	firmwarePath := GetFirmwarePath(MODEL)
	oldFilePath, err := os.Readlink(firmwarePath.ExecutableLinkPath)
	if err != nil {
		return fmt.Errorf("failed to os.Readlink:%v", err)
	}
	oldFileName := strings.Replace(oldFilePath, firmwarePath.BinaryDir, "", 1) // ディレクトリ部分を取り除く
	CurrentFWName = oldFileName

	// 1. クラウドからファームウェアをダウンロード、ファイルとして保存
	firmware := database.FirmwareUpdateInfoReadDB()
	var newFileName string
	if (strings.Contains(firmware.Name, "A9E") && MODEL == "A9E") || (strings.Contains(firmware.Name, "G3L") && MODEL == "G3L") {
		newFileName = firmware.Name
	} else {
		// ファームウェアがA9E用でない場合、現状使用しているファームウェアを使用する
		newFileName = oldFileName
	}
	if oldFileName != newFileName {
		FWUpdateFlag = true
		Logger.Writef(LOG_LEVEL_INFO, "FWUpdateFlag is true(loadFirmware). Firmware is updated: %s -> %s", oldFileName, newFileName)
		bytes, err := downloadFirmware(firmware.URL)
		if err != nil {
			return fmt.Errorf("failed to downloadFirmware:%v", err)
		}

		newFilePath := filepath.Join(firmwarePath.BinaryDir, newFileName)
		if err := saveFileIfNotExists(newFilePath, bytes); err != nil {
			return fmt.Errorf("failed to saveFileIfNotExists:%v", err)
		}
	}

	// 2. バージョンアップの必要があるか確認する
	if oldFileName == newFileName {
		Logger.Writef(LOG_LEVEL_INFO, "Firmware is latest version: %s", newFileName)
		return nil
	}

	Logger.Writef(LOG_LEVEL_DEBUG, "Need to update firmware. version: %s -> version: %s",
		oldFileName,
		newFileName,
	)
	return updateFirmware(oldFileName, newFileName)
}

// ファームウェアのダウンロード処理
func downloadFirmware(url string) ([]byte, error) {
	client := http.Client{
		Timeout: timeoutMinute,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// ファームウェアの更新処理
func updateFirmware(oldFileName, newFileName string) error {

	// 3. 実行ファイル(シンボリックリンクのリンク先)を置き換え
	firmwarePath := GetFirmwarePath(MODEL)
	newFilePath := filepath.Join(firmwarePath.BinaryDir, newFileName)
	if err := overwriteSymlink(firmwarePath.ExecutableLinkPath, newFilePath); err != nil {
		return fmt.Errorf("failed to overwriteSymlink:%v", err)
	}

	// 4. 古いファームウェアを削除
	// デフォルトのfirmware, 1世代前のファームウェア, 最新のファームウェア 以外を削除
	files, err := getFileNames(firmwarePath.BinaryDir)
	if err != nil {
		return fmt.Errorf("failed to getFileNames:%v", err)
	}
	for _, filename := range files {
		if filename != firmwarePath.DefaultExecutableFileName &&
			filename != oldFileName &&
			filename != newFileName {
			if err := os.Remove(filepath.Join(firmwarePath.BinaryDir, filename)); err != nil {
				// 古いファームウェアの削除失敗時はエラーログのみ出力する
				Logger.Writef(LOG_LEVEL_ERR, "Failed tp delete old firmware. version: %s", filename)
				continue
			}
			Logger.Writef(LOG_LEVEL_INFO, "Delete old firmware. version: %s", filename)
		}
	}

	Logger.Writef(LOG_LEVEL_INFO, "Firmware updated %s -> %s. You need restart to apply.", oldFileName, newFileName)
	return nil
}

// 引数で指定したパス内のファイル名一覧を取得
func getFileNames(filepath string) ([]string, error) {
	fp, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	files, err := fp.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	if err := fp.Close(); err != nil {
		return nil, err
	}
	return files, nil
}

/*
validFileExists ファイルが存在 かつ 整合性の取れたファイル(チェックサムを使って確認)なら true
それ以外はすべてfalse
*/
func validFileExists(filepath, md5sum string) bool {
	fp, err := os.Open(filepath)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	defer fp.Close()

	bytes, err := io.ReadAll(fp)
	if err != nil {
		return false
	}
	if getMd5(bytes) != md5sum {
		return false
	}
	return true
}

/*
saveFileIfNotExists ファイルが存在しない場合に、引数のファイルに対して引数で指定したバイト列のデータを書き込む
書き込むデータは実行バイナリを想定しているので、7xxで書き込む(実行権限が必要)
OpenFileのついでにファイルの存在を確認するのが無駄なファイルのOpenがなくなり理想なのだが、
書き込みフラグが含まれるため、 "text file busy" のエラーが発生してしまう。
一応 error の文字列に "text file busy" が含まれるかどうかでチェックはできないこともないが、
起動時一回のみでパフォーマンスの面でそこまで問題にならないと判断し、
読み込み専用で一度ファイルを開けるかチェックするようにした
*/
func saveFileIfNotExists(filepath string, bytes []byte) error {
	binary, md5sum, err := getBinaryAndChecksum(bytes)
	if err != nil {
		return fmt.Errorf("failed to getBinaryAndChecksum:%v", err)
	}
	// チェックサムの確認
	checksum := getMd5(binary)
	if md5sum != checksum {
		// ダウンロードしたデータが不正
		return fmt.Errorf("checksum is invalid. err:%v file:%s calc:%s", err, md5sum, checksum)
	}

	// 正常なファイルが既に存在するなら保存する必要がない
	if validFileExists(filepath, md5sum) {
		return nil
	}

	// ファイルを保存
	fp, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		if os.IsExist(err) {
			Logger.Writef(LOG_LEVEL_INFO, "Firmware already exists. file:%s", filepath)
			return nil
		}
		return fmt.Errorf("failed to os.OpenFile: %v", err)
	}
	defer fp.Close()
	if _, err := fp.Write(binary); err != nil {
		return fmt.Errorf("failed to fp.Write: %v", err)
	}
	return nil
}

// シンボリックリンクのリンク先をnewnameで更新
func overwriteSymlink(linkpath, newname string) error {
	// 上書きができないため os.Symlink は使用しない
	// -s シンボリックリンクを作成
	// -f 既に存在する場合でも強制的に作成(上書き)
	_, err := exec.Command("ln", "-fs", newname, linkpath).Output()
	if err != nil {
		return fmt.Errorf("faled to overwriteSymlink: %v", err)
	}

	return nil
}

// getBinaryAndChecksum zipファイルのデータからバイナリとチェックサムを解凍して取得
func getBinaryAndChecksum(zipBytes []byte) ([]byte, string, error) {
	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, "", err
	}
	const (
		binaryFileName   = "etos-mainunit"
		checksumFileName = "md5sum"
	)
	var (
		binary   []byte = nil
		checksum []byte = nil
	)
	for _, f := range reader.File {
		bytes, err := getBytesFromZipFile(f)
		if err != nil {
			return nil, "", err
		}

		switch f.FileInfo().Name() {
		case binaryFileName:
			binary = bytes
		case checksumFileName:
			checksum = bytes
		default:
			fmt.Println("unrecognized file", f.FileInfo().Name(), f.FileInfo().Size(), f.FileInfo().IsDir())
			continue
		}
	}
	return binary, string(checksum), nil
}

// getBytesFromZipFile zipファイル内のファイル単位でバイト列を取得
func getBytesFromZipFile(f *zip.File) ([]byte, error) {
	reader, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// getMd5 バイト列から16進数表示のMD5チェックサムを取得
func getMd5(bytes []byte) string {
	hash := md5.New()
	defer hash.Reset()
	hash.Write(bytes)
	return hex.EncodeToString(hash.Sum(nil))
}
