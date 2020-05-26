package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

var reVersion = regexp.MustCompile("const versionString = \"(.*?)\"")

type archiveType string

const (
	archiveTypeZip   archiveType = ".zip"
	archiveTypeTarGz archiveType = ".tar.gz"
)

type target struct {
	os          string
	arch        string
	suffix      string
	archiveType archiveType
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// test that go works
	check(exec.Command("go", "version").Run())

	// assume cwd is main roamer package
	outputDir := filepath.Join("build", "output")
	_, err := os.Stat(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			check(os.Mkdir(outputDir, 0777))
		} else {
			panic(err)
		}
	}

	// clean the output directory
	outputFiles, err := ioutil.ReadDir(outputDir)
	check(err)
	for _, outputFile := range outputFiles {
		check(os.Remove(filepath.Join(outputDir, outputFile.Name())))
	}

	// parse version
	versionFile, err := ioutil.ReadFile("version.go")
	check(err)

	version := string(reVersion.FindSubmatch(versionFile)[1])

	log.Printf("Building roamer version %s", version)

	targets := []target{
		target{"windows", "386", ".exe", archiveTypeZip},
		target{"windows", "amd64", ".exe", archiveTypeZip},

		target{"linux", "386", "", archiveTypeTarGz},
		target{"linux", "amd64", "", archiveTypeTarGz},

		target{"darwin", "amd64", "", archiveTypeTarGz},
	}

	for _, target := range targets {
		log.Printf("Building for %s/%s...", target.os, target.arch)

		buildDir, err := ioutil.TempDir(os.TempDir(), "roamer_build_")
		check(err)
		defer os.RemoveAll(buildDir)

		executableName := "roamer" + target.suffix

		buildCmd := exec.Command("go", "build", "-trimpath", "-tags", "nocgo", "-o", executableName, "github.com/thatoddmailbox/roamer/cli")
		buildCmd.Dir = buildDir
		buildCmd.Env = append(os.Environ(),
			"GOOS="+target.os,
			"GOARCH="+target.arch,
		)
		output, err := buildCmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(output))
			panic(err)
		}

		// open the executable
		executableFile, err := os.Open(filepath.Join(buildDir, executableName))
		check(err)
		defer executableFile.Close()

		executableStat, err := executableFile.Stat()
		check(err)

		// build an archive
		archiveName := "roamer_" + version + "_" + target.os + "_" + target.arch + string(target.archiveType)
		archiveFile, err := os.Create(filepath.Join(outputDir, archiveName))
		check(err)

		if target.archiveType == archiveTypeZip {
			archiveWriter := zip.NewWriter(archiveFile)

			archivedExecutable, err := archiveWriter.Create(executableName)
			check(err)

			_, err = io.Copy(archivedExecutable, executableFile)
			check(err)

			check(archiveWriter.Close())
		} else if target.archiveType == archiveTypeTarGz {
			gzipWriter := gzip.NewWriter(archiveFile)

			archiveWriter := tar.NewWriter(gzipWriter)

			archiveWriter.WriteHeader(&tar.Header{
				Name: executableName,
				Mode: 0777,
				Size: int64(executableStat.Size()),
			})

			_, err = io.Copy(archiveWriter, executableFile)
			check(err)

			check(archiveWriter.Close())

			check(gzipWriter.Close())
		}

		check(archiveFile.Close())

		log.Println("Done!")
	}
}
