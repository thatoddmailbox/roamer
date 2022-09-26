package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
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

func extractSingleFileTarGz(tarGzPath string, targetPath string) {
	f, err := os.Open(tarGzPath)
	check(err)
	defer f.Close()

	gr, err := gzip.NewReader(f)
	check(err)
	defer gr.Close()

	tarFile, err := tar.NewReader(gr).Next()
	check(err)

	targetFile, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE, 0777)
	check(err)
	defer targetFile.Close()

	_, err = io.CopyN(targetFile, gr, tarFile.Size)
	check(err)
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

	// check that we have makefat
	makefatPath := filepath.Join("build", "makefat")
	_, err = os.Stat(makefatPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalln("Missing makefat binary. Build https://github.com/randall77/makefat and place it in the build/ directory.")
		} else {
			panic(err)
		}
	}

	// clean the output directory
	outputFiles, err := os.ReadDir(outputDir)
	check(err)
	for _, outputFile := range outputFiles {
		check(os.Remove(filepath.Join(outputDir, outputFile.Name())))
	}

	// parse version
	versionFile, err := os.ReadFile("version.go")
	check(err)

	version := string(reVersion.FindSubmatch(versionFile)[1])

	log.Printf("Building roamer version %s", version)

	targets := []target{
		{"windows", "386", ".exe", archiveTypeZip},
		{"windows", "amd64", ".exe", archiveTypeZip},

		{"linux", "386", "", archiveTypeTarGz},
		{"linux", "amd64", "", archiveTypeTarGz},

		{"darwin", "amd64", "", archiveTypeTarGz},
		{"darwin", "arm64", "", archiveTypeTarGz},
		{"darwin", "universal", "", archiveTypeTarGz},
	}

	for _, target := range targets {
		log.Printf("Building for %s/%s...", target.os, target.arch)

		buildDir, err := os.MkdirTemp(os.TempDir(), "roamer_build_")
		check(err)
		defer os.RemoveAll(buildDir)

		executableName := "roamer" + target.suffix
		executablePath := filepath.Join(buildDir, executableName)

		isDarwinUniversal := target.os == "darwin" && target.arch == "universal"
		if !isDarwinUniversal {
			buildCmd := exec.Command("go", "build", "-trimpath", "-buildvcs=true", "-tags", "nocgo", "-o", executablePath, "github.com/thatoddmailbox/roamer/cli")
			buildCmd.Dir, err = os.Getwd()
			check(err)
			buildCmd.Env = append(os.Environ(),
				"GOOS="+target.os,
				"GOARCH="+target.arch,
			)
			output, err := buildCmd.CombinedOutput()
			if err != nil {
				fmt.Println(string(output))
				panic(err)
			}
		} else {
			// we already made darwin/amd64 and darwin/arm64, time to combine them

			amd64Archive := filepath.Join(outputDir, "roamer_"+version+"_darwin_amd64.tar.gz")
			amd64Binary := filepath.Join(buildDir, "roamer-amd64")
			extractSingleFileTarGz(amd64Archive, amd64Binary)

			arm64Archive := filepath.Join(outputDir, "roamer_"+version+"_darwin_arm64.tar.gz")
			arm64Binary := filepath.Join(buildDir, "roamer-arm64")
			extractSingleFileTarGz(arm64Archive, arm64Binary)

			buildCmd := exec.Command(makefatPath, executablePath, amd64Binary, arm64Binary)
			buildCmd.Dir, err = os.Getwd()
			check(err)
			output, err := buildCmd.CombinedOutput()
			if err != nil {
				fmt.Println(string(output))
				panic(err)
			}
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

			err = archiveWriter.WriteHeader(&tar.Header{
				Name: executableName,
				Mode: 0777,
				Size: int64(executableStat.Size()),
			})
			check(err)

			_, err = io.Copy(archiveWriter, executableFile)
			check(err)

			check(archiveWriter.Close())

			check(gzipWriter.Close())
		}

		check(archiveFile.Close())

		log.Println("Done!")
	}
}
