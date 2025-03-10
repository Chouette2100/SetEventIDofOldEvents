// Copyright © 2025 chouette2100@gmail.com
// Released under the MIT license
// https://opensource.org/licenses/mit-license.php
package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"runtime"
	"strconv"

	"io"
	"time"

	"github.com/go-gorp/gorp"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/Chouette2100/srapi/v2"
	"github.com/Chouette2100/srdblib/v2"
)

/*
ver.1.0.0 最初のリリース
*/

// Set EventID of Old Events That Don't Have EventID
func main() {

	const version = "1.0.0"

	var err error

	// >>>>>>>>>>>>>>>>>>>>>
	// ログファイルを開く
	logfilename := version + "_" + srdblib.Version + "_" + srapi.Version + "_" + time.Now().Format("20060102") + ".txt"
	logfile, err := os.OpenFile(logfilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Println("ログファイルを開けません:", err)
		os.Exit(1)
	}
	defer logfile.Close()

	// フォアグラウンド（端末に接続されているか）を判定
	isForeground := terminal.IsTerminal(int(os.Stdout.Fd()))

	var logOutput io.Writer
	if isForeground {
		// フォアグラウンドならログファイル + コンソール
		logOutput = io.MultiWriter(os.Stdout, logfile)
	} else {
		// バックグラウンドならログファイルのみ
		logOutput = logfile
	}

	// ロガーの設定
	log.SetOutput(logOutput)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// ログ出力テスト
	log.Println("アプリケーションを起動しました")
	// <<<<<<<<<<<<<<<<<<<<

	// Print the Go version
	log.Println("You are using", runtime.Compiler, "compiler")
	log.Println("on a", runtime.GOARCH, "machine")
	log.Println("with Go version", runtime.Version())

	// >>>>>>>>>>>>>>>>>>>>>
	// Check for the existence of a lock file to avoid double activation, and if not,
	// create a lock file. The lock file is deleted after the process is completed.
	// ロックファイルのパス
	lockFilePath := "/tmp/SetEventIDofOldEvents.lock"

	// 既存のロックファイルをチェック
	if checkExistingLock(lockFilePath) {
		log.Println("既に別のインスタンスが実行中です。終了します。")
		os.Exit(1)
	}

	// ロックファイルを作成・オープン
	lockFile, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("ロックファイル作成エラー: %v\n", err)
		os.Exit(1)
	}
	// 現在のプロセスIDを書き込む
	fmt.Fprintf(lockFile, "%d", os.Getpid())
	lockFile.Close()

	// 終了時のクリーンアップ
	defer func() {
		os.Remove(lockFilePath)
		log.Println("ロックファイルを削除しました")
	}()
	// <<<<<<<<<<<<<<<<<<<<

	// >>>>>>>>>>>>>>>>>>>>>
	// データベース接続
	dbconfig, err := srdblib.OpenDb("DBConfig.yml")
	if err != nil {
		log.Printf("Database error. err = %v\n", err)
		return
	}
	if dbconfig.UseSSH {
		defer srdblib.Dialer.Close()
	}
	defer srdblib.Db.Close()

	dial := gorp.MySQLDialect{Engine: "InnoDB", Encoding: "utf8mb4"}
	srdblib.Dbmap = &gorp.DbMap{Db: srdblib.Db,
		Dialect:         dial,
		ExpandSliceArgs: true, //スライス引数展開オプションを有効化する
	}
	srdblib.Dbmap.AddTableWithName(srdblib.Wevent{}, "wevent").SetKeys(false, "Eventid")
	srdblib.Dbmap.AddTableWithName(srdblib.Weventuser{}, "weventuser").SetKeys(false, "Eventid", "Userno")
	// <<<<<<<<<<<<<<<<<<<<

	for i := 0; i < 10; i++ {
		// SRCGIから起動されてケースではまだイベントが作成されていない場合があるためウェイトする
		time.Sleep(time.Duration(i) * time.Second)

		// Select all events that don't have an eventid
		// The eventid is a 4-character string that starts with '@@@@'
		var eu []EventAndUser
		eu, err = SelectEventsThatDontHaveEventID()
		if err != nil {
			log.Printf("SelectEventsThatDontHaveEventID() failed: %v", err)
			return
		}

		// if len(eu) == 0 {
		// 	log.Println("No events that don't have an eventid")
		// 	break
		// }

		// Set the eventid of the events that don't have an eventid
		err = SetEventIDofEvents(eu)
		if err != nil {
			log.Printf("SetEventIDofEvents() failed: %v", err)
			return
		}
		log.Printf("SetEventIDofEvents() succeeded i=%d nevent=%d\n", i, len(eu))
	}

}

// 既存のロックファイルをチェックし、有効なプロセスが実行中かを確認
func checkExistingLock(lockFilePath string) bool {
	// content, err := ioutil.ReadFile(lockFilePath)
	content, err := os.ReadFile(lockFilePath)
	if err != nil {
		// ファイルが存在しない場合はロックなし
		return false
	}

	// ロックファイルからPIDを読み取る
	pid, err := strconv.Atoi(string(content))
	if err != nil {
		// PIDが読み取れない場合は古いロックファイルと判断して削除
		os.Remove(lockFilePath)
		return false
	}

	// プロセスが実行中かチェック
	// Linuxでは /proc/{pid} にアクセスする方法もあります
	process, err := os.FindProcess(pid)
	if err != nil {
		// プロセスが見つからない場合はロックファイルを削除
		os.Remove(lockFilePath)
		return false
	}

	// プロセスにシグナル0を送信してチェック（実際にはシグナルは送信されない）
	// err = process.Signal(syscall.Signal(0))
	err = process.Signal(unix.Signal(0))
	if err != nil {
		// プロセスが存在しない場合はロックファイルを削除
		os.Remove(lockFilePath)
		return false
	}

	// プロセスは実行中
	return true
}
