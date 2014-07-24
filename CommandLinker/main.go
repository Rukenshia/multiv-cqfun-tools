package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

type Command struct {
	name string
	path string
}

func main() {
	commandDofileProto := `dofile("{Path}");`
	commandProto := `		{Command}();`
	commandFileProto := `
/*
 *		MultIV CQFun
 *	@file: Commands.nut
 *	@author: Command Linker [auto generated at ` + time.Now().Format("2.01.2006, 3:04pm") + `]
 *     
 *	@license: see "LICENSE" at root directory
 */
{dofiles}

class 
	Commands
{
	function Register ()
	{
{Commands}
	}
}
	`

	var Path string
	if len(os.Args) == 1 {
		fmt.Println("CommandLinker: invalid parameter count. Assuming '../packages/cqfun/'")
		fmt.Println("use: CommandLinker [path]")
		Path = "../packages/cqfun/Server/Commands/"
	} else {
		Path = os.Args[1] + "/"
	}

	fmt.Println("CommandLinker: started...")

	var CommandPath = Path + "Server/Commands"

	var CommandFiles []string = RecursiveRead(CommandPath)
	var Commands []Command
	var CommandName []string

	for _, f := range CommandFiles {
		if f == "Commands.nut" || f == "CQCommand.nut" {
			continue
		}
		r := GetCommandName(fmt.Sprintf("%s/%s", CommandPath, f))
		if r == "ERR" {
			fmt.Println("CommandLinker ERROR: [", f, "]")
			continue
		}
		if stringInSlice(r, CommandName) {
			fmt.Println("CommandLinker ERROR: multiple [", r, "]")
			continue
		}
		Commands = append(Commands, Command{r, f})
		CommandName = append(CommandName, r)
	}

	fmt.Println("CommandLinker: fetched", len(Commands), "commands")

	// prepare our Commands.nut list
	var strCommands string
	var strDofiles string

	for _, command := range Commands {
		strDofiles += strings.Replace(commandDofileProto, "{Path}", "Server/Commands/"+command.path, -1) + "\n"
		strCommands += strings.Replace(commandProto, "{Command}", command.name, 1) + "\n"
	}

	// Write commands into loading file
	os.Remove(CommandPath + "/Commands.nut")
	ioutil.WriteFile(CommandPath+"/Commands.nut", []byte(strings.Replace(strings.Replace(commandFileProto, "{Commands}", strCommands, 1), "{dofiles}", strDofiles, 1)), 0644)
}

func RecursiveRead(dir string) []string {
	var fileList []string

	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if f.Mode().IsDir() {
			recFiles := RecursiveRead(fmt.Sprintf("%s/%s", dir, f.Name()))
			for _, f2 := range recFiles {
				fileList = append(fileList, fmt.Sprintf("%s/%s", f.Name(), f2))
			}
		} else {
			fileList = append(fileList, f.Name())
		}
	}
	return fileList
}

func GetCommandName(file string) string {
	content, _ := ioutil.ReadFile(file)
	contentStr := string(content[:])

	classFind, _ := regexp.Compile("[\\s\\n]{0,}class[\\n\\s]{0,}(.*)extends CQ(.*)")

	if !classFind.MatchString(contentStr) {
		fmt.Println("ModelLinker: Could not find class definition")
		return "ERR"
	}

	classPos := classFind.FindStringIndex(contentStr)

	contentStr = contentStr[classPos[0] : classPos[1]-2]

	classDef, _ := regexp.Compile(`class([\n\s]+?)(.+?)(\s+?)extends(\s+?)CQ(.*?)Command(\s*?)`)

	if !classDef.MatchString(string(content[:])) {
		return "ERR"
	}

	strDef := classDef.FindString(string(content[:]))

	strDef = strings.Replace(strDef, "class", "", 1)
	strDef = strDef[:strings.Index(strDef, "extends CQ")]

	reg, _ := regexp.Compile(`[\s\n]{0,}`)
	strDef = reg.ReplaceAllString(strDef, "")

	return strDef
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
