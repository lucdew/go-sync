package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	xxhash "github.com/OneOfOne/xxhash"
	log "github.com/Sirupsen/logrus"
	docopt "github.com/docopt/docopt-go"
)

type customFormatter struct {
}

func (f *customFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(entry.Message), nil
}

var (
	destDir string
	mirror  bool
)

func doCopyFile(srcDir string, srcFi os.FileInfo, destPath string) error {
	srcFile, err := os.Open(filepath.Join(srcDir, srcFi.Name()))
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		if err = dstFile.Chmod(srcFi.Mode()); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(srcDir string, srcFi os.FileInfo, destPath string) error {

	err := doCopyFile(srcDir, srcFi, destPath)
	if err != nil {
		return err
	}
	return os.Chtimes(destPath, srcFi.ModTime(), srcFi.ModTime())
}

func hashFile(srcFilePath string) (uint64, error) {

	h := xxhash.New64()
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	if _, err = io.Copy(h, srcFile); err != nil {
		return 0, err
	}
	return h.Sum64(), nil

}

func syncFolder(folderAbsPath string, folderRelativePath string, destCreated bool) error {

	log.Infof("Syncing folder %s\n", folderAbsPath)
	fileInfos, err := ioutil.ReadDir(folderAbsPath)
	if err != nil {
		return err
	}

	destAbsPath := filepath.Join(destDir, folderRelativePath)
	destFilesMap := make(map[string]os.FileInfo, len(fileInfos))

	if !destCreated {
		_, err := os.Stat(destAbsPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		} else if err == nil {
			destFileInfos, err := ioutil.ReadDir(destAbsPath)
			if err != nil {
				return err
			}
			for _, fi := range destFileInfos {
				destFilesMap[filepath.Base(fi.Name())] = fi
			}
		} else {
			srcStats, err := os.Stat(folderAbsPath)
			if err != nil {
				return err
			}
			if err = os.Mkdir(destAbsPath, srcStats.Mode()); err != nil {
				return err
			}
			if err = os.Chtimes(destAbsPath, srcStats.ModTime(), srcStats.ModTime()); err != nil {
				return err
			}
		}
	}

	for _, srcFi := range fileInfos {
		log.Debugf("Syncing %s\n", srcFi.Name())
		srcName := srcFi.Name()
		destFi := destFilesMap[srcName]
		destPath := filepath.Join(destAbsPath, srcName)
		srcRelPath := filepath.Join(folderRelativePath, srcName)
		srcAbsPath := filepath.Join(folderAbsPath, srcName)

		if nil == destFi {

			log.Debugf("%s does not exist\n", destPath)

			if srcFi.IsDir() {

				if err = os.Mkdir(destPath, srcFi.Mode()); err != nil {
					return err
				}
				if err = os.Chtimes(destPath, srcFi.ModTime(), srcFi.ModTime()); err != nil {
					return err
				}

				syncFolder(srcAbsPath, srcRelPath, true)
			} else {
				if err = copyFile(folderAbsPath, srcFi, destPath); err != nil {
					return err
				}
				log.Debugf("%s copied\n", destPath)
			}
		} else {

			if srcFi.IsDir() {
				syncFolder(srcAbsPath, srcRelPath, false)
			} else {
				if srcFi.Size() != destFi.Size() {
					if err = copyFile(folderAbsPath, srcFi, destPath); err != nil {
						return err
					}
					log.Debugf("%s copied, source file size is different\n", destPath)
				} else {

					srcHash, err := hashFile(srcAbsPath)
					if err != nil {
						return err
					}
					dstHash, err := hashFile(destPath)
					if err != nil {
						return err
					}
					if srcHash != dstHash {
						copyFile(srcAbsPath, srcFi, destPath)
						log.Debugf("%s copied, hash do not match\n", destPath)
					}

				}
			}

			delete(destFilesMap, srcName)

		}

	}

	if mirror {
		for _, v := range destFilesMap {
			if err = os.RemoveAll(path.Join(destAbsPath, v.Name())); err != nil {
				return err
			}
			log.Debugf("Deleted %s", path.Join(destAbsPath, v.Name()))
		}
	}

	return nil

}

func main() {

	log.SetFormatter(new(customFormatter))
	usage := `gsync.

Usage:
  gsync -s source_folder... -d destination_folder [-m | --mirror] [-v | --verbose]
  gsync -h | --help
  gsync --version

Options:
  -h --help  Show this screen.
  --version  Show version.
  -s source_folder  Source folder.
  -d destination_folder  Destination folder.
	-m --mirror  Mirror.
	-v --verbose  verbose mode`

	arguments, _ := docopt.Parse(usage, nil, true, "gsync 1.0", false)

	if arguments["--verbose"].(bool) {
		log.SetLevel(log.DebugLevel)
	}

	sourceDirs := arguments["-s"].([]string)
	destDir = arguments["-d"].(string)
	mirror = arguments["--mirror"].(bool)

	sourceDirsInfo := make([]os.FileInfo, len(sourceDirs))
	for idx, sDir := range sourceDirs {
		sfi, err := os.Stat(sDir)
		if err != nil {
			log.Fatalf("source directory %s does not exist, cannot continue", sDir)
		}
		sourceDirsInfo[idx] = sfi
	}

	_, err := os.Stat(destDir)
	if err != nil && os.IsNotExist(err) {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Printf("Destination directory %s does not exit, shall it be created (Y/n)?\n ", destDir)
		scanner.Scan()
		text := scanner.Text()
		if text == "" || strings.EqualFold(text, "y") {
			if err = os.MkdirAll(destDir, 0755); err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("Exiting")
			os.Exit(0)
		}
	} else if err != nil {
		log.Fatal(err)
	}

	for _, sDir := range sourceDirs {
		if err = syncFolder(sDir, filepath.Base(sDir), false); err != nil {
			log.Fatal(err)
		}

	}

}
