package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var ModuleListFilename = "packages.txt"
var NoticesPath = "NOTICES.txt"
var NoticesTemplatePath = "assets/NOTICES.tmpl"

type Module struct {
	Name           string
	LicenseType    string
	LicenseContent string
}

type ModuleData struct {
	Modules []Module
}

func findLicenses(targetDir string, debug bool) (map[string]string, error) {
	licenseFileRegex, _ := regexp.Compile("LICENSE.*")
	licenseFileMap := map[string]string{}

	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Failure accessing a path %q: %v", path, err)
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip things that aren't license files
		if !licenseFileRegex.MatchString(info.Name()) {
			return nil
		}

		if debug {
			log.Printf("Found: %q", path)
		}

		// Turn full path into a module name
		moduleName := strings.TrimPrefix(path, targetDir+"/")
		moduleName = strings.TrimSuffix(moduleName, "/"+info.Name())

		licenseContent, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		licenseFileMap[moduleName] = string(licenseContent)

		return nil
	})

	return licenseFileMap, err
}

func getModules(moduleListPath string, debug bool) ([]Module, error) {
	moduleRawData, err := ioutil.ReadFile(moduleListPath)
	if err != nil {
		return nil, err
	}

	moduleData := string(moduleRawData)
	moduleInfoLines := strings.Split(moduleData, "\n")

	modules := []Module{}
	for _, moduleInfoLine := range moduleInfoLines {
		if debug {
			log.Println(moduleInfoLine)
		}

		splitModulesInfoLine := strings.Split(moduleInfoLine, ",")

		if len(splitModulesInfoLine) < 3 {
			log.Printf("WARN: Found unparseable line: '%s'", moduleInfoLine)
			continue
		}
		moduleName, licenseType := splitModulesInfoLine[0], splitModulesInfoLine[2]

		module := Module{
			LicenseType: licenseType,
			Name:        moduleName,
		}

		modules = append(modules, module)
	}

	// Sort the output
	sort.Slice(modules[:], func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})

	return modules, nil
}

func main() {
	targetDir := os.Args[1]
	log.Printf("Combining licenses in '%s'...", targetDir)

	// Get a list of modules and their generic license names
	moduleListPath := path.Join(targetDir, ModuleListFilename)
	modules, err := getModules(moduleListPath, false)
	if err != nil {
		log.Printf(
			"Error reading package list file '%s': %v",
			moduleListPath,
			err,
		)
		os.Exit(1)
	}

	// Initialize our template data object
	moduleData := ModuleData{
		Modules: modules,
	}

	// Collect all license files
	licenseFiles, err := findLicenses(targetDir, false)
	if err != nil {
		log.Printf("Error walking the path %q: %v", targetDir, err)
		os.Exit(1)
	}
	log.Printf("Found %d license files", len(licenseFiles))

	// Append license texts to our module objects
	for index, module := range moduleData.Modules {
		if _, ok := licenseFiles[module.Name]; !ok {
			log.Printf("WARN! Could not find license for module '%s'!", module.Name)
		}

		moduleData.Modules[index].LicenseContent = licenseFiles[module.Name]
	}

	// Open the NOTICES file
	log.Printf("Opening '%s'...", NoticesPath)
	noticesFile, err := os.Create(NoticesPath)
	if err != nil {
		log.Printf("Error creating %s: %v", NoticesPath, err)
		os.Exit(1)
	}
	defer noticesFile.Close()

	// Generate and write the license data to it
	log.Printf("Generating '%s' file from template '%s'...", NoticesPath, NoticesTemplatePath)
	tmpl := template.Must(template.ParseFiles(NoticesTemplatePath))
	err = tmpl.Execute(noticesFile, moduleData)
	if err != nil {
		log.Printf("Error running template '%s': %v", NoticesTemplatePath, err)
		os.Exit(1)
	}

	log.Println("Done!")
}
