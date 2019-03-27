package main

import (
	"bytes"
	"github.com/andlabs/ui"
	"io/ioutil"
	"os"
	"strings"
)
import _ "github.com/andlabs/ui/winmanifest"

var mainwin *ui.Window

func convNewline(str, nlcode string) string {
	return strings.NewReplacer(
		"\r\n", nlcode,
		"\r", nlcode,
		"\n", nlcode,
	).Replace(str)
}

func reportError(message string) {
	ui.MsgBox(mainwin, "エラー", message)
}

func exists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil { return true }
	if os.IsNotExist(err) { return false }
	return true
}

func findApoDirectories() ([]string) {
	result := []string{}
	paths := []string{
		"C:\\Program Files\\EqualizerAPO",
		"C:\\Program Files (x86)\\EqualizerAPO",
	}
	for _, v := range paths {
		if exists(v) {
			result = append(result, v)
		}
	}
	return result
}

func translationDirPath(rootDir string) (string) {
	return rootDir + "/translations"
}

func qtbaseQmPath(rootDir string, lang string) (string) {
	return rootDir + "/translations/qtbase_" + lang + ".qm"
}

func editorQmPath(rootDir string, lang string) (string) {
	return rootDir + "/translations/Editor_" + lang + ".qm"
}

func backupPath(rootDir string) (string) {
	return rootDir + "/Editor_backup.exe"
}

func exePath(rootDir string) (string) {
	return rootDir + "/Editor.exe"
}

func applyTranslation(rootDir string, qtbaseDeQm []byte, qtbaseQm []byte, editorDeQm []byte, editorQm []byte) {
	original, originalReadErr := ioutil.ReadFile(exePath(rootDir))
	if originalReadErr != nil {
		reportError(exePath(rootDir) + "の読み込みが失敗しました")
		return
	}

	backupWriteErr := ioutil.WriteFile(backupPath(rootDir), original, os.ModePerm)
	if backupWriteErr != nil {
		reportError("バックアップの作成が失敗しました")
		return
	}

	mkdirErr := os.Mkdir(translationDirPath(rootDir), os.ModePerm)
	if mkdirErr != nil {
		reportError("翻訳ファイル用ディレクトリの作成が失敗しました")
		return
	}

	errQtbaseQm := ioutil.WriteFile(qtbaseQmPath(rootDir, "en"), qtbaseQm, os.ModePerm)
	if errQtbaseQm != nil {
		reportError(qtbaseQmPath(rootDir, "en") + "の書き込みが失敗しました")
		return
	}

	errQtbaseDeQm := ioutil.WriteFile(qtbaseQmPath(rootDir, "de"), qtbaseDeQm, os.ModePerm)
	if errQtbaseDeQm != nil {
		reportError(qtbaseQmPath(rootDir, "de") + "の書き込みが失敗しました")
		return
	}

	errEditorQm := ioutil.WriteFile(editorQmPath(rootDir, "en"), editorQm, os.ModePerm)
	if errEditorQm != nil {
		reportError(editorQmPath(rootDir, "en") + "の書き込みが失敗しました")
		return
	}

	errEditorDeQm := ioutil.WriteFile(editorQmPath(rootDir, "de"), editorDeQm, os.ModePerm)
	if errEditorDeQm != nil {
		reportError(editorQmPath(rootDir, "de") + "の書き込みが失敗しました")
		return
	}

	replaced := bytes.Replace(original, []byte(":/translations/qtbase"), []byte("translations/qtbase\000\000"),1)
	replaced = bytes.Replace(original, []byte(":/translations/Editor"), []byte("translations/Editor\000\000"),1)

	if len(replaced) != len(original) {
		reportError(exePath(rootDir) + "の書き換えが失敗しました (置換後の長さが異なる)")
		return
	}
	if !bytes.Contains(replaced, []byte("translations/qtbase\000\000")) {
		reportError(exePath(rootDir) + "の書き換えが失敗しました (:/translations/qtbaseの置換失敗)")
		return
	}
	if !bytes.Contains(replaced, []byte("translations/Editor\000\000")) {
		reportError(exePath(rootDir) + "の書き換えが失敗しました (:/translations/Editorの置換失敗)")
		return
	}

	replacedErr := ioutil.WriteFile(exePath(rootDir), replaced, os.ModePerm)
	if replacedErr != nil {
		reportError(exePath(rootDir) + "の書き込みが失敗しました")
		return
	}

	ui.MsgBox(mainwin, "成功", rootDir + "に日本語化を適用しました")
}

func resetTranslation(rootDir string, showSuccess bool) {
	os.RemoveAll(translationDirPath(rootDir))
	if exists(backupPath(rootDir)) {
		err := os.Rename(backupPath(rootDir), exePath(rootDir))
		if err != nil {
			reportError("バックアップの復元に失敗しました")
			return
		}
	}
	if showSuccess {
		ui.MsgBox(mainwin, "成功", rootDir + "を元に戻しました")
	}
}

func setupUI() {
	mainwin = ui.NewWindow("Equalizer APO Translator", 480, 240, true)
	mainwin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})

	vbox := ui.NewVerticalBox()
	vbox.SetPadded(true)

	applyJapaneseButton := ui.NewButton("日本語化を適用する")
	applyJapaneseButton.OnClicked(func(*ui.Button) {
		dirs := findApoDirectories()
		if len(dirs) == 0 {
			reportError("インストールされたEqualizer APOが見つかりませんでした")
			return
		}

		qtbaseDeQm, err := Asset("data/qtbase_de.qm")
		if err != nil {
			reportError("qtbase_de.qmの読み込みが失敗しました")
			return
		}

		qtbaseQm, err2 := Asset("data/qtbase_ja.qm")
		if err2 != nil {
			reportError("qtbase_ja.qmの読み込みが失敗しました")
			return
		}

		editorDeQm, err3 := Asset("data/Editor_de.qm")
		if err3 != nil {
			reportError("Editor_de.qmの読み込みが失敗しました")
			return
		}

		editorQm, err4 := Asset("data/Editor_ja.qm")
		if err4 != nil {
			reportError("Editor_ja.qmの読み込みが失敗しました")
			return
		}

		for _, v := range dirs {
			resetTranslation(v, false)
			applyTranslation(v, qtbaseDeQm, qtbaseQm, editorDeQm, editorQm)
		}
	})

	resetButton := ui.NewButton("全ての変更を元に戻す")
	resetButton.OnClicked(func(*ui.Button) {
		dirs := findApoDirectories()
		if len(dirs) == 0 {
			reportError("インストールされたEqualizer APOが見つかりませんでした")
			return
		}
		for _, v := range dirs {
			resetTranslation(v, true)
		}
	})

	vbox.Append(applyJapaneseButton, false)
	vbox.Append(resetButton, false)

	entry := ui.NewNonWrappingMultilineEntry()
	entry.SetReadOnly(true)
	readme, err := Asset("data/readme.txt")
	if err != nil {
		reportError("readme.txtの読み込みが失敗しました")
	}
	entry.SetText(convNewline(string(readme), "\r\n"))
	vbox.Append(entry, true)

	mainwin.SetChild(vbox)

	mainwin.Show()
}

func main() {
	ui.Main(setupUI)
}

