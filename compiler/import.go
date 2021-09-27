package compiler

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/looplanguage/loop/lexer"
	"github.com/looplanguage/loop/models/ast"
	"github.com/looplanguage/loop/models/object"
	"github.com/looplanguage/loop/parser"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type RemoveStartsWith struct {
	StartsWith string
}

type GithubResponse struct {
	Url string `json:"zipball_url"`
	Tag string `json:"tag_name"`
}

func (r *RemoveStartsWith) Check(path string, packageName string) {
	files, err := ioutil.ReadDir(path)

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), r.StartsWith) {
			s := strings.Split(file.Name(), "-")
			version := s[len(s)-1]

			fmt.Printf("Removing package %q version %s\n", packageName, version)
			os.RemoveAll(path + "/" + file.Name())
		}
	}
}

func (c *Compiler) importString(root, data, identifier, path string) error {
	l := lexer.Create(data)
	pars := parser.Create(l)
	program := pars.Parse()

	err := c.Compile(program, path, identifier, root)

	if err != nil {
		return err
	}

	return nil
}

func (c *Compiler) importPackageLocal(root string, node *ast.Import) error {
	p := filepath.Join(filepath.Dir(root), node.File)
	content, err := ioutil.ReadFile(p)

	if err != nil {
		return fmt.Errorf("unable to import. file=%q. error=%q", node.File, err)
	}

	return c.importString(root, string(content), node.Identifier, p)
}

func (c *Compiler) importPackageGithub(root string, node *ast.Import) error {
	err := os.MkdirAll("./packages", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll("./packages/cache", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	httpClient := http.Client{
		Timeout: time.Second * 2,
	}

	r := regexp.MustCompile("https://github\\.com/(?P<repo>\\w*/\\w*)?.*")
	names := r.SubexpNames()

	urlParts := strings.Split(node.File, "@")
	url := urlParts[0]
	version := "latest"

	if len(urlParts) == 2 && urlParts[1] != "latest" {
		version = "tags/" + urlParts[1]
	}

	result := r.FindAllStringSubmatch(url, -1)
	m := map[string]string{}
	for i, n := range result[0] {
		m[names[i]] = n
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/repos/%s/releases/%s", m["repo"], version), nil)

	if err != nil {
		log.Fatal(err)
	}

	res, getErr := httpClient.Do(req)

	if getErr != nil {
		log.Fatal(getErr)
	}

	// TODO: Change fatal errors to eval errors
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	response := GithubResponse{}
	jsonErr := json.Unmarshal(body, &response)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	// Check before re-downloading
	targetDirectory := fmt.Sprintf("./packages/%v", strings.Replace(m["repo"], "/", "-", 1)+"-"+response.Tag)
	filename := strings.Split(m["repo"], "/")[1]

	input, err := ioutil.ReadFile(targetDirectory + "/" + filename + ".lp")

	if err == nil {
		return c.importString(root, string(input), node.Identifier, targetDirectory+"/"+filename+".lp")
	}

	// If old version exists, remove (TODO: Allow specifying specific version)
	check := &RemoveStartsWith{StartsWith: strings.Replace(m["repo"], "/", "-", 1)}
	check.Check("./packages", m["repo"])

	fmt.Println(fmt.Sprintf("Downloading package %q version %s", m["repo"], response.Tag))
	resp := downloadFileTo(response.Url, "./packages/cache/temp.zip")

	if resp.Type() != object.NONE {
		return fmt.Errorf(resp.Inspect())
	}

	zip, _ := Unzip("./packages/cache/temp.zip", "./packages/cache")
	os.Remove("./packages/cache/temp.zip")
	os.Rename(zip[0], targetDirectory)

	input, err = ioutil.ReadFile(targetDirectory + "/" + filename + ".lp")

	return c.importString(root, string(input), node.Identifier, targetDirectory+"/"+filename+".lp")

	return nil
}

func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

func downloadFileTo(url, location string) object.Object {
	specUrl := url
	resp, err := http.Get(specUrl)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("unable to download file, invalid repository. got=%q", err.Error())}
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return &object.Error{Message: fmt.Sprintf("unable to download file, invalid response code. got=%q", resp.StatusCode)}
	}

	// Create the file
	out, err := os.Create(location)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("unable to save file. got=%q", err.Error())}
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("unable to write to file. got=%q", err.Error())}
	}

	return &object.Null{}
}

func (c *Compiler) importPackage(root string, node *ast.Import) error {
	switch {
	case strings.HasPrefix(node.File, "https://github.com"):
		return c.importPackageGithub(root, node)
	default:
		return c.importPackageLocal(root, node)
	}
}
