package main

import (
	"flag"
	"strings"
	"os/user"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"log"
)

type PgyerResponse struct {
	BuildKey 				string
	BuildType 				string
	BuildIsFirst 			string
	BuildIsLastest			string
	BuildFileKey			string
	BuildFileName			string
	BuildFileSize			string
	BuildName				string
	BuildVersion			string
	BuildVersionNo			string
	BuildIdentifier			string
	BuildIcon				string
	BuildDescription		string
	BuildUpdateDescription	string
	BuildScreenshots		string
	BuildShortcutUrl		string
	BuildCreated			string
	BuildUpdated			string
	BuildQRCodeURL			string
}

const (
	Pgyer_BaseUrl = "https://www.pgyer.com/"
	Pgyer_APIKey = "your pgyer api key"
)

var workspaceName string = "demo.xcworkspace"
var schemeName string = "demo"
var archivePath string = "~/Desktop/Hysport"
var ipaPath string = "~/Desktop/hysport-ipa"
var exportOptionsPath string = "./AdHocExportOptions.plist"

var uploadDescription *string = flag.String("description", "", "string value used to describe the version")
var configuration *string = flag.String("configuration", "Debug", "build configuration(Debug or Release).")

func main() {
	flag.Parse()

	var upload = selectWhetherUpload()
	
	clean()
	archive()
	exportArchiveToIPA()

	if upload == false {
		return
	}

	log.Println("start upload...")
	resp, err := uploadIpaToPgyer(getIpaFilePath())
	if err != nil {
		log.Fatalf("upload to pgyer failed: %s\n", err)
	}
	log.Println("Upload finished")

	pgyerResp, err := parseUploadResult(resp)
	if err != nil {
		log.Fatal("parse upload result failed")
	}
	printResult(pgyerResp)
}

func selectExportMethod() (b byte, err error) {
	fmt.Println("Please select export method")
	fmt.Println("1. ad-hoc")
	fmt.Println("2. app-store")
	var c byte
	_, e := fmt.Scanf("%c", &c)
	if e != nil {
		return b, e
	}
	return c, err
}

func selectWhetherUpload() bool {
	fmt.Println("Do you want upload to pgyer.com, after export ipa. (y/n)")
	var c byte
	_, e := fmt.Scanf("%c", &c)
	if e != nil || (c != 'n' && c != 'N' && c != 'y' && c != 'Y') {
		fmt.Println("invalid input.")
		return selectWhetherUpload()
	}
	if (c == 'n' || c == 'N') {
		return false
	}
	if (c == 'y' || c == 'Y') {
		return true
	}
	return true
}

func printResult(pgyerResp PgyerResponse) {
	log.Printf("shortcut url: %s\n", Pgyer_BaseUrl + pgyerResp.BuildShortcutUrl)
	log.Printf("RRCode url: %s\n", pgyerResp.BuildQRCodeURL)
	printImportantInfo("UPLOAD TO PGYER SUCCESSFULLY")
}

func printImportantInfo(str string) {
	log.Printf("*****	%s	*****\n", str)
}

// clean build
func clean() {
	log.Println("Start clean...")
	cmd := exec.Command("xcodebuild", "clean", "-workspace", workspaceName, "-scheme", schemeName, "-configuration", *configuration)
	_, err := cmd.Output()
	if err != nil {
		log.Fatalf("clean failed: %s\n", err)
	}
	log.Println("clean successfully")
}

// archive
func archive() {
	log.Println("Start archive...")
	cmd := exec.Command("xcodebuild", "archive", "-workspace", workspaceName, "-scheme", schemeName, "-configuration", *configuration, "-archivePath", archivePath)
	_, err := cmd.Output()
	if err != nil {
		log.Fatal("archive failed")
	}
	log.Printf("archive complete successfully, path is: %s\n", archivePath)
}

// export xxx.xcarchive to xxx.ipa
func exportArchiveToIPA() {
	log.Println("start export archive to ipa")
	cmd := exec.Command("xcodebuild", "-exportArchive", "-archivePath", getArchivePath(), "-exportPath", ipaPath, "-exportOptionsPlist", exportOptionsPath)
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("export failed: %s\n", err)
	}
	log.Println(string(out))
}

// get xxx.archive file path
func getArchivePath() string {
	return archivePath + ".xcarchive"
}

// get xxx.ipa file path
func getIpaFilePath() string {
	if strings.HasPrefix(ipaPath, "~/") {
		user, err := user.Current()
		if nil != err {
			log.Fatalf("parse user home dir failed")
		}
		ipaPath = strings.Replace(ipaPath, "~", user.HomeDir, 1)
	}

	path := ipaPath + "/" + schemeName + ".ipa"
	log.Printf("ipa file path: %s\n", path)
	return path
}

// upload ipa file to pgyer
func uploadIpaToPgyer(filename string) (response []byte, err error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	bodyWriter.WriteField("_api_key", Pgyer_APIKey)
	bodyWriter.WriteField("buildUpdateDescription", *uploadDescription)

	fileWriter, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return response, err
	}

	// open file handle
	log.Printf("open ipa file: %s\n", filename)
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return response, err
	}
	defer file.Close()
	
	// iocopy
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return response, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(Pgyer_BaseUrl + "apiv2/app/upload", contentType, bodyBuf)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()
	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	user.Current()
	return response, nil
}

func parseUploadResult(jsonData []byte) (pgyerResponse PgyerResponse, err error) {
	// 1. json->map
	var dict map[string]interface{}
	if err := json.Unmarshal(jsonData, &dict); err != nil {
		return pgyerResponse, err
	}

	// 2. 把“data”对应的map解析成json
	tmp, err := json.Marshal(dict["data"])
	if err != nil {
		return pgyerResponse, err
	}

	// 3. json->struct
	json.Unmarshal(tmp, &pgyerResponse)
	return pgyerResponse, nil
}