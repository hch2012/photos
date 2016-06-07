package main

import(
	"log"
	"os"
	"io"
	//"syscall"
	"net/http"
	"io/ioutil"
	"html/template"
	"path"
	"fmt"
)

var UPLOAD_DIR string
var HTML_DIR string
var STATIC_DIR string
var templates =make(map[string]*template.Template)

func init() {
	path:=os.Args[1]
	UPLOAD_DIR=path+"uploads"
	HTML_DIR=path+"html"
	STATIC_DIR=path+"static"
	fmt.Println(HTML_DIR)
	listFile(HTML_DIR,"")
	
	
	for key,_:=range templates{
		fmt.Println(key)
	}
}

func listFile(dirPath string ,keyNamePre string) {
	fmt.Println(dirPath)
	files,err:=ioutil.ReadDir(dirPath)
	check(err)
	for _,file:=range files{
		name:=file.Name()
		if file.IsDir() {
			listFile(dirPath+"/"+name,keyNamePre+name+"/")
		} else {
			if path.Ext(name)!=".html"{
				continue
			}
			path:=dirPath+"/"+name
			t:=template.Must(template.ParseFiles(path))
			templates[keyNamePre+name]=t
		}
	}
}

func renderHtml(w http.ResponseWriter,temp string,params map[string]interface{}) {
	if params==nil{
		params=make(map[string]interface{})
	}
	fmt.Println("-------------------------")
	fmt.Println(templates[temp+".html"])

	err:=templates[temp+".html"].Execute(w,params)
	check(err)
}

func isExists(path string) bool{
	_,err:=os.Stat(path)
	if(err!=nil){
		return os.IsExist(err)
	}
	return true	
}

func main() {
	http.HandleFunc("/upload",uploadPicHandler)
	http.HandleFunc("/view",viewHandler)
	log.Fatal(http.ListenAndServe(":8080",nil))
}

func check(err error) {
	if err!=nil{	
		panic(err)
	}
}


func uploadPicHandler(w http.ResponseWriter ,r *http.Request) {
	if r.Method=="GET" {
		renderHtml(w,"upload",nil)
	} 
	if r.Method=="POST" {
		file,head,err:=r.FormFile("image")
		check(err)
		filename:=head.Filename
		defer file.Close()
		newFile,err:=ioutil.TempFile(UPLOAD_DIR,filename)
		check(err)
		defer newFile.Close()
		_,err =io.Copy(newFile,file)
		check(err)
		http.Redirect(w,r,"/view?id="+removePrefix(newFile.Name(),UPLOAD_DIR),http.StatusFound)
	}
}
func removePrefix(fullName string,prefix string) string {
	return fullName[len(prefix):]
}
func viewHandler(w http.ResponseWriter ,r *http.Request) {

	id:=r.FormValue("id")
	fmt.Println(id)
	if id=="" {
		id="/"
	}
	path:=UPLOAD_DIR+id
	//params:=make(map[string]interface{})
	if !isExists(path) {
		http.NotFound(w,r)
		return
	}
	fmt.Println(path)
	info,err:=os.Stat(path)
	check(err)

	if info.IsDir() {
		showDir(w,r,id)
	} else{
		showFile(w,r,id)
	}
	
}
func showDir(w http.ResponseWriter,r * http.Request ,path string) {
	
	params:=make(map[string]interface{})
	path=appendDirEnd(path)
	params["curDir"]=path
	params["parentDir"]=parentDir(path)
	files,err:=ioutil.ReadDir(UPLOAD_DIR+path)
	check(err)
	params["list"]=files
	renderHtml(w,"list",params)
}

func appendDirEnd(dirPath string) string {
	if dirPath[len(dirPath)-1]=='/' {
		return dirPath
	}
	return dirPath+`/`
}
func parentDir(dirPath string) string {
	fmt.Println(dirPath)
	if dirPath==`/`{
		return ""
	}
	str:=[]rune(dirPath)
	t:=0
	t2:=0
	for i,c:= range str{
		if c=='/'{
			t2=t
			t=i
		}
	}
	return string(str[:t2])
}

func showFile(w http.ResponseWriter,r *http.Request,path string) {
	w.Header().Set("Content-Type","image")
	http.ServeFile(w,r,UPLOAD_DIR+path)
}
