package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	delimitLeft  = "${"
	delimitRight = "}"
)

type GoAutomation struct {
	AppName          string `json:"appName"`
	IsParentFolder   bool
	ParentFolderName string
	DefaultPort      int `json:"defaultPort"`
	NeedGit          string
	ProjectPath      string
	Path             string
}

func main() {

	goAutomateConfig := GoAutomation{}

	// flag.IntVar(&conf.Configuration.Port, "port", conf.Configuration.Port, "Port as used to run aht app in specific")
	flag.StringVar(&goAutomateConfig.AppName, "name", goAutomateConfig.AppName, "Application/Project name")
	flag.StringVar(&goAutomateConfig.ParentFolderName, "parentFolder", goAutomateConfig.ParentFolderName, "Application/Project parent folder name")
	flag.IntVar(&goAutomateConfig.DefaultPort, "port", goAutomateConfig.DefaultPort, "Application/Project default port")
	flag.StringVar(&goAutomateConfig.NeedGit, "git", goAutomateConfig.NeedGit, "Application/Project git")
	flag.Parse()

	if goAutomateConfig.AppName == "" {
		log.Println("App name is not mention \n Kindly mention your appname like '--name='")
		return
	}
	// goAutomateConfig.AppName = getValueFromTerminal("Enter Application Name : ")

	gopathSrc := os.Getenv("GOPATH") + "/src/"
	folderPath := gopathSrc + strings.ToLower(goAutomateConfig.AppName)
	log.Println(folderPath)
	_, err := ioutil.ReadDir(folderPath)
	if err == nil {
		fmt.Printf("%s application/project already exist in go path \n", strings.ToLower(goAutomateConfig.AppName))
		return
	}
	// isParentFolder := getValueFromTerminal("Do you need parent folder (if not press enter to return)? ")
	if goAutomateConfig.ParentFolderName != "" {
		goAutomateConfig.IsParentFolder = true
		folderPath = gopathSrc + strings.ToLower(goAutomateConfig.ParentFolderName)
		_, err = ioutil.ReadDir(folderPath)
		if err == nil {
			fmt.Printf("%s application/project already exist in go path \n", strings.ToLower(goAutomateConfig.ParentFolderName))
			return
		}
	}

	// defaultPort := getValueFromTerminal("Enter default port if any (otherwise press enter to return) : ")
	// if defaultPort != "" {
	// 	goAutomateConfig.DefaultPort, _ = strconv.Atoi(defaultPort)
	// }

	if goAutomateConfig.DefaultPort == 0 {
		goAutomateConfig.DefaultPort = 8085
	}

	// goAutomateConfig.NeedGit = getValueFromTerminal("Do you need git ?")

	projectPath := createProject(&goAutomateConfig, true, "")
	goAutomateConfig.ProjectPath = projectPath

	goAutomateConfig.Path = strings.ToLower(goAutomateConfig.AppName)
	if goAutomateConfig.IsParentFolder {
		goAutomateConfig.Path = strings.ToLower(goAutomateConfig.ParentFolderName) + "/" + strings.ToLower(goAutomateConfig.AppName)
	}

	log.Println(goAutomateConfig)
	gitInit(&goAutomateConfig, projectPath)

	createServerFiles(&goAutomateConfig)
	createGitIgnore(&goAutomateConfig)

}

func getValueFromTerminal(message string) string {

	inputReader := bufio.NewReader(os.Stdin)
	if len(message) > 0 {
		log.Println(message)
	}
	text, _ := inputReader.ReadString('\n')
	text = text[:len(text)-1]

	return text
}

func createProject(config *GoAutomation, isForParent bool, parentFolderName string) string {

	gopathSrc := os.Getenv("GOPATH") + "/src/"
	folderPath := gopathSrc + parentFolderName + strings.ToLower(config.AppName)

	if isForParent && config.ParentFolderName != "" {
		folderPath = gopathSrc + strings.ToLower(config.ParentFolderName)
		os.Mkdir(folderPath, os.ModePerm)
		return createProject(config, false, (strings.ToLower(config.ParentFolderName) + "/"))

	}

	os.Mkdir(folderPath, os.ModePerm)
	os.Mkdir(folderPath+"/server", os.ModePerm)
	os.Mkdir(folderPath+"/client", os.ModePerm)
	return folderPath
}

func gitInit(goAutomateConfig *GoAutomation, folderPath string) {
	if goAutomateConfig.NeedGit != "" && strings.Contains(goAutomateConfig.NeedGit, "y") {

		log.Println("Git initiating ...!")
		gitInitCmd := exec.Command("git", []string{"init", folderPath}...)

		_, gitInitErr := gitInitCmd.CombinedOutput()
		if gitInitErr != nil {
			log.Println("***** Git init has error ****")
			log.Fatal(gitInitErr)
		}
	}

}

func createGitIgnore(configIns *GoAutomation) {

	file, err := os.Create(configIns.ProjectPath + "/" + ".gitignore")
	if err != nil {
		log.Println("Error in file creation at ", configIns.ProjectPath, " of file .gitignore")
		log.Fatal(err)
		return
	}

	temp, err := template.New(".gitignore").Parse(getGitIgnore())
	if err != nil {
		log.Fatalf("Error in open file: %s", err.Error())
	}

	err = temp.Execute(file, &configIns)
	if err != nil {
		log.Fatalf("Error in parsing .gitignore file: %s", err.Error())
	}
}

func goToProjectLoc(folderpath string) error {
	goToProject := exec.Command("cd", []string{folderpath}...)

	out, err := goToProject.CombinedOutput()
	if err != nil {
		log.Println("***** Git init has error ****")
		log.Fatal(err)
		log.Fatal(out)
		return err
	}

	return nil
}

func createMainFile(configIns *GoAutomation) {
	appName := strings.Title(configIns.AppName)
	file, err := os.Create(configIns.ProjectPath + "/" + appName + ".go")
	if err != nil {
		log.Println("Error in file creation at ", configIns.ProjectPath, " of file : ", appName)
		log.Fatal(err)
		return
	}

	temp, err := template.New(appName + ".go").Parse(getMainTxt())
	if err != nil {
		log.Fatalf("Error in open file: %s", err.Error())
	}

	err = temp.Execute(file, &configIns)
	if err != nil {
		log.Fatalf("Error in parsing go/html file: %s", err.Error())
	}
}

func createServerFiles(configIns *GoAutomation) {

	createMainFile(configIns)

	readAndCreateFile(configIns, "/client/", "index", ".html", getIndexHtmlStr())

	os.Mkdir(configIns.ProjectPath+"/server/conf", os.ModePerm)
	readAndCreateFile(configIns, "/server/conf/", "Conf", ".go", getConfText())

	os.Mkdir(configIns.ProjectPath+"/server/controller", os.ModePerm)

	os.Mkdir(configIns.ProjectPath+"/server/dbConnection", os.ModePerm)
	readAndCreateFile(configIns, "/server/dbConnection/", "Collections", ".go", getCollectionTxt())
	readAndCreateFile(configIns, "/server/dbConnection/", "DbCon", ".go", getDbConTxt())

	os.Mkdir(configIns.ProjectPath+"/server/initApp", os.ModePerm)
	readAndCreateFile(configIns, "/server/initApp/", "InitApp", ".go", getinitAppTxt())

	os.Mkdir(configIns.ProjectPath+"/server/middleware", os.ModePerm)
	readAndCreateFile(configIns, "/server/middleware/", "Middleware", ".go", getMiddlewareTxt())

	os.Mkdir(configIns.ProjectPath+"/server/models", os.ModePerm)

	os.Mkdir(configIns.ProjectPath+"/server/routes", os.ModePerm)
	readAndCreateFile(configIns, "/server/routes/", "Router", ".go", getRouterTxt())

	os.Mkdir(configIns.ProjectPath+"/server/services", os.ModePerm)

	os.Mkdir(configIns.ProjectPath+"/server/utils", os.ModePerm)

	os.Mkdir(configIns.ProjectPath+"/server/views", os.ModePerm)
}

func readAndCreateFile(configIns *GoAutomation, folderPath string, fileName string,
	extension string, str string) {

	fileNameFolder := folderPath + fileName + extension
	file, err := os.Create(configIns.ProjectPath + fileNameFolder)
	if err != nil {
		log.Println("Error in file creation at ", folderPath, " of file : ", fileName)
		log.Fatal(err)
		return
	}

	defer file.Close()

	temp, err := template.New(fileName + extension).Parse(str)
	if err != nil {
		log.Fatalf("Error in open file: %s", err.Error())
	}

	err = temp.Execute(file, &configIns)
	if err != nil {
		log.Fatalf("Error in parsing go/html file: %s", err.Error())
	}
}

func getConfText() string {

	// rb, rberr := ioutil.ReadFile("server/conf/Conf.go")
	// if rberr != nil {
	// 	log.Fatalf("Error in reading file : %v", rberr)
	// }

	// log.Println("File = ", string(rb))

	return `package conf

		const (
			AppName = "{{.AppName}}"
		)

		type Conf struct {
			Port       int
			ServerName string
		}

		var Configuration = Conf{
			Port:       {{.DefaultPort}},
			ServerName: "",
		}
		`
}

func getCollectionTxt() string {
	return `package dbConnection

		const ()
	`
}

func getDbConTxt() string {
	return `package dbConnection

			var dbName = "{{.AppName}}"

			func MakeDBConnection() {
				connectDB()
			}

			func CloseDbConnection() {

			}

			func connectDB() {

			}

			func CopyDB() {

			}
			`
}

func getinitAppTxt() string {
	return `package initApp

			import "log"

			func AppInit() {

				log.Println("App Init ...!")
			}
			`
}

func getMiddlewareTxt() string {
	return `package middleware

			import (
				"context"
				"net/http"

				"github.com/julienschmidt/httprouter"
			)

			func WrapHandler(next http.Handler) httprouter.Handle {
				return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
					ctx := context.WithValue(r.Context(), "params", ps)
					next.ServeHTTP(w, r.WithContext(ctx))
				}
			}`
}

func getRouterTxt() string {
	return `package routes

			import (
				"{{.Path}}/server/conf"
				"{{.Path}}/server/middleware"
				"fmt"
				"net/http"

				"github.com/julienschmidt/httprouter"
				"github.com/justinas/alice"
			)

			func RouterConfig() (router *httprouter.Router) {

				indexHandlers := alice.New()

				router = httprouter.New()

				/// Serve Public ///
				staticFilePath := fmt.Sprintf("client/%s", conf.Configuration.ServerName)
				router.ServeFiles("/client/*filepath", http.Dir("client"))
				router.GET("/", middleware.WrapHandler(indexHandlers.Then(http.FileServer(http.Dir(staticFilePath)))))

				return
			}
			`
}

func getMainTxt() string {
	return `package main

			import (
				"{{.Path}}/server/conf"
				"{{.Path}}/server/dbConnection"
				"{{.Path}}/server/initApp"
				"{{.Path}}/server/routes"
				"flag"
				"fmt"
				"log"
				"net/http"
				"time"

				"github.com/rs/cors"
			)

			func main() {
				// Set flag along with std flags to print file name in log
				log.SetFlags(log.LstdFlags | log.Lshortfile)

				flag.IntVar(&conf.Configuration.Port, "port", conf.Configuration.Port, "Port as used to run aht app in specific")
				flag.Parse()

				c := cors.New(cors.Options{
					AllowedOrigins: []string{"*"},
					AllowedMethods: []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
					AllowedHeaders: []string{"Origin", "X-Requested-With", "Content-Type"},
				})

				router := routes.RouterConfig()

				server := http.Server{
					Addr:         fmt.Sprintf(":%d", conf.Configuration.Port),
					ReadTimeout:  90 * time.Second,
					WriteTimeout: 90 * time.Second,
					Handler:      c.Handler(router),
				}

				dbConnection.MakeDBConnection()

				initApp.AppInit()

				defer func() {
					dbConnection.CloseDbConnection()
				}()

				log.Printf("Listening on: %d", conf.Configuration.Port)
				err := server.ListenAndServe()
				if err != nil {
					log.Fatalf("Error in listening: %s", err.Error())
				}

			}
			`
}

func getGitIgnore() string {
	return `# Specifies intentionally untracked files to ignore when using Git
			# http://git-scm.com/docs/gitignore

			*~
			*.sw[mnpcod]
			*.log
			*.tmp
			*.tmp.*
			log.txt
			*.sublime-project
			*.sublime-workspace
			.vscode/
			npm-debug.log*

			*.md
			.idea/
			.sass-cache/
			.tmp/
			.versions/
			coverage/
			dist/
			node_modules/
			tmp/
			temp/
			hooks/
			platforms/
			plugins/
			plugins/android.json
			plugins/ios.json
			www/
			$RECYCLE.BIN/

			.DS_Store
			Thumbs.db
			UserInterfaceState.xcuserstate`
}

func getIndexHtmlStr() string {
	return `<!DOCTYPE html>
			<html>

			<head lang="en">
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">

				<title>{{.AppName}}</title>


			</head>

			<body>

				<h1 style="padding:20%;">
					Go Automation
				</h1>
			</body>

			</html>`
}

// func serverInit() {
// 	// Set flag along with std flags to print file name in log
// 	log.SetFlags(log.LstdFlags | log.Lshortfile)

// 	log.Println("Go Automation for project set up")

// 	flag.IntVar(&conf.Configuration.Port, "port", conf.Configuration.Port, "Port as used to run aht app in specific")
// 	flag.Parse()

// 	c := cors.New(cors.Options{
// 		AllowedOrigins: []string{"*"},
// 		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
// 		AllowedHeaders: []string{"Origin", "X-Requested-With", "Content-Type"},
// 	})

// 	router := routes.RouterConfig()

// 	server := http.Server{
// 		Addr:         fmt.Sprintf(":%d", conf.Configuration.Port),
// 		ReadTimeout:  90 * time.Second,
// 		WriteTimeout: 90 * time.Second,
// 		Handler:      c.Handler(router),
// 	}

// 	dbConnection.MakeDBConnection()

// 	initApp.AppInit()

// 	defer func() {
// 		dbConnection.CloseDbConnection()
// 	}()

// 	log.Printf("Listening on: %d", conf.Configuration.Port)
// 	err := server.ListenAndServe()
// 	if err != nil {
// 		log.Fatalf("Error in listening: %s", err.Error())
// 	}

// }
