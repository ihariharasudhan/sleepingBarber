package main


import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"strconv"
	"sync"
	"time"
)

type LockHandler struct{
	fileLocks map[string]sync.RWMutex
	lock sync.Mutex
}

// getFileLock looks if the mutex has already been created for the file and returns if any value 
// exists otherwise creates a new lock for the path and returns it..

func (lockHandler *LockHandler) getFileLock(filePath string) (*sync.RWMutex){
	if _, lockAlreadyPresent := lockHandler.fileLocks[filePath]; !lockAlreadyPresent{
		lockHandler.lock.Lock()
		lockHandler.fileLocks[filePath] = sync.RWMutex{}
		lockHandler.lock.Unlock()
	}
	lockHandler.lock.Lock()	
	defer lockHandler.lock.Unlock()
	value := lockHandler.fileLocks[filePath]
	return &value
}

//opens any file in desired mode and handles regular errors.... reduces boilerplate

func openFile(filePath string, mode int) (*os.File, error, bool){
	openStatus := true
	file, fileError := os.OpenFile(filePath, mode, 666) 
	if fileError != nil{
		if(mode == os.O_WRONLY){
			openStatus = false
			return file, fileError, openStatus
		}
		fmt.Printf("Error: File %s ", filePath)

		if os.IsPermission(fileError){
			fmt.Printf("has permission denied.")

		} else if os.IsNotExist(fileError){
			fmt.Printf("does not exist.")
		}
		fmt.Printf("\n")	
		openStatus = false
	}
	return file, fileError, openStatus
}

// copyFile checks for source file and destination file and creates lock for them. if the destination
// file does not exist, it creates a new file for copying. Source file is required though.

func copyFile(sourcePath, destinationPath string){

	sourceFileLock := lockHandler.getFileLock(sourcePath)
	lockHandler.lock.Lock()
	sourceFileLock.RLock()
	lockHandler.lock.Unlock()

	sourceFile, _, sourceFileOpenSuccessful := openFile(sourcePath, os.O_RDONLY)
	if(!sourceFileOpenSuccessful){
		fmt.Println("Copy process failed:", sourcePath, "->", destinationPath)
		return
	}
	defer sourceFile.Close()

	destinationFileLock := lockHandler.getFileLock(destinationPath)

	lockHandler.lock.Lock()
	destinationFileLock.Lock()
	lockHandler.lock.Unlock()



	destinationFile, _, destinationFileOpenSuccessful := openFile(destinationPath, os.O_WRONLY)	
	if(!destinationFileOpenSuccessful){
		_, createError := os.Create(destinationPath)
		if createError != nil{
			fmt.Println("Copy process failed:", sourcePath, "->", destinationPath)
			return
		}
		openFile, _, _ := openFile(destinationPath, os.O_WRONLY)
		destinationFile = openFile
	}
	defer destinationFile.Close()

//	Copying and commiting code	
	_, copyError := io.Copy(destinationFile, sourceFile)	
	if(copyError != nil){
		fmt.Println("Error: Copy process failed")
		fmt.Println(copyError)
		return
	}
	
	lockHandler.lock.Lock()
	sourceFileLock.RUnlock()	
	lockHandler.lock.Lock()

	commitError := destinationFile.Sync()
	if commitError != nil {
		fmt.Println("Error: Unable to save changes to the file")
		fmt.Println(commitError)
	}
	lockHandler.lock.Lock()
	destinationFileLock.Unlock()
	lockHandler.lock.Unlock()
}

func moveFile(sourcePath, destinationPath string){

	sourceFileLock := lockHandler.getFileLock(sourcePath)
	sourceFileLock.Lock()
	defer sourceFileLock.Unlock()	
	sourceFile, _, sourceFileOpenSuccessful := openFile(sourcePath, os.O_WRONLY)
	if(!sourceFileOpenSuccessful){
		fmt.Println("Move process failed:", sourcePath, "->", destinationPath)
		return
	}
	defer sourceFile.Close()

	destinationFileLock := lockHandler.getFileLock(destinationPath)
	destinationFileLock.Lock()
	defer destinationFileLock.Unlock()

//	Move operation
	moveError := os.Rename(sourcePath, destinationPath)
	if moveError != nil{
		fmt.Println("Move process failed:", sourcePath, "->", destinationPath)
		return
	}
}


// global lockhandler for all file related locks
var lockHandler = LockHandler{fileLocks:make(map[string]sync.RWMutex)}


func check(e error) {

	//handles errors from the file read operation
	if e != nil {
		panic(e)
	}
}

func LonerBarber(fileList chan string, data chan<- string, group *sync.WaitGroup, id int) {

	//fileList is the name of the channel that has the file names
	//data is a channel that will be loaded with the contents in the file
	//group is the waitgroup
	//id denotes the process ID
	defer group.Done() //wait group defer
	for fName := range fileList {
		start := time.Now()                      //storing the time at which the go routine starts
		fmt.Println("Time is ", start)           //printing the time
		fmt.Println("This is loner barber ", id) //Printing the process ID
		fmt.Println("Data is from: ", fName)     //printing the name of the file
		dat, err := ioutil.ReadFile(fName)       //storing the contents of the file
		check(err)                               //checking for error
		data <- string(dat)                      //loading the contents into the channel
		time.Sleep(time.Second)                  //to be replaced with your copy code***************************
		go copyFile(fName, fName+"-copy.txt")
	}
}

func main() {

	var group sync.WaitGroup //wait group variable
	var files, nProcess int
	data := make(chan string, 10)        //channel that receives the contents of the file
	fileList := make(chan string, files) //channel that holds the list of files to be read

	fmt.Println("Enter the number of barbers you want: ")
	fmt.Scanln(&nProcess)

	start := time.Now() //storing the start time
	for i := 1; i <= nProcess; i++ {
		group.Add(1)                              //adding the process to wait group
		go LonerBarber(fileList, data, &group, i) //creation of worker pool

	}

	fmt.Println("Enter the number of files: ")
	fmt.Scanln(&files)
	path := "./files/"
	fileN := "text"
	var name string
	for i := 0; i < files; i++ { //generating the filenames
		name = path + fileN + strconv.Itoa(i) + ".txt"
		fileList <- name
	}
	close(fileList)
	for i := 0; i < files; i++ {
		dat := <-data                 //storing the data from the file
		fmt.Println("In file: ", dat) //printing the data from the file
	}
	end := time.Now()                          //storing the end time
	elapsed := end.Sub(start)                  //calulating time elapsed
	fmt.Println("Total time taken: ", elapsed) //printing the elapsed time

	group.Wait() //blocking until all wait groups have completed the task
}
